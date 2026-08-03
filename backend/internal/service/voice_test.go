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

	recordings          []*model.VoiceRecording
	listRecordingsErr   error
	recordingsTotal     int64
	countByApiaryErr    error
	actionsByIDs        []*model.VoiceAction
	listActionsByIDsErr error
	lastListLimit       int
	lastListOffset      int
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

func (m *mockVoiceRepo) ListRecordingsByApiaryID(ctx context.Context, apiaryID int64, limit, offset int) ([]*model.VoiceRecording, error) {
	m.lastListLimit = limit
	m.lastListOffset = offset
	if m.listRecordingsErr != nil {
		return nil, m.listRecordingsErr
	}
	return m.recordings, nil
}

func (m *mockVoiceRepo) CountRecordingsByApiaryID(ctx context.Context, apiaryID int64) (int64, error) {
	if m.countByApiaryErr != nil {
		return 0, m.countByApiaryErr
	}
	return m.recordingsTotal, nil
}

func (m *mockVoiceRepo) ListActionsByRecordingIDs(ctx context.Context, recordingIDs []int64) ([]*model.VoiceAction, error) {
	if m.listActionsByIDsErr != nil {
		return nil, m.listActionsByIDsErr
	}
	return m.actionsByIDs, nil
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
	hive        *model.Hive
	hiveErr     error
	diseases    []*model.HiveDisease
	diseasesErr error

	updateErr        error
	updatedHive      *model.Hive
	addedDiseases    []string
	addDiseaseErr    error
	removedDiseases  []int64
	removeDiseaseErr error
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

func (m *mockVoiceHiveReader) Update(ctx context.Context, userID, apiaryID, hiveID int64, name, hiveType string, active, readyForHarvest, queenNeedsReplacement, needsFood, boxNeedsAdding bool) (*model.Hive, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updatedHive = &model.Hive{
		ID: hiveID, Name: name, Type: hiveType, Active: active,
		ReadyForHarvest: readyForHarvest, QueenNeedsReplacement: queenNeedsReplacement,
		NeedsFood: needsFood, BoxNeedsAdding: boxNeedsAdding,
	}
	return m.updatedHive, nil
}

func (m *mockVoiceHiveReader) DiseasesByHive(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error) {
	if m.diseasesErr != nil {
		return nil, m.diseasesErr
	}
	return m.diseases, nil
}

func (m *mockVoiceHiveReader) AddDisease(ctx context.Context, userID, apiaryID, hiveID int64, disease string) (*model.HiveDisease, error) {
	m.addedDiseases = append(m.addedDiseases, disease)
	if m.addDiseaseErr != nil {
		return nil, m.addDiseaseErr
	}
	return &model.HiveDisease{ID: int64(len(m.addedDiseases)), HiveID: hiveID, Disease: disease}, nil
}

func (m *mockVoiceHiveReader) RemoveDisease(ctx context.Context, userID, apiaryID, hiveID, diseaseID int64) error {
	m.removedDiseases = append(m.removedDiseases, diseaseID)
	return m.removeDiseaseErr
}

func mustMarshalJSON(t *testing.T, v any) datatypes.JSON {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal tool arguments: %v", err)
	}
	return datatypes.JSON(b)
}

func strPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64 { return &i }

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

