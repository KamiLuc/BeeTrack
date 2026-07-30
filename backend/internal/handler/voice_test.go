package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
)

// fakeVoiceRepo is a minimal service.VoiceRepository for handler tests.
type fakeVoiceRepo struct {
	count     int64
	createErr error
}

func (f *fakeVoiceRepo) CreateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	if f.createErr != nil {
		return f.createErr
	}
	rec.ID = 1
	return nil
}

func (f *fakeVoiceRepo) CountRecordingsByUserID(ctx context.Context, userID int64) (int64, error) {
	return f.count, nil
}

func newVoiceUploadRequest(t *testing.T, apiaryID, contentType string, size int) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="audio"; filename="test.webm"`)
	header.Set("Content-Type", contentType)
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(make([]byte, size))); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/apiaries/"+apiaryID+"/voice", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetPathValue("id", apiaryID)
	return req
}

func decodeVoiceUploadResponse(t *testing.T, rec *httptest.ResponseRecorder) (int64, string) {
	t.Helper()
	var body struct {
		RecordingID int64  `json:"recording_id"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body.RecordingID, body.Status
}

func TestVoiceUpload_Success_Handler(t *testing.T) {
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeVoiceRepo{},
		t.TempDir(),
	)
	h := NewVoiceHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newVoiceUploadRequest(t, "1", "audio/webm", 1024)
	req = authedRequest(t, req, 1)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}

	recordingID, status := decodeVoiceUploadResponse(t, rec)
	if recordingID != 1 {
		t.Errorf("expected recording_id 1, got %d", recordingID)
	}
	if status != model.VoiceRecordingStatusPending {
		t.Errorf("expected status %q, got %q", model.VoiceRecordingStatusPending, status)
	}
}

func TestVoiceUpload_MissingAuth(t *testing.T) {
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeVoiceRepo{},
		t.TempDir(),
	)
	h := NewVoiceHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newVoiceUploadRequest(t, "1", "audio/webm", 1024)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVoiceUpload_InvalidApiaryID(t *testing.T) {
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeVoiceRepo{},
		t.TempDir(),
	)
	h := NewVoiceHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newVoiceUploadRequest(t, "not-a-number", "audio/webm", 1024)
	req = authedRequest(t, req, 1)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVoiceUpload_MissingAudioField(t *testing.T) {
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeVoiceRepo{},
		t.TempDir(),
	)
	h := NewVoiceHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if err := w.WriteField("note", "no audio here"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/apiaries/1/voice", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetPathValue("id", "1")
	req = authedRequest(t, req, 1)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "MISSING_FILE" {
		t.Errorf("expected code MISSING_FILE, got %q", code)
	}
}

func TestVoiceUpload_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		apiary     *model.Apiary
		mimeType   string
		size       int
		count      int64
		wantStatus int
		wantCode   string
	}{
		{
			name:       "apiary not found",
			apiary:     nil,
			mimeType:   "audio/webm",
			size:       1024,
			wantStatus: http.StatusNotFound,
			wantCode:   "APIARY_NOT_FOUND",
		},
		{
			name:       "invalid audio type",
			apiary:     &model.Apiary{ID: 1},
			mimeType:   "audio/mpeg",
			size:       1024,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_AUDIO_TYPE",
		},
		{
			name:       "max recordings reached",
			apiary:     &model.Apiary{ID: 1},
			mimeType:   "audio/webm",
			size:       1024,
			count:      20,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "MAX_RECORDINGS_REACHED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewVoiceService(
				&fakeApiaryMembershipReader{apiary: tt.apiary},
				&fakeVoiceRepo{count: tt.count},
				t.TempDir(),
			)
			h := NewVoiceHandler(svc)
			handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

			req := newVoiceUploadRequest(t, "1", tt.mimeType, tt.size)
			req = authedRequest(t, req, 1)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if code := decodeErrorCode(t, rec); code != tt.wantCode {
				t.Errorf("expected code %q, got %q", tt.wantCode, code)
			}
		})
	}
}

func TestVoiceUpload_RecordingTooLarge(t *testing.T) {
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeVoiceRepo{},
		t.TempDir(),
	)
	h := NewVoiceHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newVoiceUploadRequest(t, "1", "audio/webm", service.MaxAudioBytes+1024*1024)
	req = authedRequest(t, req, 1)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "RECORDING_TOO_LONG" {
		t.Errorf("expected code RECORDING_TOO_LONG, got %q", code)
	}
}
