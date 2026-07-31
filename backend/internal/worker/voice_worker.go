package worker

import (
	"context"
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
	"github.com/beetrack/backend/internal/model"
)

const (
	stuckProcessingTimeout = 5 * time.Minute

	poorQualityAvgLogprob      = -1.0
	poorQualityNoSpeechProb    = 0.6
	poorQualitySegmentFraction = 0.5
)

type VoiceRecordingRepository interface {
	ClaimNext(ctx context.Context) (*model.VoiceRecording, error)
	MarkCompleted(ctx context.Context, id int64, transcript, language string) error
	MarkFailed(ctx context.Context, id int64, errMsg string) error
	MarkRetry(ctx context.Context, id int64, errMsg string, nextAttemptAt time.Time) error
	SweepStuckProcessing(ctx context.Context, olderThan time.Duration) (int64, error)
}

type Transcriber interface {
	Transcribe(ctx context.Context, audio io.Reader, filename string) (*llm.TranscriptionResult, error)
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
}

func NewVoiceWorker(recordings VoiceRecordingRepository, transcriber Transcriber, audio AudioStore) *VoiceWorker {
	return &VoiceWorker{recordings: recordings, transcriber: transcriber, audio: audio}
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
