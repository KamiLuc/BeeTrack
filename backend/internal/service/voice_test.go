package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/datatypes"
)

type mockVoiceRepo struct {
	created   *model.VoiceRecording
	count     int64
	createErr error

	recording        *model.VoiceRecording
	getRecordingErr  error
	actions          []*model.VoiceAction
	listActionsErr   error
	updatedActions   []*model.VoiceAction
	updateActionErr  error
	updatedRecording *model.VoiceRecording
	updateRecErr     error
	deletedRecID     int64
	deleteCalled     bool
	deleteActionsErr error
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

func (m *mockVoiceRepo) GetRecordingByID(ctx context.Context, id int64) (*model.VoiceRecording, error) {
	if m.getRecordingErr != nil {
		return nil, m.getRecordingErr
	}
	return m.recording, nil
}

func (m *mockVoiceRepo) UpdateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	if m.updateRecErr != nil {
		return m.updateRecErr
	}
	m.updatedRecording = rec
	return nil
}

func (m *mockVoiceRepo) ListActionsByRecordingID(ctx context.Context, recordingID int64) ([]*model.VoiceAction, error) {
	if m.listActionsErr != nil {
		return nil, m.listActionsErr
	}
	return m.actions, nil
}

func (m *mockVoiceRepo) UpdateAction(ctx context.Context, action *model.VoiceAction) error {
	if m.updateActionErr != nil {
		return m.updateActionErr
	}
	m.updatedActions = append(m.updatedActions, action)
	return nil
}

func (m *mockVoiceRepo) DeleteActionsByRecordingID(ctx context.Context, recordingID int64) error {
	m.deleteCalled = true
	m.deletedRecID = recordingID
	if m.deleteActionsErr != nil {
		return m.deleteActionsErr
	}
	return nil
}

type mockInspectionCreator struct {
	insp        *model.Inspection
	createErr   error
	diseaseErr  error
	createCalls int
	diseases    []string
}

func (m *mockInspectionCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params InspectionParams) (*model.Inspection, error) {
	m.createCalls++
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.insp != nil {
		return m.insp, nil
	}
	return &model.Inspection{ID: 42}, nil
}

func (m *mockInspectionCreator) AddDisease(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64, disease, notes string) (*model.InspectionDisease, error) {
	m.diseases = append(m.diseases, disease)
	if m.diseaseErr != nil {
		return nil, m.diseaseErr
	}
	return &model.InspectionDisease{ID: 1, Disease: disease}, nil
}

type mockTreatmentCreator struct {
	treatment   *model.Treatment
	createErr   error
	createCalls int
}

func (m *mockTreatmentCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params TreatmentParams) (*model.Treatment, error) {
	m.createCalls++
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.treatment != nil {
		return m.treatment, nil
	}
	return &model.Treatment{ID: 43}, nil
}

type mockHarvestCreator struct {
	harvest     *model.Harvest
	createErr   error
	createCalls int
}

func (m *mockHarvestCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params HarvestParams) (*model.Harvest, error) {
	m.createCalls++
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.harvest != nil {
		return m.harvest, nil
	}
	return &model.Harvest{ID: 44}, nil
}

type mockFeedingCreator struct {
	feeding     *model.Feeding
	createErr   error
	createCalls int
}

func (m *mockFeedingCreator) Create(ctx context.Context, userID, apiaryID, hiveID int64, params FeedingParams) (*model.Feeding, error) {
	m.createCalls++
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.feeding != nil {
		return m.feeding, nil
	}
	return &model.Feeding{ID: 45}, nil
}

type mockVoiceHiveReader struct {
	hive         *model.Hive
	hiveErr      error
	diseases     []*model.HiveDisease
	diseasesErr  error
}

func (m *mockVoiceHiveReader) Get(ctx context.Context, userID, apiaryID, hiveID int64) (*model.Hive, error) {
	if m.hiveErr != nil {
		return nil, m.hiveErr
	}
	if m.hive != nil {
		return m.hive, nil
	}
	return &model.Hive{ID: hiveID}, nil
}

func (m *mockVoiceHiveReader) DiseasesByHive(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error) {
	if m.diseasesErr != nil {
		return nil, m.diseasesErr
	}
	return m.diseases, nil
}

