package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/llm"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/model"
)

type mockVoiceRepo struct {
	next     *model.VoiceRecording
	claimErr error

	completedID       int64
	completedText     string
	completedLanguage string
	completedCalled   bool

	failedID     int64
	failedErr    string
	failedCalled bool

	retryID            int64
	retryErr           string
	retryNextAttemptAt time.Time
	retryCalled        bool

	sweptCalled bool
	sweptCount  int64

	createdAction      *model.VoiceAction
	createActionCalled bool
	createdActions     []*model.VoiceAction

	createdLLMCalls  []*model.VoiceLLMCall
	createLLMCallErr error
}

func (m *mockVoiceRepo) ClaimNext(ctx context.Context) (*model.VoiceRecording, error) {
	return m.next, m.claimErr
}

func (m *mockVoiceRepo) MarkCompleted(ctx context.Context, id int64, transcript, language string) error {
	m.completedCalled = true
	m.completedID = id
	m.completedText = transcript
	m.completedLanguage = language
	return nil
}

func (m *mockVoiceRepo) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	m.failedCalled = true
	m.failedID = id
	m.failedErr = errMsg
	return nil
}

func (m *mockVoiceRepo) MarkRetry(ctx context.Context, id int64, errMsg string, nextAttemptAt time.Time) error {
	m.retryCalled = true
	m.retryID = id
	m.retryErr = errMsg
	m.retryNextAttemptAt = nextAttemptAt
	return nil
}

func (m *mockVoiceRepo) SweepStuckProcessing(ctx context.Context, olderThan time.Duration) (int64, error) {
	m.sweptCalled = true
	return m.sweptCount, nil
}

func (m *mockVoiceRepo) CreateAction(ctx context.Context, action *model.VoiceAction) error {
	m.createActionCalled = true
	m.createdAction = action
	m.createdActions = append(m.createdActions, action)
	return nil
}

func (m *mockVoiceRepo) CreateLLMCall(ctx context.Context, call *model.VoiceLLMCall) error {
	m.createdLLMCalls = append(m.createdLLMCalls, call)
	return m.createLLMCallErr
}

type mockTranscriber struct {
	result *llm.TranscriptionResult
	err    error

	blockUntilCtxDone bool
	capturedCtx       context.Context
}

func (m *mockTranscriber) Transcribe(ctx context.Context, audio io.Reader, filename string) (*llm.TranscriptionResult, error) {
	m.capturedCtx = ctx
	if m.blockUntilCtxDone {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return m.result, m.err
}

type mockAudioStore struct {
	openErr   error
	deleted   []string
	deleteErr error
}

func (m *mockAudioStore) Open(path string) (io.ReadCloser, error) {
	if m.openErr != nil {
		return nil, m.openErr
	}
	return io.NopCloser(bytes.NewReader(nil)), nil
}

func (m *mockAudioStore) Delete(path string) error {
	m.deleted = append(m.deleted, path)
	return m.deleteErr
}

type mockHiveLister struct {
	hives      []mcp.HiveSummary
	err        error
	callCount  int
	summary    *mcp.HiveHistory
	summaryErr error
}

func (m *mockHiveLister) ListHives(ctx context.Context, userID int64, apiaryID *int64) ([]mcp.HiveSummary, error) {
	m.callCount++
	return m.hives, m.err
}

func (m *mockHiveLister) GetHiveSummary(ctx context.Context, userID, hiveID int64, days *int) (*mcp.HiveHistory, error) {
	if m.summary == nil && m.summaryErr == nil {
		return &mcp.HiveHistory{}, nil
	}
	return m.summary, m.summaryErr
}

type mockSuggestionProvider struct {
	medicineNames    []string
	medicineNamesErr error
	doses            []string
	dosesErr         error
	feedTypes        []string
	feedTypesErr     error
	amounts          []string
	amountsErr       error
}

func (m *mockSuggestionProvider) MedicineSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return m.medicineNames, m.medicineNamesErr
}

func (m *mockSuggestionProvider) DoseSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return m.doses, m.dosesErr
}

func (m *mockSuggestionProvider) FeedTypeSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return m.feedTypes, m.feedTypesErr
}

func (m *mockSuggestionProvider) AmountSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return m.amounts, m.amountsErr
}

type mockHiveResolver struct {
	resolveMessage *llm.ChatCompletionResponse
	resolveErr     error
	proposeMessage *llm.ChatCompletionResponse
	proposeErr     error
	callCount      int
	requests       []llm.ChatCompletionRequest

	resolveCtx context.Context
	proposeCtx context.Context
}

func (m *mockHiveResolver) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	m.callCount++
	m.requests = append(m.requests, req)
	if req.ToolChoice != nil {
		m.resolveCtx = ctx
		return m.resolveMessage, m.resolveErr
	}
	m.proposeCtx = ctx
	return m.proposeMessage, m.proposeErr
}

func resolveHiveMessage(t *testing.T, input resolveHiveInput) *llm.ChatCompletionResponse {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal resolve_hive input: %v", err)
	}
	return toolCallResponse([2]string{resolveHiveToolName, string(data)})
}

func toolUseMessage(t *testing.T, calls ...[2]string) *llm.ChatCompletionResponse {
	t.Helper()
	return toolCallResponse(calls...)
}

func toolCallResponse(calls ...[2]string) *llm.ChatCompletionResponse {
	toolCalls := make([]llm.ToolCall, len(calls))
	for i, c := range calls {
		toolCalls[i] = llm.ToolCall{ID: c[0], Type: "function", Function: llm.FunctionCall{Name: c[0], Arguments: c[1]}}
	}
	return &llm.ChatCompletionResponse{Choices: []llm.Choice{{Message: llm.ChatMessage{Role: "assistant", ToolCalls: toolCalls}}}}
}

