package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beetrack/backend/internal/llm"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/model"
	"gorm.io/datatypes"
)

const (
	stuckProcessingTimeout = 5 * time.Minute

	poorQualityAvgLogprob      = -1.0
	poorQualityNoSpeechProb    = 0.6
	poorQualitySegmentFraction = 0.5

	defaultHiveResolutionModel = "anthropic/claude-haiku-4.5"
	hiveResolutionMaxTokens    = 256
	resolveHiveToolName        = "resolve_hive"

	resolveHiveOutcomeMatched       = "matched"
	resolveHiveOutcomeNotIdentified = "not_identified"
	resolveHiveOutcomeMultiple      = "multiple"

	errHiveNotIdentified      = "HIVE_NOT_IDENTIFIED"
	errMultipleHivesMentioned = "MULTIPLE_HIVES_MENTIONED"

	hiveResolutionSystemPrompt = "You resolve which hive a beekeeper's voice note is about. You are given a transcript and the list of hives (id, name) in the apiary the beekeeper is currently recording in. Call resolve_hive exactly once: outcome=matched with that hive's id if the transcript clearly names exactly one hive from the list (allow for minor mishearing/transcription noise); outcome=multiple if it names more than one distinct hive from the list; outcome=not_identified if no hive from the list is clearly named or the match is ambiguous (include spoken_hive_name if you can tell what name was said)."

	actionProposalMaxTokens = 1024
	hiveContextDays         = 90

	actionProposalSystemPrompt = "You propose logging actions from a beekeeper's voice note about one specific hive — the hive is already fixed, don't try to identify it. You are given the transcript and that hive's current context (type, status flags, diseases, and its most recent inspection's frame counts, if any). Call any of the available tools the transcript actually describes — zero, one, or several, one call per distinct topic (e.g. an inspection and a feeding are two separate calls). Only fill in a field if the beekeeper actually said something it maps to; leave every other field unset rather than guessing. Frame-count deltas (the frames_added_* fields) should be computed relative to the hive's last known frame counts given in its context, if there are any. If the transcript doesn't describe any loggable action at all, don't call any tool."
)

type VoiceRecordingRepository interface {
	ClaimNext(ctx context.Context) (*model.VoiceRecording, error)
	MarkCompleted(ctx context.Context, id int64, transcript, language string) error
	MarkFailed(ctx context.Context, id int64, errMsg string) error
	MarkRetry(ctx context.Context, id int64, errMsg string, nextAttemptAt time.Time) error
	SweepStuckProcessing(ctx context.Context, olderThan time.Duration) (int64, error)
	CreateAction(ctx context.Context, action *model.VoiceAction) error
	CreateLLMCall(ctx context.Context, call *model.VoiceLLMCall) error
}

type Transcriber interface {
	Transcribe(ctx context.Context, audio io.Reader, filename string) (*llm.TranscriptionResult, error)
}

type HiveLister interface {
	ListHives(ctx context.Context, userID int64, apiaryID *int64) ([]mcp.HiveSummary, error)
	GetHiveSummary(ctx context.Context, userID, hiveID int64, days *int) (*mcp.HiveHistory, error)
}

type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error)
}

type AudioStore interface {
	Open(path string) (io.ReadCloser, error)
	Delete(path string) error
}

type FileAudioStore struct {
	basePath string
}

func NewFileAudioStore(basePath string) *FileAudioStore {
	return &FileAudioStore{basePath: basePath}
}

func (s *FileAudioStore) Open(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.basePath, path))
}

func (s *FileAudioStore) Delete(path string) error {
	return os.Remove(filepath.Join(s.basePath, path))
}

type VoiceWorker struct {
	recordings  VoiceRecordingRepository
	transcriber Transcriber
	audio       AudioStore
	hives       HiveLister
	llmClient   ChatCompleter
	model       string
}

// model selects the model used for hive-name resolution and action proposal; pass "" to use
// defaultHiveResolutionModel.
func NewVoiceWorker(recordings VoiceRecordingRepository, transcriber Transcriber, audio AudioStore, hives HiveLister, llmClient ChatCompleter, model string) *VoiceWorker {
	if model == "" {
		model = defaultHiveResolutionModel
	}
	return &VoiceWorker{recordings: recordings, transcriber: transcriber, audio: audio, hives: hives, llmClient: llmClient, model: model}
}