func mustMarshalJSON(t *testing.T, v any) datatypes.JSON {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal tool arguments: %v", err)
	}
	return datatypes.JSON(b)
}

func strPtr(s string) *string   { return &s }
func int64Ptr(i int64) *int64   { return &i }

func newTestVoiceService(t *testing.T) (*VoiceService, *mockApiaryMembershipReader, *mockVoiceRepo, string) {
	t.Helper()
	dir := t.TempDir()
	apiaryMock := &mockApiaryMembershipReader{}
	voiceMock := &mockVoiceRepo{}
	svc := NewVoiceService(apiaryMock, voiceMock, nil, nil, nil, nil, nil, dir)
	return svc, apiaryMock, voiceMock, dir
}

type voiceAcceptDeps struct {
	apiary      *mockApiaryMembershipReader
	voice       *mockVoiceRepo
	inspections *mockInspectionCreator
	treatments  *mockTreatmentCreator
	harvests    *mockHarvestCreator
	feedings    *mockFeedingCreator
	hives       *mockVoiceHiveReader
}

func newTestVoiceAcceptService(t *testing.T) (*VoiceService, *voiceAcceptDeps) {
	t.Helper()
	deps := &voiceAcceptDeps{
		apiary:      &mockApiaryMembershipReader{},
		voice:       &mockVoiceRepo{},
		inspections: &mockInspectionCreator{},
		treatments:  &mockTreatmentCreator{},
		harvests:    &mockHarvestCreator{},
		feedings:    &mockFeedingCreator{},
		hives:       &mockVoiceHiveReader{},
	}
	svc := NewVoiceService(deps.apiary, deps.voice, deps.inspections, deps.treatments, deps.harvests, deps.feedings, deps.hives, t.TempDir())
	return svc, deps
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

func completedRecording() *model.VoiceRecording {
	return &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusCompleted}
}

func TestVoiceAccept_CreateInspection(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:               10,
			VoiceRecordingID: 1,
			HiveID:           int64Ptr(5),
			ToolName:         strPtr(model.VoiceActionToolCreateInspection),
			Status:           model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{
				"queen_status": "seen",
				"diseases":     []string{"varroa"},
			}),
		},
	}
	deps.inspections.insp = &model.Inspection{ID: 99}

	rec, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusAccepted {
		t.Errorf("expected recording accepted, got %s", rec.Status)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	a := actions[0]
	if a.Status != model.VoiceActionStatusApplied {
		t.Errorf("expected applied, got %s (%v)", a.Status, a.ErrorMessage)
	}
	if a.ResultType == nil || *a.ResultType != model.VoiceActionResultTypeInspection {
		t.Errorf("expected result type inspection, got %v", a.ResultType)
	}
	if a.ResultRecordID == nil || *a.ResultRecordID != 99 {
		t.Errorf("expected result record id 99, got %v", a.ResultRecordID)
	}
	if deps.inspections.createCalls != 1 {
		t.Errorf("expected 1 Create call, got %d", deps.inspections.createCalls)
	}
	if len(deps.inspections.diseases) != 1 || deps.inspections.diseases[0] != "varroa" {
		t.Errorf("expected AddDisease called with varroa, got %v", deps.inspections.diseases)
	}
	if deps.voice.updatedRecording == nil || deps.voice.updatedRecording.Status != model.VoiceRecordingStatusAccepted {
		t.Error("expected UpdateRecording to persist accepted status")
	}
	if len(deps.voice.updatedActions) != 1 {
		t.Errorf("expected UpdateAction called once, got %d", len(deps.voice.updatedActions))
	}
}

func TestVoiceAccept_CreateTreatmentHarvestFeeding(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            11,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolCreateTreatment),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"medicine_name": "oxalic acid"}),
		},
		{
			ID:            12,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolCreateHarvest),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"frames": 2, "kilograms": 4.5}),
		},
		{
			ID:            13,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolCreateFeeding),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"feed_type": "syrup", "amount": "1L"}),
		},
	}

	rec, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusAccepted {
		t.Errorf("expected recording accepted, got %s", rec.Status)
	}
	for _, a := range actions {
		if a.Status != model.VoiceActionStatusApplied {
			t.Errorf("action %d expected applied, got %s (%v)", a.ID, a.Status, a.ErrorMessage)
		}
	}
	if deps.treatments.createCalls != 1 || deps.harvests.createCalls != 1 || deps.feedings.createCalls != 1 {
		t.Errorf("expected each creator called once, got treatments=%d harvests=%d feedings=%d",
			deps.treatments.createCalls, deps.harvests.createCalls, deps.feedings.createCalls)
	}
}

