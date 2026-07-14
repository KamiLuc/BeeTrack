package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockTreatmentRepo struct {
	treatment   *model.Treatment
	treatments  []*model.Treatment
	bulkCreated []*model.Treatment
	created     *model.Treatment
	updated     *model.Treatment
	deletedID   int64
}

func (m *mockTreatmentRepo) BulkCreate(ctx context.Context, treatments []*model.Treatment) error {
	for i, t := range treatments {
		t.ID = int64(i + 1)
	}
	m.bulkCreated = treatments
	return nil
}

func (m *mockTreatmentRepo) Create(ctx context.Context, t *model.Treatment) error {
	t.ID = 1
	m.created = t
	return nil
}

func (m *mockTreatmentRepo) Delete(ctx context.Context, treatmentID int64) error {
	m.deletedID = treatmentID
	return nil
}

func (m *mockTreatmentRepo) GetByID(ctx context.Context, treatmentID, hiveID int64) (*model.Treatment, error) {
	if m.treatment == nil || m.treatment.ID != treatmentID || m.treatment.HiveID != hiveID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.treatment, nil
}

func (m *mockTreatmentRepo) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	return int64(len(m.treatments)), nil
}

func (m *mockTreatmentRepo) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Treatment, error) {
	return m.treatments, nil
}

func (m *mockTreatmentRepo) Update(ctx context.Context, t *model.Treatment) error {
	m.updated = t
	return nil
}

type mockBulkHiveReader struct {
	hives []*model.Hive
}

func (m *mockBulkHiveReader) ListByApiaryID(_ context.Context, _ int64) ([]*model.Hive, error) {
	return m.hives, nil
}

func newTreatmentSvc(repo *mockTreatmentRepo) *TreatmentService {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	return NewTreatmentService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: hive},
		&mockBulkHiveReader{hives: []*model.Hive{hive}},
		repo,
	)
}

func validTreatmentParams() TreatmentParams {
	return TreatmentParams{
		TreatedAt:    time.Now(),
		MedicineName: "Apiwarol",
		Dose:         "2",
		Notes:        "applied evenly",
	}
}

func TestTreatmentCreate_MedicineNameTooLong(t *testing.T) {
	svc := newTreatmentSvc(&mockTreatmentRepo{})

	params := validTreatmentParams()
	params.MedicineName = strings.Repeat("a", 51)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrMedicineNameTooLong) {
		t.Errorf("expected ErrMedicineNameTooLong, got %v", err)
	}
}

func TestTreatmentCreate_DoseTooLong(t *testing.T) {
	svc := newTreatmentSvc(&mockTreatmentRepo{})

	params := validTreatmentParams()
	params.Dose = strings.Repeat("1", 21)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrDoseTooLong) {
		t.Errorf("expected ErrDoseTooLong, got %v", err)
	}
}

func TestTreatmentCreate_NotesTooLong(t *testing.T) {
	svc := newTreatmentSvc(&mockTreatmentRepo{})

	params := validTreatmentParams()
	params.Notes = strings.Repeat("a", 5001)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrTreatmentNotesTooLong) {
		t.Errorf("expected ErrTreatmentNotesTooLong, got %v", err)
	}
}

func TestTreatmentCreate(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := newTreatmentSvc(repo)

	tr, err := svc.Create(context.Background(), 1, 1, 10, validTreatmentParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.MedicineName != "Apiwarol" {
		t.Errorf("expected medicine_name Apiwarol, got %s", tr.MedicineName)
	}
	if tr.Dose != "2" {
		t.Errorf("expected dose '2', got %s", tr.Dose)
	}
}

func TestTreatmentCreate_DefaultDose(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := newTreatmentSvc(repo)

	params := validTreatmentParams()
	params.Dose = ""
	tr, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Dose != "1" {
		t.Errorf("expected default dose '1', got %s", tr.Dose)
	}
}

func TestTreatmentCreate_MissingTreatedAt(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := newTreatmentSvc(repo)

	params := validTreatmentParams()
	params.TreatedAt = time.Time{}
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != ErrTreatedAtRequired {
		t.Errorf("expected ErrTreatedAtRequired, got %v", err)
	}
}

func TestTreatmentCreate_MissingMedicineName(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := newTreatmentSvc(repo)

	params := validTreatmentParams()
	params.MedicineName = ""
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != ErrMedicineNameRequired {
		t.Errorf("expected ErrMedicineNameRequired, got %v", err)
	}
}

func TestTreatmentGet_NotFound(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := newTreatmentSvc(repo)

	_, err := svc.Get(context.Background(), 1, 1, 10, 99)
	if err != ErrTreatmentNotFound {
		t.Errorf("expected ErrTreatmentNotFound, got %v", err)
	}
}

func TestTreatmentDelete(t *testing.T) {
	repo := &mockTreatmentRepo{
		treatment: &model.Treatment{ID: 5, HiveID: 10},
	}
	svc := newTreatmentSvc(repo)

	if err := svc.Delete(context.Background(), 1, 1, 10, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedID != 5 {
		t.Errorf("expected deletedID 5, got %d", repo.deletedID)
	}
}

func TestBulkTreat(t *testing.T) {
	hives := []*model.Hive{
		{ID: 10, ApiaryID: 1},
		{ID: 11, ApiaryID: 1},
		{ID: 12, ApiaryID: 1},
	}
	repo := &mockTreatmentRepo{}
	svc := NewTreatmentService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: hives[0]},
		&mockBulkHiveReader{hives: hives},
		repo,
	)

	count, err := svc.BulkTreat(context.Background(), 1, 1, validTreatmentParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
	if len(repo.bulkCreated) != 3 {
		t.Errorf("expected 3 bulk created, got %d", len(repo.bulkCreated))
	}
	for _, tr := range repo.bulkCreated {
		if tr.MedicineName != "Apiwarol" {
			t.Errorf("expected medicine_name Apiwarol, got %s", tr.MedicineName)
		}
	}
}

func TestBulkTreat_NoHives(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := NewTreatmentService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{},
		&mockBulkHiveReader{hives: []*model.Hive{}},
		repo,
	)

	count, err := svc.BulkTreat(context.Background(), 1, 1, validTreatmentParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestBulkTreat_DefaultDose(t *testing.T) {
	hives := []*model.Hive{{ID: 10, ApiaryID: 1}}
	repo := &mockTreatmentRepo{}
	svc := NewTreatmentService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: hives[0]},
		&mockBulkHiveReader{hives: hives},
		repo,
	)

	params := validTreatmentParams()
	params.Dose = ""
	_, err := svc.BulkTreat(context.Background(), 1, 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.bulkCreated[0].Dose != "1" {
		t.Errorf("expected default dose '1', got %s", repo.bulkCreated[0].Dose)
	}
}

func TestBulkTreat_MissingMedicine(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := newTreatmentSvc(repo)

	params := validTreatmentParams()
	params.MedicineName = ""
	_, err := svc.BulkTreat(context.Background(), 1, 1, params)
	if err != ErrMedicineNameRequired {
		t.Errorf("expected ErrMedicineNameRequired, got %v", err)
	}
}

func TestBulkTreat_ApiaryNotFound(t *testing.T) {
	repo := &mockTreatmentRepo{}
	svc := NewTreatmentService(
		&mockApiaryRepo{},
		&mockInspectionHiveReader{},
		&mockBulkHiveReader{},
		repo,
	)

	_, err := svc.BulkTreat(context.Background(), 1, 99, validTreatmentParams())
	if err != ErrApiaryNotFound {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}
