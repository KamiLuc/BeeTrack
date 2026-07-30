package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beetrack/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidAudioType     = errors.New("unsupported audio type; allowed: audio/webm, audio/mp4, audio/wav")
	ErrRecordingTooLong     = errors.New("recording exceeds 15 MB limit")
	ErrMaxRecordingsReached = errors.New("you already have the maximum of 20 stored recordings")
)

// MaxAudioBytes bounds an uploaded recording's size as a proxy for the client's 3-minute duration
// cap (VC-18-FE), sized to the worst case: a 3-minute mono 16-bit 44.1kHz WAV is ~15MB, the largest
// of the accepted formats — avoids needing to parse webm/m4a/wav headers for an exact duration.
const MaxAudioBytes = 15 * 1024 * 1024

const maxRecordingsPerUser = 20

var allowedAudioMIME = map[string]string{
	"audio/webm":  ".webm",
	"audio/mp4":   ".m4a",
	"audio/wav":   ".wav",
	"audio/x-wav": ".wav",
}

type VoiceRepository interface {
	CreateRecording(ctx context.Context, rec *model.VoiceRecording) error
	CountRecordingsByUserID(ctx context.Context, userID int64) (int64, error)
}

type VoiceService struct {
	apiaries    ApiaryMembershipReader
	recordings  VoiceRepository
	storagePath string
}

func NewVoiceService(apiaries ApiaryMembershipReader, recordings VoiceRepository, storagePath string) *VoiceService {
	return &VoiceService{apiaries: apiaries, recordings: recordings, storagePath: storagePath}
}

func (s *VoiceService) Upload(ctx context.Context, userID, apiaryID int64, mimeType string, data []byte) (*model.VoiceRecording, error) {
	if len(data) > MaxAudioBytes {
		return nil, ErrRecordingTooLong
	}
	ext, ok := allowedAudioMIME[mimeType]
	if !ok {
		return nil, ErrInvalidAudioType
	}
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}
	count, err := s.recordings.CountRecordingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count recordings: %w", err)
	}
	if count >= maxRecordingsPerUser {
		return nil, ErrMaxRecordingsReached
	}
	if err := os.MkdirAll(s.storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	filename := uuid.New().String() + ext
	if err := os.WriteFile(filepath.Join(s.storagePath, filename), data, 0o644); err != nil {
		return nil, fmt.Errorf("write audio: %w", err)
	}
	rec := &model.VoiceRecording{
		UserID:    userID,
		ApiaryID:  apiaryID,
		Status:    model.VoiceRecordingStatusPending,
		AudioPath: &filename,
	}
	if err := s.recordings.CreateRecording(ctx, rec); err != nil {
		_ = os.Remove(filepath.Join(s.storagePath, filename))
		return nil, fmt.Errorf("create recording: %w", err)
	}
	return rec, nil
}