func (w *VoiceWorker) Run(ctx context.Context, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if n, err := w.recordings.SweepStuckProcessing(ctx, stuckProcessingTimeout); err != nil {
				slog.Error("sweep stuck processing recordings failed", "component", "voice_worker", "error", err)
			} else if n > 0 {
				slog.Warn("reset stuck processing recordings", "component", "voice_worker", "count", n)
			}

			for {
				processed, err := w.ProcessNext(ctx)
				if err != nil {
					slog.Error("process recording failed", "component", "voice_worker", "error", err)
				}
				if !processed {
					break
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *VoiceWorker) ProcessNext(ctx context.Context) (processed bool, err error) {
	rec, err := w.recordings.ClaimNext(ctx)
	if err != nil {
		return false, fmt.Errorf("claim next recording: %w", err)
	}
	if rec == nil {
		return false, nil
	}

	if rec.AudioPath == nil {
		return true, w.recordings.MarkFailed(ctx, rec.ID, "recording has no audio file")
	}

	file, err := w.audio.Open(*rec.AudioPath)
	if err != nil {
		return true, w.recordings.MarkFailed(ctx, rec.ID, fmt.Sprintf("open audio file: %v", err))
	}
	defer file.Close()

	result, err := w.transcriber.Transcribe(ctx, file, *rec.AudioPath)
	if err != nil {
		return true, w.handleTranscribeError(ctx, rec, err)
	}

	if isPoorAudioQuality(result) {
		w.deleteAudio(*rec.AudioPath)
		return true, w.recordings.MarkFailed(ctx, rec.ID, "POOR_AUDIO_QUALITY")
	}
	w.deleteAudio(*rec.AudioPath)

	resolved, err := w.resolveHive(ctx, rec, result.Text)
	if err != nil {
		return true, w.recordings.MarkFailed(ctx, rec.ID, err.Error())
	}
	if resolved.Outcome != resolveHiveOutcomeMatched {
		if err := w.recordings.CreateAction(ctx, hiveResolutionErrorAction(rec.ID, resolved)); err != nil {
			return true, fmt.Errorf("create hive resolution error action: %w", err)
		}
		return true, w.recordings.MarkCompleted(ctx, rec.ID, result.Text, result.Language)
	}

	if err := w.proposeActions(ctx, rec, *resolved.HiveID, result.Text); err != nil {
		return true, w.recordings.MarkFailed(ctx, rec.ID, err.Error())
	}

	return true, w.recordings.MarkCompleted(ctx, rec.ID, result.Text, result.Language)
}

func (w *VoiceWorker) handleTranscribeError(ctx context.Context, rec *model.VoiceRecording, cause error) error {
	attempt := rec.RetryCount + 1
	if !isTransientWhisperError(cause) || attempt >= model.MaxWhisperRetries {
		w.deleteAudio(*rec.AudioPath)
		return w.recordings.MarkFailed(ctx, rec.ID, cause.Error())
	}
	return w.recordings.MarkRetry(ctx, rec.ID, cause.Error(), time.Now().Add(whisperBackoff(attempt)))
}

func (w *VoiceWorker) deleteAudio(path string) {
	if err := w.audio.Delete(path); err != nil {
		slog.Warn("failed to delete audio file", "component", "voice_worker", "path", path, "error", err)
	}
}

func (w *VoiceWorker) logLLMCall(ctx context.Context, recordingID int64, phase string, request any, response any, callErr error) {
	reqJSON, err := json.Marshal(request)
	if err != nil {
		slog.Warn("marshal voice llm call request", "component", "voice_worker", "phase", phase, "error", err)
		return
	}

	var respJSON []byte
	if callErr != nil {
		respJSON, _ = json.Marshal(map[string]string{"error": callErr.Error()})
	} else {
		respJSON, err = json.Marshal(response)
		if err != nil {
			slog.Warn("marshal voice llm call response", "component", "voice_worker", "phase", phase, "error", err)
			return
		}
	}

	call := &model.VoiceLLMCall{
		VoiceRecordingID: recordingID,
		Phase:            phase,
		Request:          datatypes.JSON(reqJSON),
		Response:         datatypes.JSON(respJSON),
		IsError:          callErr != nil,
	}
	if err := w.recordings.CreateLLMCall(ctx, call); err != nil {
		slog.Warn("persist voice llm call", "component", "voice_worker", "phase", phase, "error", err)
	}
}

// whisperBackoff returns the fixed 5s / 30s / 2m schedule for attempt 1, 2, 3+.
func whisperBackoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 5 * time.Second
	case 2:
		return 30 * time.Second
	default:
		return 2 * time.Minute
	}
}

