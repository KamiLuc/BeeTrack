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
	count           int64
	createErr       error
	recording       *model.VoiceRecording
	actions         []*model.VoiceAction
	deleteCalled    bool
	recordings      []*model.VoiceRecording
	recordingsTotal int64
	actionsByIDs    []*model.VoiceAction
	lastListLimit   int
	lastListOffset  int
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

func (f *fakeVoiceRepo) GetRecordingByID(ctx context.Context, id int64) (*model.VoiceRecording, error) {
	return f.recording, nil
}

func (f *fakeVoiceRepo) UpdateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	return nil
}

func (f *fakeVoiceRepo) ListRecordingsByApiaryID(ctx context.Context, apiaryID int64, limit, offset int) ([]*model.VoiceRecording, error) {
	f.lastListLimit = limit
	f.lastListOffset = offset
	return f.recordings, nil
}

func (f *fakeVoiceRepo) CountRecordingsByApiaryID(ctx context.Context, apiaryID int64) (int64, error) {
	return f.recordingsTotal, nil
}

func (f *fakeVoiceRepo) ListActionsByRecordingIDs(ctx context.Context, recordingIDs []int64) ([]*model.VoiceAction, error) {
	return f.actionsByIDs, nil
}

func (f *fakeVoiceRepo) ListActionsByRecordingID(ctx context.Context, recordingID int64) ([]*model.VoiceAction, error) {
	return f.actions, nil
}

func (f *fakeVoiceRepo) UpdateAction(ctx context.Context, action *model.VoiceAction) error {
	return nil
}

func (f *fakeVoiceRepo) DeleteActionsByRecordingID(ctx context.Context, recordingID int64) error {
	f.deleteCalled = true
	return nil
}

// fakeInspectionCreator is a minimal service.InspectionCreator for handler tests.
type fakeInspectionCreator struct{ createErr error }

func (f *fakeInspectionCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params service.InspectionParams) (*model.Inspection, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &model.Inspection{ID: 1}, nil
}

func (f *fakeInspectionCreator) AddDisease(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64, disease, notes string) (*model.InspectionDisease, error) {
	return &model.InspectionDisease{ID: 1}, nil
}

// fakeTreatmentCreator is a minimal service.TreatmentCreator for handler tests.
type fakeTreatmentCreator struct{}

func (f *fakeTreatmentCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params service.TreatmentParams) (*model.Treatment, error) {
	return &model.Treatment{ID: 1}, nil
}

// fakeHarvestCreator is a minimal service.HarvestCreator for handler tests.
type fakeHarvestCreator struct{}

func (f *fakeHarvestCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params service.HarvestParams) (*model.Harvest, error) {
	return &model.Harvest{ID: 1}, nil
}

// fakeFeedingCreator is a minimal service.FeedingCreator for handler tests.
type fakeFeedingCreator struct{}

func (f *fakeFeedingCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params service.FeedingParams) (*model.Feeding, error) {
	return &model.Feeding{ID: 1}, nil
}

// fakeVoiceHiveReader is a minimal service.VoiceHiveService for handler tests.
type fakeVoiceHiveReader struct{}

func (f *fakeVoiceHiveReader) Get(ctx context.Context, userID, apiaryID, hiveID int64) (*model.Hive, error) {
	return &model.Hive{ID: hiveID}, nil
}

func (f *fakeVoiceHiveReader) Update(ctx context.Context, userID, apiaryID, hiveID int64, name, hiveType string, active, readyForHarvest, queenNeedsReplacement, needsFood, boxNeedsAdding bool) (*model.Hive, error) {
	return &model.Hive{ID: hiveID, Name: name, Type: hiveType, Active: active, ReadyForHarvest: readyForHarvest, QueenNeedsReplacement: queenNeedsReplacement, NeedsFood: needsFood, BoxNeedsAdding: boxNeedsAdding}, nil
}