func clearAudioPath(rec *model.VoiceRecording, path string) *model.VoiceRecording {
	rec.AudioPath = &path
	return rec
}

func newTranscribedWorker(repo *mockVoiceRepo, transcriber *mockTranscriber, audio *mockAudioStore, hives *mockHiveLister, resolver *mockHiveResolver) *VoiceWorker {
	return NewVoiceWorker(repo, transcriber, audio, hives, nil, resolver, "")
}

func newTranscribedWorkerWithSuggestions(repo *mockVoiceRepo, transcriber *mockTranscriber, audio *mockAudioStore, hives *mockHiveLister, suggestions *mockSuggestionProvider, resolver *mockHiveResolver) *VoiceWorker {
	return NewVoiceWorker(repo, transcriber, audio, hives, suggestions, resolver, "")
}

func TestVoiceWorker_ProcessNext_NoRecordingAvailable(t *testing.T) {
	w := newTranscribedWorker(&mockVoiceRepo{}, &mockTranscriber{}, &mockAudioStore{}, &mockHiveLister{}, &mockHiveResolver{})

	processed, err := w.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if processed {
		t.Error("expected processed=false when no recording is claimable")
	}
}

func TestVoiceWorker_ProcessNext_HappyPath(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	audio := &mockAudioStore{}
	hives := &mockHiveLister{
		hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}},
		summary: &mcp.HiveHistory{
			Hive: mcp.HiveSummary{Type: "langstroth", QueenNeedsReplacement: true, Diseases: []string{"varroa"}},
		},
	}
	hiveID := int64(42)
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, audio, hives, resolver)

	processed, err := w.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if !repo.completedCalled {
		t.Fatal("expected recording to be marked completed")
	}
	if repo.completedText != "hive three looked good" || repo.completedLanguage != "en" {
		t.Errorf("unexpected completed transcript/language: %q/%q", repo.completedText, repo.completedLanguage)
	}
	if repo.createActionCalled {
		t.Error("expected no voice_actions row when Claude proposes nothing")
	}
	if len(audio.deleted) != 1 || audio.deleted[0] != "audio.wav" {
		t.Errorf("expected audio file to be deleted, got %+v", audio.deleted)
	}
	if len(repo.createdLLMCalls) != 2 {
		t.Fatalf("expected 2 voice_llm_calls rows (resolve_hive + propose_actions), got %d", len(repo.createdLLMCalls))
	}
	if repo.createdLLMCalls[0].Phase != model.VoiceLLMCallPhaseResolveHive || repo.createdLLMCalls[0].IsError {
		t.Errorf("expected a successful resolve_hive llm call row, got %+v", repo.createdLLMCalls[0])
	}
	if repo.createdLLMCalls[1].Phase != model.VoiceLLMCallPhaseProposeActions || repo.createdLLMCalls[1].IsError {
		t.Errorf("expected a successful propose_actions llm call row, got %+v", repo.createdLLMCalls[1])
	}

	var resolveReq resolveHiveRequest
	if err := json.Unmarshal(repo.createdLLMCalls[0].Request, &resolveReq); err != nil {
		t.Fatalf("unmarshal resolve_hive request: %v", err)
	}
	if resolveReq.Transcript != "hive three looked good" || len(resolveReq.Hives) != 1 || resolveReq.Hives[0] != (hiveOption{ID: 42, Name: "Hive 3"}) {
		t.Errorf("expected resolve_hive request to contain the transcript and hive list actually sent, got %+v", resolveReq)
	}

	var proposeReq proposeActionsRequest
	if err := json.Unmarshal(repo.createdLLMCalls[1].Request, &proposeReq); err != nil {
		t.Fatalf("unmarshal propose_actions request: %v", err)
	}
	if proposeReq.Transcript != "hive three looked good" || proposeReq.HiveContext.Type != "langstroth" || !proposeReq.HiveContext.QueenNeedsReplacement || len(proposeReq.HiveContext.Diseases) != 1 || proposeReq.HiveContext.Diseases[0] != "varroa" {
		t.Errorf("expected propose_actions request to contain the transcript and hive context actually sent, got %+v", proposeReq)
	}
}

func TestVoiceWorker_ProcessNext_TranscribeCallGetsBoundedTimeout(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{Text: "  "}}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, &mockHiveLister{}, &mockHiveResolver{})

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}

	deadline, ok := transcriber.capturedCtx.Deadline()
	if !ok {
		t.Fatal("expected the context passed to Transcribe to carry a deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > transcribeCallTimeout {
		t.Errorf("expected transcribe deadline within %s, got %s remaining", transcribeCallTimeout, remaining)
	}
}

func TestVoiceWorker_ProcessNext_TranscribeContextCancellationSurfacesAsFailureInsteadOfHanging(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, RetryCount: 0}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{blockUntilCtxDone: true}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, &mockHiveLister{}, &mockHiveResolver{})

	parentCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := w.ProcessNext(parentCtx); err != nil {
			t.Errorf("ProcessNext() error = %v", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ProcessNext did not return once the transcribe context was cancelled -- worker is hanging")
	}

	if !repo.retryCalled {
		t.Fatalf("expected recording to be retried once the transcribe context was cancelled, failedCalled=%v", repo.failedCalled)
	}
	if repo.retryErr != context.DeadlineExceeded.Error() {
		t.Errorf("expected the deadline-exceeded error to be recorded, got %q", repo.retryErr)
	}
}