func TestVoiceAccept_ActionFailureDoesNotBlockOthers(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            20,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolCreateInspection),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"queen_status": "seen"}),
		},
		{
			ID:            21,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolCreateTreatment),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"medicine_name": "oxalic acid"}),
		},
	}
	deps.inspections.createErr = errors.New("hive not found")

	rec, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusAccepted {
		t.Errorf("expected recording still accepted, got %s", rec.Status)
	}
	if actions[0].Status != model.VoiceActionStatusError {
		t.Errorf("expected first action to be error, got %s", actions[0].Status)
	}
	if actions[0].ErrorMessage == nil || *actions[0].ErrorMessage == "" {
		t.Error("expected error message to be set")
	}
	if actions[1].Status != model.VoiceActionStatusApplied {
		t.Errorf("expected second action to still be applied, got %s", actions[1].Status)
	}
}

func TestVoiceAccept_SkipsAlreadyProcessedAction(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	existingErr := "hive not resolved"
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:           30,
			HiveID:       nil,
			Status:       model.VoiceActionStatusError,
			ErrorMessage: &existingErr,
		},
	}

	rec, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusAccepted {
		t.Errorf("expected recording accepted, got %s", rec.Status)
	}
	if actions[0].Status != model.VoiceActionStatusError || actions[0].ErrorMessage == nil || *actions[0].ErrorMessage != existingErr {
		t.Errorf("expected already-error action to be left untouched, got %s / %v", actions[0].Status, actions[0].ErrorMessage)
	}
	if deps.inspections.createCalls != 0 {
		t.Errorf("expected no create calls for skipped action, got %d", deps.inspections.createCalls)
	}
	if len(deps.voice.updatedActions) != 0 {
		t.Errorf("expected UpdateAction not called for skipped action, got %d calls", len(deps.voice.updatedActions))
	}
}

func TestVoiceAccept_RecordingNotCompleted(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending}

	_, _, err := svc.Accept(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotCompleted) {
		t.Errorf("expected ErrRecordingNotCompleted, got %v", err)
	}
}

func TestVoiceAccept_RecordingWrongApiary(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = &model.VoiceRecording{ID: 1, ApiaryID: 2, Status: model.VoiceRecordingStatusCompleted}

	_, _, err := svc.Accept(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotFound) {
		t.Errorf("expected ErrRecordingNotFound, got %v", err)
	}
}

func TestVoiceAccept_RecordingNotFound(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = nil

	_, _, err := svc.Accept(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotFound) {
		t.Errorf("expected ErrRecordingNotFound, got %v", err)
	}
}

func TestVoiceAccept_ApiaryNotFound(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.voice.recording = completedRecording()

	_, _, err := svc.Accept(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestVoiceReject_Success(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()

	rec, err := svc.Reject(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusRejected {
		t.Errorf("expected rejected, got %s", rec.Status)
	}
	if !deps.voice.deleteCalled || deps.voice.deletedRecID != 1 {
		t.Error("expected DeleteActionsByRecordingID to be called with recording id 1")
	}
	if deps.voice.updatedRecording == nil || deps.voice.updatedRecording.Status != model.VoiceRecordingStatusRejected {
		t.Error("expected UpdateRecording to persist rejected status")
	}
}

func TestVoiceReject_RecordingNotCompleted(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusProcessing}

	_, err := svc.Reject(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotCompleted) {
		t.Errorf("expected ErrRecordingNotCompleted, got %v", err)
	}
}

func TestVoiceReject_RecordingNotFound(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = nil

	_, err := svc.Reject(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotFound) {
		t.Errorf("expected ErrRecordingNotFound, got %v", err)
	}
}

func TestVoiceReject_ApiaryNotFound(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.voice.recording = completedRecording()

	_, err := svc.Reject(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}
