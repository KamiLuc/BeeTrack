package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

type mockVoiceRepo struct {
	created   *model.VoiceRecording
	count     int64
	createErr error
}

func (m *mockVoiceRepo) CreateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	if m.createErr != nil {
		return m.createErr
	}
	rec.ID = 1
	m.created = rec
	return nil
}

func (m *mockVoiceRepo) CountRecordingsByUserID(ctx context.Context, userID int64) (int64, error) {
	return m.count, nil
}

func newTestVoiceService(t *testing.T) (*VoiceService, *mockApiaryMembershipReader, *mockVoiceRepo, string) {
	t.Helper()
	dir := t.TempDir()
	apiaryMock := &mockApiaryMembershipReader{}
	voiceMock := &mockVoiceRepo{}
	svc := NewVoiceService(apiaryMock, voiceMock, dir)
	return svc, apiaryMock, voiceMock, dir
}

func TestVoiceUpload_Success(t *testing.T) {
	svc, apiaryMock, voiceMock, dir := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}

	data := []byte{0x1A, 0x45, 0xDF, 0xA3}
	rec, err := svc.Upload(context.Background(), 1, 1, "audio/webm", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusPending {
		t.Errorf("unexpected status: %s", rec.Status)
	}
	if voiceMock.created == nil {
		t.Fatal("expected CreateRecording to be called")
	}
	if rec.AudioPath == nil {
		t.Fatal("expected AudioPath to be set")
	}
	if _, err := os.Stat(filepath.Join(dir, *rec.AudioPath)); err != nil {
		t.Errorf("expected audio file to be written: %v", err)
	}
}

func TestVoiceUpload_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestVoiceService(t)

	_, err := svc.Upload(context.Background(), 1, 1, "audio/webm", []byte{1})
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestVoiceUpload_InvalidType(t *testing.T) {
	svc, apiaryMock, _, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}

	_, err := svc.Upload(context.Background(), 1, 1, "audio/mpeg", []byte{1})
	if !errors.Is(err, ErrInvalidAudioType) {
		t.Errorf("expected ErrInvalidAudioType, got %v", err)
	}
}

func TestVoiceUpload_TooLarge(t *testing.T) {
	svc, _, _, _ := newTestVoiceService(t)

	data := make([]byte, MaxAudioBytes+1)
	_, err := svc.Upload(context.Background(), 1, 1, "audio/webm", data)
	if !errors.Is(err, ErrRecordingTooLong) {
		t.Errorf("expected ErrRecordingTooLong, got %v", err)
	}
}

func TestVoiceUpload_MaxRecordingsReached(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.count = maxRecordingsPerUser

	_, err := svc.Upload(context.Background(), 1, 1, "audio/webm", []byte{1})
	if !errors.Is(err, ErrMaxRecordingsReached) {
		t.Errorf("expected ErrMaxRecordingsReached, got %v", err)
	}
}

func TestVoiceUpload_CreateFailsRollsBackFile(t *testing.T) {
	svc, apiaryMock, voiceMock, dir := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.createErr = errors.New("db error")

	_, err := svc.Upload(context.Background(), 1, 1, "audio/webm", []byte{1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("read storage dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("expected audio file to be removed, found %d entries", len(entries))
	}
}