func TestVoiceWorker_ProcessNext_ResolveHiveCallGetsBoundedTimeout(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}

	deadline, ok := resolver.resolveCtx.Deadline()
	if !ok {
		t.Fatal("expected the context passed to resolve_hive's CreateChatCompletion to carry a deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > llmCallTimeout {
		t.Errorf("expected resolve_hive deadline within %s, got %s remaining", llmCallTimeout, remaining)
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsCallGetsBoundedTimeout(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}

	deadline, ok := resolver.proposeCtx.Deadline()
	if !ok {
		t.Fatal("expected the context passed to propose_actions's CreateChatCompletion to carry a deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > llmCallTimeout {
		t.Errorf("expected propose_actions deadline within %s, got %s remaining", llmCallTimeout, remaining)
	}
}

func TestVoiceWorker_ProcessNext_PoorAudioQuality(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "mumble mumble",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "mumble", AvgLogprob: -3.0, NoSpeechProb: 0.9}},
	}}
	audio := &mockAudioStore{}
	resolver := &mockHiveResolver{}
	w := newTranscribedWorker(repo, transcriber, audio, &mockHiveLister{}, resolver)

	processed, err := w.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if !repo.failedCalled || repo.failedErr != "POOR_AUDIO_QUALITY" {
		t.Fatalf("expected recording to be marked failed with POOR_AUDIO_QUALITY, got called=%v err=%q", repo.failedCalled, repo.failedErr)
	}
	if len(audio.deleted) != 1 {
		t.Errorf("expected audio file to be deleted, got %+v", audio.deleted)
	}
	if resolver.callCount != 0 {
		t.Error("expected hive resolution to be skipped on poor audio quality")
	}
}

func TestVoiceWorker_ProcessNext_EmptyTranscriptIsPoorQuality(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{Text: "  ", Segments: nil}}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, &mockHiveLister{}, &mockHiveResolver{})

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled || repo.failedErr != "POOR_AUDIO_QUALITY" {
		t.Fatalf("expected POOR_AUDIO_QUALITY, got called=%v err=%q", repo.failedCalled, repo.failedErr)
	}
}

func TestVoiceWorker_ProcessNext_TransientErrorRetries(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, RetryCount: 0}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{err: &llm.WhisperAPIError{StatusCode: http.StatusServiceUnavailable, Body: "down"}}
	audio := &mockAudioStore{}
	w := newTranscribedWorker(repo, transcriber, audio, &mockHiveLister{}, &mockHiveResolver{})

	processed, err := w.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if !repo.retryCalled {
		t.Fatal("expected recording to be retried")
	}
	if repo.failedCalled {
		t.Error("expected recording NOT to be marked failed on first transient error")
	}
	if len(audio.deleted) != 0 {
		t.Errorf("expected audio file to be kept for retry, got deleted=%+v", audio.deleted)
	}
	wantBackoff := 5 * time.Second
	if got := time.Until(repo.retryNextAttemptAt); got < wantBackoff-time.Second || got > wantBackoff+time.Second {
		t.Errorf("expected ~%s backoff, got %s", wantBackoff, got)
	}
}

func TestVoiceWorker_ProcessNext_TransientErrorExhaustsRetriesToFailed(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, RetryCount: model.MaxWhisperRetries - 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{err: &llm.WhisperAPIError{StatusCode: http.StatusTooManyRequests, Body: "rate limited"}}
	audio := &mockAudioStore{}
	w := newTranscribedWorker(repo, transcriber, audio, &mockHiveLister{}, &mockHiveResolver{})

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed once retries are exhausted")
	}
	if repo.retryCalled {
		t.Error("expected recording NOT to be retried once MaxWhisperRetries is reached")
	}
	if len(audio.deleted) != 1 {
		t.Errorf("expected audio file to be deleted on terminal failure, got %+v", audio.deleted)
	}
}

func TestVoiceWorker_ProcessNext_NonTransientErrorFailsImmediately(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, RetryCount: 0}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{err: &llm.WhisperAPIError{StatusCode: http.StatusBadRequest, Body: "corrupt file"}}
	audio := &mockAudioStore{}
	w := newTranscribedWorker(repo, transcriber, audio, &mockHiveLister{}, &mockHiveResolver{})

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed immediately")
	}
	if repo.retryCalled {
		t.Error("expected recording NOT to be retried on a non-transient error")
	}
	if len(audio.deleted) != 1 {
		t.Errorf("expected audio file to be deleted, got %+v", audio.deleted)
	}
}

func TestVoiceWorker_ProcessNext_NilAudioPathFails(t *testing.T) {
	rec := &model.VoiceRecording{ID: 1}
	repo := &mockVoiceRepo{next: rec}
	audio := &mockAudioStore{}
	w := newTranscribedWorker(repo, &mockTranscriber{}, audio, &mockHiveLister{}, &mockHiveResolver{})

	processed, err := w.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if !repo.failedCalled || repo.failedErr != "recording has no audio file" {
		t.Fatalf("expected recording to be marked failed for missing audio path, got called=%v err=%q", repo.failedCalled, repo.failedErr)
	}
	if len(audio.deleted) != 0 {
		t.Errorf("expected no delete attempt when there is no audio path, got %+v", audio.deleted)
	}
}

func TestVoiceWorker_ProcessNext_OpenAudioErrorFails(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	audio := &mockAudioStore{openErr: errors.New("file not found")}
	w := newTranscribedWorker(repo, &mockTranscriber{}, audio, &mockHiveLister{}, &mockHiveResolver{})

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when the audio file can't be opened")
	}
}

func TestVoiceWorker_ProcessNext_HiveNotIdentified(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "the bees looked fine",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "the bees looked fine", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}, {ID: 2, Name: "Hive 2"}}}
	spoken := "the bees"
	resolver := &mockHiveResolver{resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeNotIdentified, SpokenHiveName: &spoken})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.completedCalled {
		t.Fatal("expected recording to still be marked completed")
	}
	if !repo.createActionCalled {
		t.Fatal("expected a voice_actions error row to be created")
	}
	if repo.createdAction.ErrorMessage == nil || *repo.createdAction.ErrorMessage != "HIVE_NOT_IDENTIFIED" {
		t.Errorf("expected HIVE_NOT_IDENTIFIED error, got %+v", repo.createdAction.ErrorMessage)
	}
	if repo.createdAction.HiveID != nil {
		t.Errorf("expected nil hive_id, got %+v", repo.createdAction.HiveID)
	}
	if repo.createdAction.SpokenHiveName == nil || *repo.createdAction.SpokenHiveName != "the bees" {
		t.Errorf("expected spoken_hive_name to be preserved, got %+v", repo.createdAction.SpokenHiveName)
	}
}