func TestVoiceUploadForHive_Success(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.hives.hive = &model.Hive{ID: 9}

	data := []byte{0x1A, 0x45, 0xDF, 0xA3}
	rec, err := svc.UploadForHive(context.Background(), 1, 1, 9, "audio/webm", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusPending {
		t.Errorf("unexpected status: %s", rec.Status)
	}
	if deps.voice.created == nil {
		t.Fatal("expected CreateRecording to be called")
	}
	if deps.voice.created.HiveID == nil || *deps.voice.created.HiveID != 9 {
		t.Errorf("expected created recording HiveID=9, got %v", deps.voice.created.HiveID)
	}
	if rec.HiveID == nil || *rec.HiveID != 9 {
		t.Errorf("expected returned recording HiveID=9, got %v", rec.HiveID)
	}
}

func TestVoiceUploadForHive_HiveNotFound(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.hives.hiveErr = ErrHiveNotFound

	_, err := svc.UploadForHive(context.Background(), 1, 1, 9, "audio/webm", []byte{1})
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
	if !errors.Is(err, ErrHiveNotFound) || errors.Unwrap(err) != nil {
		t.Errorf("expected ErrHiveNotFound to be propagated unwrapped, got %v (unwrap=%v)", err, errors.Unwrap(err))
	}
}

func TestVoiceUploadForHive_ApiaryNotFound(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.hives.hiveErr = ErrApiaryNotFound

	_, err := svc.UploadForHive(context.Background(), 1, 1, 9, "audio/webm", []byte{1})
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestVoiceUploadForHive_MaxRecordingsReached(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.hives.hive = &model.Hive{ID: 9}
	deps.voice.count = maxRecordingsPerUser

	_, err := svc.UploadForHive(context.Background(), 1, 1, 9, "audio/webm", []byte{1})
	if !errors.Is(err, ErrMaxRecordingsReached) {
		t.Errorf("expected ErrMaxRecordingsReached, got %v", err)
	}
}

func TestVoiceCancel_Success(t *testing.T) {
	svc, apiaryMock, voiceMock, dir := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	if err := os.WriteFile(filepath.Join(dir, "audio.webm"), []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatalf("write audio file: %v", err)
	}
	audioPath := "audio.webm"
	voiceMock.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending, AudioPath: &audioPath}

	rec, err := svc.Cancel(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusCancelled {
		t.Errorf("expected cancelled, got %s", rec.Status)
	}
	if rec.AudioPath != nil {
		t.Errorf("expected AudioPath cleared, got %v", *rec.AudioPath)
	}
	if voiceMock.updatedRecording == nil || voiceMock.updatedRecording.Status != model.VoiceRecordingStatusCancelled {
		t.Error("expected UpdateRecording to persist cancelled status")
	}
	if _, err := os.Stat(filepath.Join(dir, "audio.webm")); !os.IsNotExist(err) {
		t.Errorf("expected audio file to be removed, stat err: %v", err)
	}
}

func TestVoiceCancel_NoAudioPath(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending, AudioPath: nil}

	rec, err := svc.Cancel(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusCancelled {
		t.Errorf("expected cancelled, got %s", rec.Status)
	}
	if rec.AudioPath != nil {
		t.Errorf("expected AudioPath to remain nil, got %v", *rec.AudioPath)
	}
}

func TestVoiceCancel_NotCancelable(t *testing.T) {
	statuses := []string{
		model.VoiceRecordingStatusProcessing,
		model.VoiceRecordingStatusCompleted,
		model.VoiceRecordingStatusAccepted,
		model.VoiceRecordingStatusRejected,
		model.VoiceRecordingStatusCancelled,
	}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
			apiaryMock.apiary = &model.Apiary{ID: 1}
			voiceMock.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: status}

			_, err := svc.Cancel(context.Background(), 1, 1, 1)
			if !errors.Is(err, ErrRecordingNotCancelable) {
				t.Errorf("expected ErrRecordingNotCancelable, got %v", err)
			}
			if voiceMock.updatedRecording != nil {
				t.Error("expected UpdateRecording not to be called")
			}
		})
	}
}

func TestVoiceCancel_RecordingWrongApiary(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recording = &model.VoiceRecording{ID: 1, ApiaryID: 2, Status: model.VoiceRecordingStatusPending}

	_, err := svc.Cancel(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotFound) {
		t.Errorf("expected ErrRecordingNotFound, got %v", err)
	}
}

func TestVoiceCancel_RecordingNotFound(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recording = nil

	_, err := svc.Cancel(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrRecordingNotFound) {
		t.Errorf("expected ErrRecordingNotFound, got %v", err)
	}
}