func (f *fakeVoiceHiveReader) DiseasesByHive(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error) {
	return nil, nil
}

func (f *fakeVoiceHiveReader) AddDisease(ctx context.Context, userID, apiaryID, hiveID int64, disease string) (*model.HiveDisease, error) {
	return &model.HiveDisease{ID: 1, HiveID: hiveID, Disease: disease}, nil
}

func (f *fakeVoiceHiveReader) RemoveDisease(ctx context.Context, userID, apiaryID, hiveID, diseaseID int64) error {
	return nil
}

func newAcceptRejectHandler(t *testing.T, apiary *model.Apiary, repo *fakeVoiceRepo) *VoiceHandler {
	t.Helper()
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: apiary},
		repo,
		&fakeInspectionCreator{},
		&fakeTreatmentCreator{},
		&fakeHarvestCreator{},
		&fakeFeedingCreator{},
		&fakeVoiceHiveReader{},
		t.TempDir(),
	)
	return NewVoiceHandler(svc)
}

func newAcceptRejectRequest(method, apiaryID, recordingID string) *http.Request {
	req := httptest.NewRequest(method, "/api/v1/apiaries/"+apiaryID+"/voice-recordings/"+recordingID+"/accept", nil)
	req.SetPathValue("id", apiaryID)
	req.SetPathValue("recordingId", recordingID)
	return req
}

func newListHandler(t *testing.T, apiary *model.Apiary, repo *fakeVoiceRepo) *VoiceHandler {
	t.Helper()
	svc := service.NewVoiceService(
		&fakeApiaryMembershipReader{apiary: apiary},
		repo,
		&fakeInspectionCreator{},
		&fakeTreatmentCreator{},
		&fakeHarvestCreator{},
		&fakeFeedingCreator{},
		&fakeVoiceHiveReader{},
		t.TempDir(),
	)
	return NewVoiceHandler(svc)
}

func newListRequest(apiaryID, query string) *http.Request {
	url := "/api/v1/apiaries/" + apiaryID + "/voice-recordings"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.SetPathValue("id", apiaryID)
	return req
}

func TestVoiceList_Handler_Success(t *testing.T) {
	transcript := "some transcript"
	repo := &fakeVoiceRepo{
		recordings: []*model.VoiceRecording{
			{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusCompleted, Transcript: &transcript},
		},
		recordingsTotal: 1,
		actionsByIDs: []*model.VoiceAction{
			{ID: 10, VoiceRecordingID: 1, Sequence: 1, Status: model.VoiceActionStatusProposed},
		},
	}
	h := newListHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.List))

	req := authedRequest(t, newListRequest("1", ""), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Items []map[string]any `json:"items"`
		Total int64            `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Total != 1 {
		t.Errorf("expected total 1, got %d", body.Total)
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(body.Items))
	}
	item := body.Items[0]
	if item["recording_id"] != float64(1) {
		t.Errorf("expected recording_id 1, got %v", item["recording_id"])
	}
	if item["transcript"] != transcript {
		t.Errorf("expected transcript %q, got %v", transcript, item["transcript"])
	}
	actions, ok := item["voice_actions"].([]any)
	if !ok || len(actions) != 1 {
		t.Fatalf("expected 1 voice_action, got %v", item["voice_actions"])
	}
	if repo.lastListLimit != 20 || repo.lastListOffset != 0 {
		t.Errorf("expected default limit=20 offset=0, got limit=%d offset=%d", repo.lastListLimit, repo.lastListOffset)
	}
}

func TestVoiceList_Handler_CustomLimitOffset(t *testing.T) {
	repo := &fakeVoiceRepo{}
	h := newListHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.List))

	req := authedRequest(t, newListRequest("1", "limit=5&offset=10"), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.lastListLimit != 5 || repo.lastListOffset != 10 {
		t.Errorf("expected limit=5 offset=10, got limit=%d offset=%d", repo.lastListLimit, repo.lastListOffset)
	}
}

func TestVoiceList_Handler_InvalidLimitOffsetFallsBackToDefault(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"non-numeric limit", "limit=abc"},
		{"negative limit", "limit=-1"},
		{"zero limit", "limit=0"},
		{"negative offset", "offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeVoiceRepo{}
			h := newListHandler(t, &model.Apiary{ID: 1}, repo)
			handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.List))

			req := authedRequest(t, newListRequest("1", tt.query), 1)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if repo.lastListLimit != 20 || repo.lastListOffset != 0 {
				t.Errorf("expected fallback to default limit=20 offset=0, got limit=%d offset=%d", repo.lastListLimit, repo.lastListOffset)
			}
		})
	}
}

func TestVoiceList_Handler_ApiaryNotFound(t *testing.T) {
	repo := &fakeVoiceRepo{}
	h := newListHandler(t, nil, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.List))

	req := authedRequest(t, newListRequest("1", ""), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "APIARY_NOT_FOUND" {
		t.Errorf("expected code APIARY_NOT_FOUND, got %q", code)
	}
}

func TestVoiceList_Handler_Empty(t *testing.T) {
	repo := &fakeVoiceRepo{recordings: nil, recordingsTotal: 0}
	h := newListHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.List))

	req := authedRequest(t, newListRequest("1", ""), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Items []map[string]any `json:"items"`
		Total int64            `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Items == nil {
		t.Error("expected items to be an empty array, not null")
	}
	if len(body.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(body.Items))
	}
	if body.Total != 0 {
		t.Errorf("expected total 0, got %d", body.Total)
	}
}