func TestVoiceWorker_ProcessNext_MultipleHivesMentioned(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive 3 looked good, hive 4 needs feeding",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive 3 looked good, hive 4 needs feeding", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 3, Name: "Hive 3"}, {ID: 4, Name: "Hive 4"}}}
	resolver := &mockHiveResolver{resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMultiple})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.createActionCalled || repo.createdAction.ErrorMessage == nil || *repo.createdAction.ErrorMessage != "MULTIPLE_HIVES_MENTIONED" {
		t.Fatalf("expected MULTIPLE_HIVES_MENTIONED error action, got %+v", repo.createdAction)
	}
	if !repo.completedCalled {
		t.Error("expected recording to still be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_NoHivesSkipsClaudeCall(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: nil}
	resolver := &mockHiveResolver{}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if resolver.callCount != 0 {
		t.Error("expected no Claude call when the apiary has no hives")
	}
	if !repo.createActionCalled || *repo.createdAction.ErrorMessage != "HIVE_NOT_IDENTIFIED" {
		t.Fatalf("expected a HIVE_NOT_IDENTIFIED error action, got %+v", repo.createdAction)
	}
}

func TestVoiceWorker_ProcessNext_HallucinatedHiveIDFallsBackToNotIdentified(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}}}
	bogusID := int64(999)
	resolver := &mockHiveResolver{resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &bogusID})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.createActionCalled || *repo.createdAction.ErrorMessage != "HIVE_NOT_IDENTIFIED" {
		t.Fatalf("expected a hive_id outside the known list to fall back to HIVE_NOT_IDENTIFIED, got %+v", repo.createdAction)
	}
}

func TestVoiceWorker_ProcessNext_ResolveHiveForcesToolChoiceButProposeActionsDoesNot(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(resolver.requests) != 2 {
		t.Fatalf("expected 2 llm requests (resolve_hive + propose_actions), got %d", len(resolver.requests))
	}

	resolveReq := resolver.requests[0]
	if resolveReq.ToolChoice == nil || resolveReq.ToolChoice.Type != "function" || resolveReq.ToolChoice.Function.Name != resolveHiveToolName {
		t.Errorf("expected resolve_hive request to force tool_choice=%q, got %+v", resolveHiveToolName, resolveReq.ToolChoice)
	}
	if len(resolveReq.Tools) != 1 || resolveReq.Tools[0].Function.Name != resolveHiveToolName {
		t.Errorf("expected resolve_hive request to carry exactly the resolve_hive tool, got %+v", resolveReq.Tools)
	}

	proposeReq := resolver.requests[1]
	if proposeReq.ToolChoice != nil {
		t.Errorf("expected propose_actions request to leave tool choice unset, got %+v", proposeReq.ToolChoice)
	}
}

func TestVoiceWorker_ProcessNext_ResolveHiveEmptyChoicesFailsRecording(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}}}
	resolver := &mockHiveResolver{resolveMessage: &llm.ChatCompletionResponse{Choices: []llm.Choice{}}}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when the response has zero choices")
	}
	if repo.completedCalled {
		t.Error("expected recording NOT to be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_ClaudeNoToolCallFailsRecording(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}}}
	resolver := &mockHiveResolver{resolveMessage: &llm.ChatCompletionResponse{
		Choices: []llm.Choice{{Message: llm.ChatMessage{Role: "assistant", Content: "sorry, I can't help with that"}}},
	}}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when the model never calls resolve_hive")
	}
	if repo.completedCalled {
		t.Error("expected recording NOT to be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_MalformedToolInputFailsRecording(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}}}
	resolver := &mockHiveResolver{resolveMessage: toolCallResponse([2]string{resolveHiveToolName, "{not valid json"})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when resolve_hive input is unparseable")
	}
	if repo.completedCalled {
		t.Error("expected recording NOT to be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_ClaudeErrorFailsRecording(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}}}
	resolver := &mockHiveResolver{resolveErr: errors.New("anthropic api unavailable")}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when hive resolution fails")
	}
	if repo.completedCalled {
		t.Error("expected recording NOT to be marked completed")
	}
	if len(repo.createdLLMCalls) != 1 || !repo.createdLLMCalls[0].IsError || repo.createdLLMCalls[0].Phase != model.VoiceLLMCallPhaseResolveHive {
		t.Fatalf("expected a failed resolve_hive llm call row, got %+v", repo.createdLLMCalls)
	}
	var stored map[string]string
	if err := json.Unmarshal(repo.createdLLMCalls[0].Response, &stored); err != nil {
		t.Fatalf("unmarshal stored error response: %v", err)
	}
	if stored["error"] != "anthropic api unavailable" {
		t.Errorf("expected the raw Claude error to be stored, got %+v", stored)
	}
}

func TestVoiceWorker_ProcessNext_LLMCallLogFailureIsBestEffort(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec, createLLMCallErr: errors.New("db unavailable")}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.completedCalled {
		t.Error("expected recording to still be marked completed even when logging the llm call fails")
	}
}

func TestVoiceWorker_Run_SweepsStuckProcessingEachTick(t *testing.T) {
	repo := &mockVoiceRepo{sweptCount: 2}
	w := newTranscribedWorker(repo, &mockTranscriber{}, &mockAudioStore{}, &mockHiveLister{}, &mockHiveResolver{})

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	w.Run(ctx, 10*time.Millisecond)

	if !repo.sweptCalled {
		t.Error("expected SweepStuckProcessing to be called")
	}
}

func TestVoiceWorker_ProcessNext_MatchedWithNilHiveIDFallsBackToNotIdentified(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 1, Name: "Hive 1"}}}
	resolver := &mockHiveResolver{resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.createActionCalled || *repo.createdAction.ErrorMessage != "HIVE_NOT_IDENTIFIED" {
		t.Fatalf("expected matched with nil hive_id to fall back to HIVE_NOT_IDENTIFIED, got %+v", repo.createdAction)
	}
}