func isTransientWhisperError(err error) bool {
	var apiErr *llm.WhisperAPIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

type hiveOption struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type resolveHiveInput struct {
	Outcome        string  `json:"outcome"`
	HiveID         *int64  `json:"hive_id,omitempty"`
	SpokenHiveName *string `json:"spoken_hive_name,omitempty"`
}

func resolveHiveTool() llm.Tool {
	return functionTool(resolveHiveToolName, "Report which single hive (if any) the transcript is about, given the apiary's hive list.", map[string]any{
		"outcome": map[string]any{
			"type": "string",
			"enum": []string{resolveHiveOutcomeMatched, resolveHiveOutcomeNotIdentified, resolveHiveOutcomeMultiple},
		},
		"hive_id": map[string]any{
			"type":        "integer",
			"description": "Required when outcome is matched: the id of the one hive from the list this transcript is about.",
		},
		"spoken_hive_name": map[string]any{
			"type":        "string",
			"description": "Best-effort transcription of the hive name mentioned, when outcome is not_identified.",
		},
	}, []string{"outcome"})
}

func functionTool(name, description string, properties map[string]any, required []string) llm.Tool {
	schema, _ := json.Marshal(map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	})
	return llm.Tool{Type: "function", Function: llm.FunctionDef{Name: name, Description: description, Parameters: schema}}
}

type resolveHiveRequest struct {
	Transcript string       `json:"transcript"`
	Hives      []hiveOption `json:"hives"`
}

func (w *VoiceWorker) resolveHive(ctx context.Context, rec *model.VoiceRecording, transcript string) (*resolveHiveInput, error) {
	apiaryID := rec.ApiaryID
	hives, err := w.hives.ListHives(ctx, rec.UserID, &apiaryID)
	if err != nil {
		return nil, fmt.Errorf("list hives: %w", err)
	}
	if len(hives) == 0 {
		return &resolveHiveInput{Outcome: resolveHiveOutcomeNotIdentified}, nil
	}

	options := make([]hiveOption, len(hives))
	for i, h := range hives {
		options[i] = hiveOption{ID: h.ID, Name: h.Name}
	}
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("marshal hive options: %w", err)
	}

	resp, err := w.llmClient.CreateChatCompletion(ctx, llm.ChatCompletionRequest{
		Model:     w.model,
		MaxTokens: hiveResolutionMaxTokens,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: hiveResolutionSystemPrompt},
			{Role: "user", Content: fmt.Sprintf("Transcript: %q\n\nHives in this apiary:\n%s", transcript, optionsJSON)},
		},
		Tools:      []llm.Tool{resolveHiveTool()},
		ToolChoice: llm.ForceTool(resolveHiveToolName),
	})
	w.logLLMCall(ctx, rec.ID, model.VoiceLLMCallPhaseResolveHive, resolveHiveRequest{Transcript: transcript, Hives: options}, resp, err)
	if err != nil {
		return nil, fmt.Errorf("resolve hive: %w", err)
	}

	for _, call := range chatToolCalls(resp) {
		if call.Function.Name != resolveHiveToolName {
			continue
		}
		var in resolveHiveInput
		if err := json.Unmarshal([]byte(call.Function.Arguments), &in); err != nil {
			return nil, fmt.Errorf("decode resolve_hive input: %w", err)
		}
		if in.Outcome == resolveHiveOutcomeMatched && (in.HiveID == nil || !containsHiveID(options, *in.HiveID)) {
			return &resolveHiveInput{Outcome: resolveHiveOutcomeNotIdentified}, nil
		}
		return &in, nil
	}
	return nil, errors.New("model did not call resolve_hive")
}

func chatToolCalls(resp *llm.ChatCompletionResponse) []llm.ToolCall {
	if len(resp.Choices) == 0 {
		return nil
	}
	return resp.Choices[0].Message.ToolCalls
}

func containsHiveID(options []hiveOption, id int64) bool {
	for _, o := range options {
		if o.ID == id {
			return true
		}
	}
	return false
}

