package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
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
	return nil
}

type mockTranscriber struct {
	result *llm.TranscriptionResult
	err    error
}

func (m *mockTranscriber) Transcribe(ctx context.Context, audio io.Reader, filename string) (*llm.TranscriptionResult, error) {
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
	hives     []mcp.HiveSummary
	err       error
	callCount int
}

func (m *mockHiveLister) ListHives(ctx context.Context, userID int64, apiaryID *int64) ([]mcp.HiveSummary, error) {
	m.callCount++
	return m.hives, m.err
}

type mockHiveResolver struct {
	message   *anthropic.Message
	err       error
	callCount int
}

func (m *mockHiveResolver) New(ctx context.Context, params anthropic.MessageNewParams, opts ...option.RequestOption) (*anthropic.Message, error) {
	m.callCount++
	return m.message, m.err
}

func resolveHiveMessage(t *testing.T, input resolveHiveInput) *anthropic.Message {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal resolve_hive input: %v", err)
	}
	return &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{{Type: "tool_use", Name: resolveHiveToolName, Input: data}},
	}
}

func clearAudioPath(rec *model.VoiceRecording, path string) *model.VoiceRecording {
	rec.AudioPath = &path
	return rec
}

func newTranscribedWorker(repo *mockVoiceRepo, transcriber *mockTranscriber, audio *mockAudioStore, hives *mockHiveLister, resolver *mockHiveResolver) *VoiceWorker {
	return NewVoiceWorker(repo, transcriber, audio, hives, resolver, "")
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
	hives := &mockHiveLister{hives: []mcp.HiveSummary{{ID: 42, Name: "Hive 3"}}}
	hiveID := int64(42)
	resolver := &mockHiveResolver{message: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &hiveID})}
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
		t.Error("expected no voice_actions row when the hive resolved cleanly")
	}
	if len(audio.deleted) != 1 || audio.deleted[0] != "audio.wav" {
		t.Errorf("expected audio file to be deleted, got %+v", audio.deleted)
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
	resolver := &mockHiveResolver{message: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeNotIdentified, SpokenHiveName: &spoken})}
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
	resolver := &mockHiveResolver{message: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMultiple})}
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
	resolver := &mockHiveResolver{message: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched, HiveID: &bogusID})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.createActionCalled || *repo.createdAction.ErrorMessage != "HIVE_NOT_IDENTIFIED" {
		t.Fatalf("expected a hive_id outside the known list to fall back to HIVE_NOT_IDENTIFIED, got %+v", repo.createdAction)
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
	resolver := &mockHiveResolver{message: &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{{Type: "text", Text: "sorry, I can't help with that"}},
	}}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when Claude never calls resolve_hive")
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
	resolver := &mockHiveResolver{message: &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{{Type: "tool_use", Name: resolveHiveToolName, Input: []byte("{not valid json")}},
	}}
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
	resolver := &mockHiveResolver{err: errors.New("anthropic api unavailable")}
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
	resolver := &mockHiveResolver{message: resolveHiveMessage(t, resolveHiveInput{Outcome: resolveHiveOutcomeMatched})}
	w := newTranscribedWorker(repo, transcriber, &mockAudioStore{}, hives, resolver)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.createActionCalled || *repo.createdAction.ErrorMessage != "HIVE_NOT_IDENTIFIED" {
		t.Fatalf("expected matched with nil hive_id to fall back to HIVE_NOT_IDENTIFIED, got %+v", repo.createdAction)
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

func TestResolveHiveTool_Schema(t *testing.T) {
	tool := resolveHiveTool().OfTool
	if tool == nil {
		t.Fatal("expected OfTool to be set")
	}
	if tool.Name != resolveHiveToolName {
		t.Errorf("Name = %q, want %q", tool.Name, resolveHiveToolName)
	}
	if len(tool.InputSchema.Required) != 1 || tool.InputSchema.Required[0] != "outcome" {
		t.Errorf("Required = %+v, want [outcome]", tool.InputSchema.Required)
	}

	properties, ok := tool.InputSchema.Properties.(map[string]any)
	if !ok {
		t.Fatalf("expected InputSchema.Properties to be a map[string]any, got %T", tool.InputSchema.Properties)
	}
	outcomeProp, ok := properties["outcome"].(map[string]any)
	if !ok {
		t.Fatal("expected outcome property to be a map")
	}
	enum, ok := outcomeProp["enum"].([]string)
	if !ok {
		t.Fatal("expected outcome enum to be a []string")
	}
	wantEnum := []string{resolveHiveOutcomeMatched, resolveHiveOutcomeNotIdentified, resolveHiveOutcomeMultiple}
	if len(enum) != len(wantEnum) {
		t.Fatalf("enum = %+v, want %+v", enum, wantEnum)
	}
	for i, v := range wantEnum {
		if enum[i] != v {
			t.Errorf("enum[%d] = %q, want %q", i, enum[i], v)
		}
	}

	for _, key := range []string{"hive_id", "spoken_hive_name"} {
		if _, ok := properties[key]; !ok {
			t.Errorf("expected InputSchema.Properties to contain %q", key)
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