func TestVoiceWorker_ProcessNext_ProposesSingleAction(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good, brood pattern was excellent",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good, brood pattern was excellent", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{
		hives:   []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}},
		summary: &mcp.HiveHistory{Hive: mcp.HiveSummary{ID: 42, Type: "langstroth"}},
	}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t, [2]string{model.VoiceActionToolCreateInspection, `{"brood_pattern":"excellent"}`}),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(repo.createdActions) != 1 {
		t.Fatalf("expected 1 proposed action, got %d: %+v", len(repo.createdActions), repo.createdActions)
	}
	action := repo.createdActions[0]
	if action.Status != model.VoiceActionStatusProposed {
		t.Errorf("expected status proposed, got %q", action.Status)
	}
	if action.HiveID == nil || *action.HiveID != hiveID {
		t.Errorf("expected hive_id %d, got %+v", hiveID, action.HiveID)
	}
	if action.ToolName == nil || *action.ToolName != model.VoiceActionToolCreateInspection {
		t.Errorf("expected tool_name create_inspection, got %+v", action.ToolName)
	}
	if string(action.ToolArguments) != `{"brood_pattern":"excellent"}` {
		t.Errorf("expected tool_arguments to be passed through verbatim, got %s", action.ToolArguments)
	}
	if action.Sequence != 1 {
		t.Errorf("expected sequence 1, got %d", action.Sequence)
	}
	if !repo.completedCalled {
		t.Error("expected recording to be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_SkipsHiveResolutionWhenHiveIDAlreadySet(t *testing.T) {
	hiveID := int64(42)
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3, HiveID: &hiveID}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good, brood pattern was excellent",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good, brood pattern was excellent", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hives := &mockHiveLister{
		hives:   []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}},
		summary: &mcp.HiveHistory{Hive: mcp.HiveSummary{ID: 42, Type: "langstroth"}},
	}
	resolver := &mockHiveResolver{
		proposeMessage: toolUseMessage(t, [2]string{model.VoiceActionToolCreateInspection, `{"brood_pattern":"excellent"}`}),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if hives.callCount != 0 {
		t.Errorf("expected ListHives (hive resolution) never called, got %d calls", hives.callCount)
	}
	if resolver.callCount != 1 {
		t.Fatalf("expected exactly 1 LLM call (propose_actions only), got %d", resolver.callCount)
	}
	if len(repo.createdActions) != 1 {
		t.Fatalf("expected 1 proposed action, got %d: %+v", len(repo.createdActions), repo.createdActions)
	}
	action := repo.createdActions[0]
	if action.HiveID == nil || *action.HiveID != hiveID {
		t.Errorf("expected hive_id %d, got %+v", hiveID, action.HiveID)
	}
	if !repo.completedCalled {
		t.Error("expected recording to be marked completed")
	}
	if len(repo.createdLLMCalls) != 1 || repo.createdLLMCalls[0].Phase != model.VoiceLLMCallPhaseProposeActions {
		t.Errorf("expected only a propose_actions llm call row, got %+v", repo.createdLLMCalls)
	}
}

func TestVoiceWorker_ProcessNext_ProposesMultipleActionsInOrder(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good and I gave them a litre of syrup",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good and I gave them a litre of syrup", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t,
			[2]string{model.VoiceActionToolCreateInspection, `{"brood_pattern":"good"}`},
			[2]string{model.VoiceActionToolCreateFeeding, `{"feed_type":"syrup","amount":"1L"}`},
		),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(repo.createdActions) != 2 {
		t.Fatalf("expected 2 proposed actions, got %d: %+v", len(repo.createdActions), repo.createdActions)
	}
	if *repo.createdActions[0].ToolName != model.VoiceActionToolCreateInspection || repo.createdActions[0].Sequence != 1 {
		t.Errorf("expected first action to be create_inspection with sequence 1, got %+v", repo.createdActions[0])
	}
	if *repo.createdActions[1].ToolName != model.VoiceActionToolCreateFeeding || repo.createdActions[1].Sequence != 2 {
		t.Errorf("expected second action to be create_feeding with sequence 2, got %+v", repo.createdActions[1])
	}
}

