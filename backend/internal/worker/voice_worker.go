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

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/beetrack/backend/internal/llm"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/model"
)

const (
	stuckProcessingTimeout = 5 * time.Minute

	poorQualityAvgLogprob      = -1.0
	poorQualityNoSpeechProb    = 0.6
	poorQualitySegmentFraction = 0.5

	defaultHiveResolutionModel = anthropic.ModelClaudeHaiku4_5
	hiveResolutionMaxTokens    = 256
	resolveHiveToolName        = "resolve_hive"

	resolveHiveOutcomeMatched       = "matched"
	resolveHiveOutcomeNotIdentified = "not_identified"
	resolveHiveOutcomeMultiple      = "multiple"

	errHiveNotIdentified      = "HIVE_NOT_IDENTIFIED"
	errMultipleHivesMentioned = "MULTIPLE_HIVES_MENTIONED"

	hiveResolutionSystemPrompt = "You resolve which hive a beekeeper's voice note is about. You are given a transcript and the list of hives (id, name) in the apiary the beekeeper is currently recording in. Call resolve_hive exactly once: outcome=matched with that hive's id if the transcript clearly names exactly one hive from the list (allow for minor mishearing/transcription noise); outcome=multiple if it names more than one distinct hive from the list; outcome=not_identified if no hive from the list is clearly named or the match is ambiguous (include spoken_hive_name if you can tell what name was said)."
)

type VoiceRecordingRepository interface {
	ClaimNext(ctx context.Context) (*model.VoiceRecording, error)
	MarkCompleted(ctx context.Context, id int64, transcript, language string) error
	MarkFailed(ctx context.Context, id int64, errMsg string) error
	MarkRetry(ctx context.Context, id int64, errMsg string, nextAttemptAt time.Time) error
	SweepStuckProcessing(ctx context.Context, olderThan time.Duration) (int64, error)
	CreateAction(ctx context.Context, action *model.VoiceAction) error
}

type Transcriber interface {
	Transcribe(ctx context.Context, audio io.Reader, filename string) (*llm.TranscriptionResult, error)
}

type HiveLister interface {
	ListHives(ctx context.Context, userID int64, apiaryID *int64) ([]mcp.HiveSummary, error)
}

type HiveNameResolver interface {
	New(ctx context.Context, params anthropic.MessageNewParams, opts ...option.RequestOption) (*anthropic.Message, error)
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
	claude      HiveNameResolver
	model       anthropic.Model
}

// model selects the Claude model used for hive-name resolution; pass "" to use defaultHiveResolutionModel.
func NewVoiceWorker(recordings VoiceRecordingRepository, transcriber Transcriber, audio AudioStore, hives HiveLister, claude HiveNameResolver, model string) *VoiceWorker {
	m := anthropic.Model(model)
	if m == "" {
		m = defaultHiveResolutionModel
	}
	return &VoiceWorker{recordings: recordings, transcriber: transcriber, audio: audio, hives: hives, claude: claude, model: m}
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

func resolveHiveTool() anthropic.ToolUnionParam {
	return anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{
		Name:        resolveHiveToolName,
		Description: param.NewOpt("Report which single hive (if any) the transcript is about, given the apiary's hive list."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
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
			},
			Required: []string{"outcome"},
		},
	}}
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

	msg, err := w.claude.New(ctx, anthropic.MessageNewParams{
		Model:      w.model,
		MaxTokens:  hiveResolutionMaxTokens,
		System:     []anthropic.TextBlockParam{{Text: hiveResolutionSystemPrompt}},
		Messages:   []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf("Transcript: %q\n\nHives in this apiary:\n%s", transcript, optionsJSON)))},
		Tools:      []anthropic.ToolUnionParam{resolveHiveTool()},
		ToolChoice: anthropic.ToolChoiceParamOfTool(resolveHiveToolName),
	})
	if err != nil {
		return nil, fmt.Errorf("resolve hive: %w", err)
	}

	for _, block := range msg.Content {
		if block.Type != "tool_use" || block.Name != resolveHiveToolName {
			continue
		}
		var in resolveHiveInput
		if err := json.Unmarshal(block.Input, &in); err != nil {
			return nil, fmt.Errorf("decode resolve_hive input: %w", err)
		}
		if in.Outcome == resolveHiveOutcomeMatched && (in.HiveID == nil || !containsHiveID(options, *in.HiveID)) {
			return &resolveHiveInput{Outcome: resolveHiveOutcomeNotIdentified}, nil
		}
		return &in, nil
	}
	return nil, errors.New("claude did not call resolve_hive")
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