func newCancelRequest(apiaryID, recordingID string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/apiaries/"+apiaryID+"/voice-recordings/"+recordingID, nil)
	req.SetPathValue("id", apiaryID)
	req.SetPathValue("recordingId", recordingID)
	return req
}

func TestVoiceCancel_Handler_Success(t *testing.T) {
	audioPath := "audio.webm"
	repo := &fakeVoiceRepo{
		recording: &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending, AudioPath: &audioPath},
	}
	h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Cancel))

	req := authedRequest(t, newCancelRequest("1", "1"), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", rec.Body.String())
	}
}

func TestVoiceCancel_Handler_NotCancelable(t *testing.T) {
	repo := &fakeVoiceRepo{
		recording: &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusCompleted},
	}
	h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Cancel))

	req := authedRequest(t, newCancelRequest("1", "1"), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "RECORDING_NOT_CANCELABLE" {
		t.Errorf("expected code RECORDING_NOT_CANCELABLE, got %q", code)
	}
}

func TestVoiceCancel_Handler_RecordingNotFound(t *testing.T) {
	repo := &fakeVoiceRepo{recording: nil}
	h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Cancel))

	req := authedRequest(t, newCancelRequest("1", "1"), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "RECORDING_NOT_FOUND" {
		t.Errorf("expected code RECORDING_NOT_FOUND, got %q", code)
	}
}