func hiveResolutionErrorAction(recordingID int64, resolved *resolveHiveInput) *model.VoiceAction {
	errCode := errHiveNotIdentified
	if resolved.Outcome == resolveHiveOutcomeMultiple {
		errCode = errMultipleHivesMentioned
	}
	resultType := model.VoiceActionResultTypeError
	return &model.VoiceAction{
		VoiceRecordingID: recordingID,
		Sequence:         1,
		SpokenHiveName:   resolved.SpokenHiveName,
		Status:           model.VoiceActionStatusError,
		ResultType:       &resultType,
		ErrorMessage:     &errCode,
	}
}

var proposableTools = map[string]bool{
	model.VoiceActionToolCreateInspection: true,
	model.VoiceActionToolCreateTreatment:  true,
	model.VoiceActionToolCreateHarvest:    true,
	model.VoiceActionToolCreateFeeding:    true,
}

type lastInspectionContext struct {
	InspectedAt  time.Time `json:"inspected_at"`
	FramesBrood  *int      `json:"frames_brood"`
	FramesFeed   *int      `json:"frames_feed"`
	FramesPollen *int      `json:"frames_pollen"`
}

type hiveActionContext struct {
	Type            string                 `json:"type"`
	Queenless       bool                   `json:"queenless"`
	NeedsFood       bool                   `json:"needs_food"`
	ReadyForHarvest bool                   `json:"ready_for_harvest"`
	Diseases        []string               `json:"diseases"`
	LastInspection  *lastInspectionContext `json:"last_inspection,omitempty"`
}

func buildHiveActionContext(h *mcp.HiveHistory) hiveActionContext {
	c := hiveActionContext{
		Type:            h.Hive.Type,
		Queenless:       h.Hive.Queenless,
		NeedsFood:       h.Hive.NeedsFood,
		ReadyForHarvest: h.Hive.ReadyForHarvest,
		Diseases:        h.Hive.Diseases,
	}
	var latest *mcp.InspectionSummary
	for i := range h.Inspections {
		insp := &h.Inspections[i]
		if latest == nil || insp.InspectedAt.After(latest.InspectedAt) {
			latest = insp
		}
	}
	if latest != nil {
		c.LastInspection = &lastInspectionContext{
			InspectedAt:  latest.InspectedAt,
			FramesBrood:  latest.FramesBrood,
			FramesFeed:   latest.FramesFeed,
			FramesPollen: latest.FramesPollen,
		}
	}
	return c
}

func createInspectionTool() llm.Tool {
	return functionTool(model.VoiceActionToolCreateInspection, "Propose logging a hive inspection.", map[string]any{
		"queen_status":            map[string]any{"type": "string", "enum": model.ValidQueenStatuses},
		"brood_pattern":           map[string]any{"type": "string", "enum": model.ValidBroodPatterns},
		"aggressiveness":          map[string]any{"type": "string", "enum": model.ValidAggressiveness},
		"frames_brood":            map[string]any{"type": "integer", "description": "Absolute frame count with brood, 0-99."},
		"frames_feed":             map[string]any{"type": "integer", "description": "Absolute frame count with feed, 0-99."},
		"frames_pollen":           map[string]any{"type": "integer", "description": "Absolute frame count with pollen, 0-99."},
		"queen_cells_count":       map[string]any{"type": "integer", "description": "Number of queen cells seen, 0-99."},
		"frames_added_foundation": map[string]any{"type": "integer", "description": "Signed delta of foundation frames added since the hive's last known state."},
		"frames_added_drawn":      map[string]any{"type": "integer", "description": "Signed delta of drawn frames added."},
		"frames_added_brood":      map[string]any{"type": "integer", "description": "Signed delta of brood frames added."},
		"frames_added_feed":       map[string]any{"type": "integer", "description": "Signed delta of feed frames added."},
		"queen_added":             map[string]any{"type": "boolean", "description": "True if a new queen was introduced during this inspection."},
		"diseases": map[string]any{
			"type":        "array",
			"description": "Diseases observed during this inspection, if any.",
			"items": map[string]any{
				"type": "string",
				"enum": model.ValidDiseases,
			},
		},
		"notes": map[string]any{"type": "string", "description": "Freeform notes, close to what the beekeeper said."},
	}, nil)
}