func TestVoiceCancel_ApiaryNotFound(t *testing.T) {
	svc, _, voiceMock, _ := newTestVoiceService(t)
	voiceMock.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending}

	_, err := svc.Cancel(context.Background(), 1, 1, 1)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestVoiceCancel_UpdateRecordingFails(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recording = &model.VoiceRecording{ID: 1, ApiaryID: 1, Status: model.VoiceRecordingStatusPending}
	voiceMock.updateRecErr = errors.New("db error")

	_, err := svc.Cancel(context.Background(), 1, 1, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
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

func TestVoiceAccept_UpdateHiveStatus_PartialFlagsMergeWithCurrent(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hive = &model.Hive{
		ID: 5, Name: "Hive 5", Type: "langstroth", Active: true,
		ReadyForHarvest: false, QueenNeedsReplacement: true, NeedsFood: true, BoxNeedsAdding: false,
	}
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            40,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"ready_for_harvest": true}),
		},
	}

	rec, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Status != model.VoiceRecordingStatusAccepted {
		t.Errorf("expected recording accepted, got %s", rec.Status)
	}
	a := actions[0]
	if a.Status != model.VoiceActionStatusApplied {
		t.Errorf("expected applied, got %s (%v)", a.Status, a.ErrorMessage)
	}
	if a.ResultType == nil || *a.ResultType != model.VoiceActionResultTypeHiveStatus {
		t.Errorf("expected result type hive_status, got %v", a.ResultType)
	}
	if a.ResultRecordID == nil || *a.ResultRecordID != 5 {
		t.Errorf("expected result record id 5, got %v", a.ResultRecordID)
	}
	updated := deps.hives.updatedHive
	if updated == nil {
		t.Fatal("expected hives.Update to be called")
	}
	if !updated.ReadyForHarvest {
		t.Error("expected ready_for_harvest to be set true")
	}
	if !updated.QueenNeedsReplacement {
		t.Error("expected queen_needs_replacement to be carried through unchanged (true)")
	}
	if !updated.NeedsFood {
		t.Error("expected needs_food to be carried through unchanged (true)")
	}
	if updated.BoxNeedsAdding {
		t.Error("expected box_needs_adding to be carried through unchanged (false)")
	}
	if updated.Name != "Hive 5" || updated.Type != "langstroth" || !updated.Active {
		t.Errorf("expected name/type/active passed through unchanged, got %+v", updated)
	}
}

func TestVoiceAccept_UpdateHiveStatus_AllFlagsSet(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hive = &model.Hive{ID: 5, Name: "Hive 5", Type: "langstroth", Active: true}
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:       41,
			HiveID:   int64Ptr(5),
			ToolName: strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:   model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{
				"ready_for_harvest":       true,
				"queen_needs_replacement": true,
				"needs_food":              true,
				"box_needs_adding":        true,
			}),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusApplied {
		t.Errorf("expected applied, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
	}
	updated := deps.hives.updatedHive
	if updated == nil {
		t.Fatal("expected hives.Update to be called")
	}
	if !updated.ReadyForHarvest || !updated.QueenNeedsReplacement || !updated.NeedsFood || !updated.BoxNeedsAdding {
		t.Errorf("expected all flags set true, got %+v", updated)
	}
}

func TestVoiceAccept_UpdateHiveStatus_SyncsDiseases(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hive = &model.Hive{ID: 5}
	deps.hives.diseases = []*model.HiveDisease{
		{ID: 100, HiveID: 5, Disease: "varroa"},
		{ID: 101, HiveID: 5, Disease: "nosema"},
	}
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            42,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"diseases": []string{"varroa", "chalkbrood"}}),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusApplied {
		t.Errorf("expected applied, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
	}
	if len(deps.hives.addedDiseases) != 1 || deps.hives.addedDiseases[0] != "chalkbrood" {
		t.Errorf("expected AddDisease called only with chalkbrood, got %v", deps.hives.addedDiseases)
	}
	if len(deps.hives.removedDiseases) != 1 || deps.hives.removedDiseases[0] != 101 {
		t.Errorf("expected RemoveDisease called only with id 101 (nosema), got %v", deps.hives.removedDiseases)
	}
}

func TestVoiceAccept_UpdateHiveStatus_NoDiseasesFieldSkipsSync(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hive = &model.Hive{ID: 5}
	deps.hives.diseases = []*model.HiveDisease{{ID: 100, HiveID: 5, Disease: "varroa"}}
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            43,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"ready_for_harvest": true}),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusApplied {
		t.Errorf("expected applied, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
	}
	if len(deps.hives.addedDiseases) != 0 {
		t.Errorf("expected no AddDisease calls, got %v", deps.hives.addedDiseases)
	}
	if len(deps.hives.removedDiseases) != 0 {
		t.Errorf("expected no RemoveDisease calls, got %v", deps.hives.removedDiseases)
	}
}

func TestVoiceAccept_UpdateHiveStatus_GetHiveFails(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hiveErr = errors.New("hive not found")
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            44,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"ready_for_harvest": true}),
		},
		{
			ID:            45,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolCreateTreatment),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"medicine_name": "oxalic acid"}),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusError || actions[0].ErrorMessage == nil {
		t.Errorf("expected first action error, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
	}
	if actions[1].Status != model.VoiceActionStatusApplied {
		t.Errorf("expected second action still applied, got %s (%v)", actions[1].Status, actions[1].ErrorMessage)
	}
}