func TestVoiceCancel_Handler_MissingAuth(t *testing.T) {
	repo := &fakeVoiceRepo{
		recording: &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending},
	}
	h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Cancel))

	req := newCancelRequest("1", "1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVoiceCancel_Handler_InvalidPathIDs(t *testing.T) {
	tests := []struct {
		name        string
		apiaryID    string
		recordingID string
	}{
		{"invalid apiary id", "not-a-number", "1"},
		{"invalid recording id", "1", "not-a-number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeVoiceRepo{
				recording: &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending},
			}
			h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
			handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Cancel))

			req := authedRequest(t, newCancelRequest(tt.apiaryID, tt.recordingID), 1)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestVoiceAccept_Handler_Success(t *testing.T) {
	repo := &fakeVoiceRepo{
		recording: &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusCompleted},
		actions: []*model.VoiceAction{
			{
				ID:            1,
				HiveID:        int64Ptr(5),
				ToolName:      strPtr(model.VoiceActionToolCreateInspection),
				Status:        model.VoiceActionStatusProposed,
				ToolArguments: []byte(`{"queen_status":"seen"}`),
			},
		},
	}
	h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Accept))

	req := authedRequest(t, newAcceptRejectRequest(http.MethodPost, "1", "1"), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		RecordingID int64            `json:"recording_id"`
		Status      string           `json:"status"`
		Actions     []map[string]any `json:"actions"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.RecordingID != 1 {
		t.Errorf("expected recording_id 1, got %d", body.RecordingID)
	}
	if body.Status != model.VoiceRecordingStatusAccepted {
		t.Errorf("expected status %q, got %q", model.VoiceRecordingStatusAccepted, body.Status)
	}
	if len(body.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(body.Actions))
	}
	if body.Actions[0]["status"] != model.VoiceActionStatusApplied {
		t.Errorf("expected action applied, got %v", body.Actions[0]["status"])
	}
}

func TestVoiceReject_Handler_Success(t *testing.T) {
	repo := &fakeVoiceRepo{
		recording: &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusCompleted},
	}
	h := newAcceptRejectHandler(t, &model.Apiary{ID: 1}, repo)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Reject))

	req := authedRequest(t, newAcceptRejectRequest(http.MethodPost, "1", "1"), 1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	recordingID, status := decodeVoiceUploadResponse(t, rec)
	if recordingID != 1 {
		t.Errorf("expected recording_id 1, got %d", recordingID)
	}
	if status != model.VoiceRecordingStatusRejected {
		t.Errorf("expected status %q, got %q", model.VoiceRecordingStatusRejected, status)
	}
	if !repo.deleteCalled {
		t.Error("expected DeleteActionsByRecordingID to be called")
	}
}

func TestVoiceAcceptReject_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		apiary     *model.Apiary
		recording  *model.VoiceRecording
		wantStatus int
		wantCode   string
	}{
		{
			name:       "recording not found",
			apiary:     &model.Apiary{ID: 1},
			recording:  nil,
			wantStatus: http.StatusNotFound,
			wantCode:   "RECORDING_NOT_FOUND",
		},
		{
			name:       "recording not completed",
			apiary:     &model.Apiary{ID: 1},
			recording:  &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending},
			wantStatus: http.StatusConflict,
			wantCode:   "RECORDING_NOT_COMPLETED",
		},
		{
			name:       "apiary not found",
			apiary:     nil,
			recording:  &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusCompleted},
			wantStatus: http.StatusNotFound,
			wantCode:   "APIARY_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/accept", func(t *testing.T) {
			repo := &fakeVoiceRepo{recording: tt.recording}
			h := newAcceptRejectHandler(t, tt.apiary, repo)
			handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Accept))

			req := authedRequest(t, newAcceptRejectRequest(http.MethodPost, "1", "1"), 1)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if code := decodeErrorCode(t, rec); code != tt.wantCode {
				t.Errorf("expected code %q, got %q", tt.wantCode, code)
			}
		})

		t.Run(tt.name+"/reject", func(t *testing.T) {
			repo := &fakeVoiceRepo{recording: tt.recording}
			h := newAcceptRejectHandler(t, tt.apiary, repo)
			handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Reject))

			req := authedRequest(t, newAcceptRejectRequest(http.MethodPost, "1", "1"), 1)
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

func strPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64 { return &i }

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
		nil, nil, nil, nil, nil,
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
		nil, nil, nil, nil, nil,
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
		nil, nil, nil, nil, nil,
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
		nil, nil, nil, nil, nil,
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
			count:      10,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "MAX_RECORDINGS_REACHED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewVoiceService(
				&fakeApiaryMembershipReader{apiary: tt.apiary},
				&fakeVoiceRepo{count: tt.count},
				nil, nil, nil, nil, nil,
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
		nil, nil, nil, nil, nil,
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
