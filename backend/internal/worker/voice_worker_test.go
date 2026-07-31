package worker

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/llm"
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

func clearAudioPath(rec *model.VoiceRecording, path string) *model.VoiceRecording {
	rec.AudioPath = &path
	return rec
}

func TestVoiceWorker_ProcessNext_NoRecordingAvailable(t *testing.T) {
	w := NewVoiceWorker(&mockVoiceRepo{}, &mockTranscriber{}, &mockAudioStore{})

	processed, err := w.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if processed {
		t.Error("expected processed=false when no recording is claimable")
	}
}

func TestVoiceWorker_ProcessNext_HappyPath(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{
		Text:     "hive three looked good",
		Language: "en",
		Segments: []llm.TranscriptionSegment{{Text: "hive three looked good", AvgLogprob: -0.2, NoSpeechProb: 0.05}},
	}}
	audio := &mockAudioStore{}
	w := NewVoiceWorker(repo, transcriber, audio)

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
	w := NewVoiceWorker(repo, transcriber, audio)

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
}

func TestVoiceWorker_ProcessNext_EmptyTranscriptIsPoorQuality(t *testing.T) {
	rec := clearAudioPath(&model.VoiceRecording{ID: 1}, "audio.wav")
	repo := &mockVoiceRepo{next: rec}
	transcriber := &mockTranscriber{result: &llm.TranscriptionResult{Text: "  ", Segments: nil}}
	w := NewVoiceWorker(repo, transcriber, &mockAudioStore{})

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
	w := NewVoiceWorker(repo, transcriber, audio)

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
	w := NewVoiceWorker(repo, transcriber, audio)

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
	w := NewVoiceWorker(repo, transcriber, audio)

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
	w := NewVoiceWorker(repo, &mockTranscriber{}, audio)

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
	w := NewVoiceWorker(repo, &mockTranscriber{}, audio)

	if _, err := w.ProcessNext(context.Background()); err != nil {
		t.Fatalf("ProcessNext() error = %v", err)
	}
	if !repo.failedCalled {
		t.Fatal("expected recording to be marked failed when the audio file can't be opened")
	}
}

func TestVoiceWorker_Run_SweepsStuckProcessingEachTick(t *testing.T) {
	repo := &mockVoiceRepo{sweptCount: 2}
	w := NewVoiceWorker(repo, &mockTranscriber{}, &mockAudioStore{})

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	w.Run(ctx, 10*time.Millisecond)

	if !repo.sweptCalled {
		t.Error("expected SweepStuckProcessing to be called")
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