func createTreatmentTool() llm.Tool {
	return functionTool(model.VoiceActionToolCreateTreatment, "Propose logging a treatment applied to the hive.", map[string]any{
		"medicine_name": map[string]any{"type": "string", "description": "Name of the medicine/treatment used."},
		"dose":          map[string]any{"type": "string", "description": `Dose given, e.g. "1 strip". Defaults to "1" if not mentioned.`},
		"notes":         map[string]any{"type": "string"},
	}, []string{"medicine_name"})
}

func createHarvestTool() llm.Tool {
	return functionTool(model.VoiceActionToolCreateHarvest, "Propose logging honey harvested from the hive.", map[string]any{
		"frames":      map[string]any{"type": "integer", "description": "Whole frames harvested, 0-99."},
		"half_frames": map[string]any{"type": "integer", "description": "Half frames harvested, 0-99."},
		"kilograms":   map[string]any{"type": "number", "description": "Kilograms of honey harvested, greater than 0."},
		"notes":       map[string]any{"type": "string"},
	}, []string{"kilograms"})
}

func createFeedingTool() llm.Tool {
	return functionTool(model.VoiceActionToolCreateFeeding, "Propose logging feed given to the hive.", map[string]any{
		"feed_type": map[string]any{"type": "string", "description": "Type of feed given, e.g. sugar syrup, fondant."},
		"amount":    map[string]any{"type": "string", "description": `Amount fed, e.g. "1L", "500g".`},
		"notes":     map[string]any{"type": "string"},
	}, []string{"feed_type"})
}

func actionProposalTools() []llm.Tool {
	return []llm.Tool{createInspectionTool(), createTreatmentTool(), createHarvestTool(), createFeedingTool()}
}

type proposeActionsRequest struct {
	Transcript  string            `json:"transcript"`
	HiveContext hiveActionContext `json:"hive_context"`
}

func (w *VoiceWorker) proposeActions(ctx context.Context, rec *model.VoiceRecording, hiveID int64, transcript string) error {
	summary, err := w.hives.GetHiveSummary(ctx, rec.UserID, hiveID, hiveContextDaysPtr())
	if err != nil {
		return fmt.Errorf("get hive summary: %w", err)
	}
	hiveContext := buildHiveActionContext(summary)
	contextJSON, err := json.Marshal(hiveContext)
	if err != nil {
		return fmt.Errorf("marshal hive context: %w", err)
	}

	resp, err := w.llmClient.CreateChatCompletion(ctx, llm.ChatCompletionRequest{
		Model:     w.model,
		MaxTokens: actionProposalMaxTokens,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: actionProposalSystemPrompt},
			{Role: "user", Content: fmt.Sprintf("Transcript: %q\n\nHive context:\n%s", transcript, contextJSON)},
		},
		Tools: actionProposalTools(),
	})
	w.logLLMCall(ctx, rec.ID, model.VoiceLLMCallPhaseProposeActions, proposeActionsRequest{Transcript: transcript, HiveContext: hiveContext}, resp, err)
	if err != nil {
		return fmt.Errorf("propose actions: %w", err)
	}

	sequence := 0
	for _, call := range chatToolCalls(resp) {
		if !proposableTools[call.Function.Name] {
			continue
		}
		sequence++
		toolName := call.Function.Name
		action := &model.VoiceAction{
			VoiceRecordingID: rec.ID,
			Sequence:         sequence,
			HiveID:           &hiveID,
			ToolName:         &toolName,
			ToolArguments:    datatypes.JSON(call.Function.Arguments),
			Status:           model.VoiceActionStatusProposed,
		}
		if err := w.recordings.CreateAction(ctx, action); err != nil {
			return fmt.Errorf("create proposed action: %w", err)
		}
	}
	return nil
}

func hiveContextDaysPtr() *int {
	days := hiveContextDays
	return &days
}

func isPoorAudioQuality(result *llm.TranscriptionResult) bool {
	if strings.TrimSpace(result.Text) == "" || len(result.Segments) == 0 {
		return true
	}
	var poor int
	for _, seg := range result.Segments {
		if seg.AvgLogprob < poorQualityAvgLogprob || seg.NoSpeechProb > poorQualityNoSpeechProb {
			poor++
		}
	}
	return float64(poor)/float64(len(result.Segments)) > poorQualitySegmentFraction
}