func TestVoiceWorker_ProcessNext_UnrecognizedToolNameIgnored(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t, [2]string{"delete_hive", `{}`}),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if repo.createActionCalled {
		t.Errorf("expected an unrecognized tool name to be ignored, got %+v", repo.createdActions)
	}
	if !repo.completedCalled {
		t.Error("expected recording to still be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_MixedRecognizedAndUnrecognizedToolCalls(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good and I gave them a litre of syrup",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good and I gave them a litre of syrup", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t,
			[2]string{model.VoiceActionToolCreateInspection, `{"brood_pattern":"good"}`},
			[2]string{"delete_hive", `{}`},
			[2]string{model.VoiceActionToolCreateFeeding, `{"feed_type":"syrup","amount":"1L"}`},
		),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(repo.createdActions) != 2 {
		t.Fatalf("expected the unrecognized tool call to be skipped leaving 2 proposed actions, got %d: %+v", len(repo.createdActions), repo.createdActions)
	}
	if *repo.createdActions[0].ToolName != model.VoiceActionToolCreateInspection || repo.createdActions[0].Sequence != 1 {
		t.Errorf("expected first action to be create_inspection with sequence 1, got %+v", repo.createdActions[0])
	}
	if *repo.createdActions[1].ToolName != model.VoiceActionToolCreateFeeding || repo.createdActions[1].Sequence != 2 {
		t.Errorf("expected second action to be create_feeding with sequence 2 (sequence counts only recognized tools), got %+v", repo.createdActions[1])
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsEmptyChoicesProposesNothing(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: &llm.ChatCompletionResponse{Choices: []llm.Choice{}},
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if repo.createActionCalled {
		t.Errorf("expected no proposed actions when the response has zero choices, got %+v", repo.createdActions)
	}
	if !repo.completedCalled {
		t.Error("expected recording to still be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsTruncatedByLengthRecordsIncompleteError(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good and I gave them a litre of syrup",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good and I gave them a litre of syrup", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	truncated := toolUseMessage(t, [2]string{model.VoiceActionToolCreateInspection, `{"brood_pattern":"good"}`})
	truncated.Choices[0].FinishReason = "length"
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: truncated,
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(repo.createdActions) != 1 {
		t.Fatalf("expected 1 recorded action, got %d: %+v", len(repo.createdActions), repo.createdActions)
	}
	action := repo.createdActions[0]
	if action.Status != model.VoiceActionStatusError {
		t.Errorf("expected status error, got %q", action.Status)
	}
	if action.HiveID == nil || *action.HiveID != hiveID {
		t.Errorf("expected hive_id %d, got %+v", hiveID, action.HiveID)
	}
	if action.ErrorMessage == nil || *action.ErrorMessage != errProposalIncomplete {
		t.Errorf("expected error_message %q, got %+v", errProposalIncomplete, action.ErrorMessage)
	}
	if action.ToolName != nil {
		t.Errorf("expected no tool_name on a truncated proposal, got %+v", action.ToolName)
	}
	if !repo.completedCalled {
		t.Error("expected recording to still be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsClaudeErrorFailsRecording(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeErr:     errors.New("anthropic api unavailable"),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when action proposal fails")
	}
	if repo.completedCalled {
		t.Error("expected recording NOT to be marked completed")
	}
	if len(repo.createdLLMCalls) != 2 {
		t.Fatalf("expected 2 llm call rows (resolve_hive success + propose_actions failure), got %d", len(repo.createdLLMCalls))
	}
	if repo.createdLLMCalls[1].Phase != model.VoiceLLMCallPhaseProposeActions || !repo.createdLLMCalls[1].IsError {
		t.Errorf("expected a failed propose_actions llm call row, got %+v", repo.createdLLMCalls[1])
	}
}

func TestVoiceWorker_ProcessNext_HiveSummaryErrorFailsRecording(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{
		hives:      []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}},
		summaryErr: errors.New("database unavailable"),
	}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when the hive summary lookup fails")
	}
	if resolver.callCount != 1 {
		t.Errorf("expected Claude not to be called for action proposal when the hive summary lookup fails, got %d calls", resolver.callCount)
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsIncludesKnownValues(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	suggestions := &mockSuggestionProvider{
		medicineNames: []string{"Apiguard"},
		doses:         []string{"1 strip"},
		feedTypes:     []string{"sugar syrup"},
		amounts:       []string{"1L"},
	}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorkerWithSuggestions(repo, transcriber, &mockAudioStore{}, hives, suggestions, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(resolver.requests) != 2 {
		t.Fatalf("expected 2 llm requests (resolve_hive + propose_actions), got %d", len(resolver.requests))
	}

	proposeReq := resolver.requests[1]
	if len(proposeReq.Messages) != 2 {
		t.Fatalf("expected 2 messages in propose_actions request, got %d", len(proposeReq.Messages))
	}
	userContent := proposeReq.Messages[1].Content
	var content struct {
		HiveContext hiveActionContext `json:"hive_context"`
		KnownValues knownValues       `json:"known_values"`
	}
	if idx := strings.Index(userContent, "Context:\n"); idx >= 0 {
		if err := json.Unmarshal([]byte(userContent[idx+len("Context:\n"):]), &content); err != nil {
			t.Fatalf("unmarshal user message context: %v", err)
		}
	} else {
		t.Fatalf("expected user message to contain a Context: section, got %q", userContent)
	}
	want := knownValues{MedicineNames: []string{"Apiguard"}, Doses: []string{"1 strip"}, FeedTypes: []string{"sugar syrup"}, Amounts: []string{"1L"}}
	if !reflect.DeepEqual(content.KnownValues, want) {
		t.Errorf("expected known_values in the LLM request to be %+v, got %+v", want, content.KnownValues)
	}

	var proposeReqLogged proposeActionsRequest
	if err := json.Unmarshal(repo.createdLLMCalls[1].Request, &proposeReqLogged); err != nil {
		t.Fatalf("unmarshal propose_actions request: %v", err)
	}
	if !reflect.DeepEqual(proposeReqLogged.KnownValues, want) {
		t.Errorf("expected logged known_values to be %+v, got %+v", want, proposeReqLogged.KnownValues)
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsStatesTranscriptLanguageExplicitly(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "Zrobiłem dzisiaj przegląd ula",
		Language: "polish",
		Segments: []llm.TranscriptionSegment{{Text: "Zrobiłem dzisiaj przegląd ula", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(resolver.requests) != 2 {
		t.Fatalf("expected 2 llm requests (resolve_hive + propose_actions), got %d", len(resolver.requests))
	}

	proposeReq := resolver.requests[1]
	if len(proposeReq.Messages) != 2 {
		t.Fatalf("expected 2 messages in propose_actions request, got %d", len(proposeReq.Messages))
	}
	userContent := proposeReq.Messages[1].Content
	if !strings.HasPrefix(userContent, "Transcript language: polish\n") {
		t.Errorf("expected the propose_actions user message to state the transcript language up front, got %q", userContent)
	}

	var proposeReqLogged proposeActionsRequest
	if err := json.Unmarshal(repo.createdLLMCalls[1].Request, &proposeReqLogged); err != nil {
		t.Fatalf("unmarshal propose_actions request: %v", err)
	}
	if proposeReqLogged.Language != "polish" {
		t.Errorf("expected logged language to be %q, got %q", "polish", proposeReqLogged.Language)
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsFallsBackToUnknownLanguageLabel(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}

	userContent := resolver.requests[1].Messages[1].Content
	if !strings.HasPrefix(userContent, "Transcript language: "+unknownTranscriptLanguage+"\n") {
		t.Errorf("expected an unknown-language fallback label when Whisper detected none, got %q", userContent)
	}
}

func TestVoiceWorker_FetchKnownValues_NilSuggestionProviderReturnsEmpty(t *testing.T) {
	w := newTranscribedWorker(&mockVoiceRepo{}, &mockTranscriber{}, &mockAudioStore{}, &mockHiveLister{}, &mockHiveResolver{})

	got := w.fetchKnownValues(context.Background(), 7)

	if !reflect.DeepEqual(got, knownValues{}) {
		t.Errorf("expected an empty knownValues when suggestions is nil, got %+v", got)
	}
}

func TestVoiceWorker_FetchKnownValues_PartialErrorsReturnRemainingLists(t *testing.T) {
	suggestions := &mockSuggestionProvider{
		medicineNames: []string{"Apiguard"},
		dosesErr:      errors.New("db unavailable"),
		feedTypes:     []string{"sugar syrup"},
		amounts:       []string{"1L"},
	}
	w := newTranscribedWorkerWithSuggestions(&mockVoiceRepo{}, &mockTranscriber{}, &mockAudioStore{}, &mockHiveLister{}, suggestions, &mockHiveResolver{})

	got := w.fetchKnownValues(context.Background(), 7)

	want := knownValues{MedicineNames: []string{"Apiguard"}, Doses: nil, FeedTypes: []string{"sugar syrup"}, Amounts: []string{"1L"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected the three successful lists despite one error, got %+v, want %+v", got, want)
	}
}

func TestVoiceWorker_ProcessNext_ProposeActionsNilSuggestionsOmitsKnownValues(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	var content struct {
		KnownValues knownValues `json:"known_values"`
	}
	userContent := resolver.requests[1].Messages[1].Content
	if idx := strings.Index(userContent, "Context:\n"); idx >= 0 {
		if err := json.Unmarshal([]byte(userContent[idx+len("Context:\n"):]), &content); err != nil {
			t.Fatalf("unmarshal user message context: %v", err)
		}
	} else {
		t.Fatalf("expected user message to contain a Context: section, got %q", userContent)
	}
	if !reflect.DeepEqual(content.KnownValues, knownValues{}) {
		t.Errorf("expected known_values to be empty when suggestions is nil, got %+v", content.KnownValues)
	}

	var proposeReqLogged proposeActionsRequest
	if err := json.Unmarshal(repo.createdLLMCalls[1].Request, &proposeReqLogged); err != nil {
		t.Fatalf("unmarshal propose_actions request: %v", err)
	}
	if !reflect.DeepEqual(proposeReqLogged.KnownValues, knownValues{}) {
		t.Errorf("expected logged known_values to be empty when suggestions is nil, got %+v", proposeReqLogged.KnownValues)
	}
}

func TestVoiceWorker_ProcessNext_MalformedProposedToolInputStoredVerbatim(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t, [2]string{model.VoiceActionToolCreateInspection, `{not valid json`}),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(repo.createdActions) != 1 {
		t.Fatalf("expected 1 proposed action even with malformed tool input, got %d", len(repo.createdActions))
	}
	if string(repo.createdActions[0].ToolArguments) != `{not valid json` {
		t.Errorf("expected malformed tool_arguments to be stored verbatim without parsing, got %s", repo.createdActions[0].ToolArguments)
	}
	if !repo.completedCalled {
		t.Error("expected recording to be marked completed")
	}
}

func TestVoiceWorker_ProcessNext_SignedFrameDeltaSurvivesRoundTrip(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three lost a frame of brood",
		Segments: []llm.TranscriptionSegment{{Text: "hive three lost a frame of brood", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	hiveID := int64(42)
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	resolver := &mockHiveResolver{
		resolveMessage: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID}),
		proposeMessage: toolUseMessage(t, [2]string{model.VoiceActionToolCreateInspection, `{"frames_added_brood":-1}`}),
	}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if len(repo.createdActions) != 1 {
		t.Fatalf("expected 1 proposed action, got %d", len(repo.createdActions))
	}
	var got map[string]int
	if err := json.Unmarshal(repo.createdActions[0].ToolArguments, &got); err != nil {
		t.Fatalf("unmarshal stored tool_arguments: %v", err)
	}
	if got["frames_added_brood"] != -1 {
		t.Errorf("expected frames_added_brood -1 to survive the raw JSON round-trip, got %d", got["frames_added_brood"])
	}
}

func TestBuildHiveActionContext_PicksLatestInspection(t *testing.T) {
	older := time.Now().Add(-48 * time.Hour)
	newer := time.Now()
	framesBrood := 5
	h := &mcp.HiveHistory{
		Hive: mcp.HiveSummary{Type: "langstroth", QueenNeedsReplacement: true, Diseases: []string{"varroa"}},
		Inspections: []mcp.InspectionSummary{
			{InspectedAt: older, FramesBrood: nil},
			{InspectedAt: newer, FramesBrood: &framesBrood},
		},
	}

	got := buildHiveActionContext(h)

	if got.Type != "langstroth" || !got.QueenNeedsReplacement || len(got.Diseases) != 1 || got.Diseases[0] != "varroa" {
		t.Errorf("expected hive flags to carry through, got %+v", got)
	}
	if got.LastInspection == nil || got.LastInspection.FramesBrood == nil || *got.LastInspection.FramesBrood != 5 {
		t.Fatalf("expected the most recent inspection to be picked regardless of list order, got %+v", got.LastInspection)
	}
}

func TestChatToolCalls(t *testing.T) {
	tests := []struct {
		name string
		resp *llm.ChatCompletionResponse
		want int
	}{
		{"nil choices", &llm.ChatCompletionResponse{Choices: nil}, 0},
		{"empty choices slice", &llm.ChatCompletionResponse{Choices: []llm.Choice{}}, 0},
		{"one choice with tool calls", toolCallResponse([2]string{"a", "{}"}, [2]string{"b", "{}"}), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chatToolCalls(tt.resp); len(got) != tt.want {
				t.Errorf("chatToolCalls() = %+v, want %d calls", got, tt.want)
			}
		})
	}
}

func TestContainsHiveID(t *testing.T) {
	options := []hiveOption{{ID: 1, Name: "Hive 1"}, {ID: 2, Name: "Hive 2"}}
	tests := []struct {
		name string
		id   int64
		want bool
	}{
		{"present", 1, true},
		{"also present", 2, true},
		{"absent", 99, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsHiveID(options, tt.id); got != tt.want {
				t.Errorf("containsHiveID(%v, %d) = %v, want %v", options, tt.id, got, tt.want)
			}
		})
	}
	if containsHiveID(nil, 1) {
		t.Error("containsHiveID(nil, 1) = true, want false")
	}
}

func toolSchema(t *testing.T, tool llm.Tool) (map[string]any, []string) {
	t.Helper()
	var schema struct {
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required"`
	}
	if err := json.Unmarshal(tool.Function.Parameters, &schema); err != nil {
		t.Fatalf("unmarshal tool parameters: %v", err)
	}
	return schema.Properties, schema.Required
}

func TestResolveHiveTool_Schema(t *testing.T) {
	tool := resolveHiveTool()
	if tool.Function.Name != resolveHiveToolName {
		t.Errorf("Name = %q, want %q", tool.Function.Name, resolveHiveToolName)
	}
	properties, required := toolSchema(t, tool)
	if len(required) != 1 || required[0] != "outcome" {
		t.Errorf("Required = %+v, want [outcome]", required)
	}

	outcomeProp, ok := properties["outcome"].(map[string]any)
	if !ok {
		t.Fatal("expected outcome property to be a map")
	}
	enumRaw, ok := outcomeProp["enum"].([]any)
	if !ok {
		t.Fatal("expected outcome enum to be a []any")
	}
	wantEnum := []string{resolveHiveOutcomeMatched, resolveHiveOutcomeNotIdentified, resolveHiveOutcomeMultiple}
	if len(enumRaw) != len(wantEnum) {
		t.Fatalf("enum = %+v, want %+v", enumRaw, wantEnum)
	}
	for i, v := range wantEnum {
		if enumRaw[i] != v {
			t.Errorf("enum[%d] = %q, want %q", i, enumRaw[i], v)
		}
	}

	for _, key := range []string{"hive_id", "spoken_hive_name"} {
		if _, ok := properties[key]; !ok {
			t.Errorf("expected properties to contain %q", key)
		}
	}
}

func TestActionProposalTools_RequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		tool     llm.Tool
		wantName string
		wantReq  []string
	}{
		{"inspection", createInspectionTool(), model.VoiceActionToolCreateInspection, nil},
		{"treatment", createTreatmentTool(), model.VoiceActionToolCreateTreatment, []string{"medicine_name"}},
		{"harvest", createHarvestTool(), model.VoiceActionToolCreateHarvest, nil},
		{"feeding", createFeedingTool(), model.VoiceActionToolCreateFeeding, []string{"feed_type"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tool.Function.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tt.tool.Function.Name, tt.wantName)
			}
			_, required := toolSchema(t, tt.tool)
			if len(required) != len(tt.wantReq) {
				t.Fatalf("Required = %+v, want %+v", required, tt.wantReq)
			}
			for i, key := range tt.wantReq {
				if required[i] != key {
					t.Errorf("Required[%d] = %q, want %q", i, required[i], key)
				}
			}
		})
	}
}

func TestActionProposalTools_ReturnsAllFive(t *testing.T) {
	tools := actionProposalTools()
	if len(tools) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(tools))
	}
	wantNames := map[string]bool{
		model.VoiceActionToolCreateInspection: false,
		model.VoiceActionToolCreateTreatment:  false,
		model.VoiceActionToolCreateHarvest:    false,
		model.VoiceActionToolCreateFeeding:    false,
		model.VoiceActionToolUpdateHiveStatus: false,
	}
	for _, tool := range tools {
		if _, ok := wantNames[tool.Function.Name]; !ok {
			t.Errorf("unexpected tool name %q", tool.Function.Name)
		}
		wantNames[tool.Function.Name] = true
	}
	for name, seen := range wantNames {
		if !seen {
			t.Errorf("expected tool %q to be present", name)
		}
	}
}

func TestIsTransientWhisperError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"5xx", &llm.WhisperAPIError{StatusCode: http.StatusInternalServerError}, true},
		{"429", &llm.WhisperAPIError{StatusCode: http.StatusTooManyRequests}, true},
		{"400", &llm.WhisperAPIError{StatusCode: http.StatusBadRequest}, false},
		{"404", &llm.WhisperAPIError{StatusCode: http.StatusNotFound}, false},
		{"generic error", errors.New("boom"), false},
		{"network error", &net.OpError{Op: "dial", Err: errors.New("connection refused")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientWhisperError(tt.err); got != tt.want {
				t.Errorf("isTransientWhisperError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