func TestVoiceAccept_UpdateHiveStatus_UpdateHiveFails(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hive = &model.Hive{ID: 5}
	deps.hives.updateErr = errors.New("db error")
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            46,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"ready_for_harvest": true}),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusError || actions[0].ErrorMessage == nil {
		t.Errorf("expected error, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
	}
}

func TestVoiceAccept_UpdateHiveStatus_DiseaseSyncFailureStillErrorsAction(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.hives.hive = &model.Hive{ID: 5}
	deps.hives.addDiseaseErr = errors.New("db error")
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            47,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: mustMarshalJSON(t, map[string]any{"diseases": []string{"varroa"}}),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusError || actions[0].ErrorMessage == nil {
		t.Errorf("expected error, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
	}
	if deps.hives.updatedHive == nil {
		t.Error("expected hives.Update to have already persisted before the disease sync failure")
	}
}

func TestVoiceAccept_UpdateHiveStatus_MalformedJSON(t *testing.T) {
	svc, deps := newTestVoiceAcceptService(t)
	deps.apiary.apiary = &model.Apiary{ID: 1}
	deps.voice.recording = completedRecording()
	deps.voice.actions = []*model.VoiceAction{
		{
			ID:            48,
			HiveID:        int64Ptr(5),
			ToolName:      strPtr(model.VoiceActionToolUpdateHiveStatus),
			Status:        model.VoiceActionStatusProposed,
			ToolArguments: datatypes.JSON(`{"ready_for_harvest": "not-a-bool"}`),
		},
	}

	_, actions, err := svc.Accept(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actions[0].Status != model.VoiceActionStatusError || actions[0].ErrorMessage == nil {
		t.Errorf("expected error, got %s (%v)", actions[0].Status, actions[0].ErrorMessage)
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

func TestVoiceList_Success(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recordings = []*model.VoiceRecording{
		{ID: 1, ApiaryID: 1},
		{ID: 2, ApiaryID: 1},
	}
	voiceMock.recordingsTotal = 50
	voiceMock.actionsByIDs = []*model.VoiceAction{
		{ID: 100, VoiceRecordingID: 1, Sequence: 1},
		{ID: 101, VoiceRecordingID: 1, Sequence: 2},
	}

	recs, grouped, total, err := svc.List(context.Background(), 1, 1, 20, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 recordings, got %d", len(recs))
	}
	if total != 50 {
		t.Errorf("expected total 50, got %d", total)
	}
	if len(grouped) != 2 {
		t.Fatalf("expected grouped map with 2 keys, got %d", len(grouped))
	}
	if len(grouped[1]) != 2 {
		t.Errorf("expected 2 actions for recording 1, got %d", len(grouped[1]))
	}
	actionsForRec2, ok := grouped[2]
	if !ok {
		t.Fatal("expected recording 2 to have a key in the grouped map")
	}
	if actionsForRec2 == nil || len(actionsForRec2) != 0 {
		t.Errorf("expected empty non-nil slice for recording 2, got %v", actionsForRec2)
	}
}

func TestVoiceList_Empty(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recordings = nil
	voiceMock.recordingsTotal = 0

	recs, grouped, total, err := svc.List(context.Background(), 1, 1, 20, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected no recordings, got %d", len(recs))
	}
	if len(grouped) != 0 {
		t.Errorf("expected empty grouped map, got %d entries", len(grouped))
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestVoiceList_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestVoiceService(t)

	_, _, _, err := svc.List(context.Background(), 1, 1, 20, 0)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestVoiceList_CountFails(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.countByApiaryErr = errors.New("db error")

	_, _, _, err := svc.List(context.Background(), 1, 1, 20, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVoiceList_ListRecordingsFails(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.listRecordingsErr = errors.New("db error")

	_, _, _, err := svc.List(context.Background(), 1, 1, 20, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVoiceList_ListActionsFails(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	voiceMock.recordings = []*model.VoiceRecording{{ID: 1, ApiaryID: 1}}
	voiceMock.listActionsByIDsErr = errors.New("db error")

	_, _, _, err := svc.List(context.Background(), 1, 1, 20, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVoiceList_PassesLimitOffsetToRepository(t *testing.T) {
	svc, apiaryMock, voiceMock, _ := newTestVoiceService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}

	if _, _, _, err := svc.List(context.Background(), 1, 1, 5, 10); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if voiceMock.lastListLimit != 5 || voiceMock.lastListOffset != 10 {
		t.Errorf("expected limit=5 offset=10 to reach repository, got limit=%d offset=%d", voiceMock.lastListLimit, voiceMock.lastListOffset)
	}
}
