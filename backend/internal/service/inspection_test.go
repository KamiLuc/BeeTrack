package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockInspectionHiveReader struct {
	hive *model.Hive
}

func (m *mockInspectionHiveReader) GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error) {
	if m.hive == nil || m.hive.ID != hiveID || m.hive.ApiaryID != apiaryID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.hive, nil
}

type mockInspectionRepo struct {
	inspection     *model.Inspection
	inspections    []*model.Inspection
	created        *model.Inspection
	updated        *model.Inspection
	deletedID      int64
	disease        *model.InspectionDisease
	createdDisease *model.InspectionDisease
	deletedDiseaseID int64
}

func (m *mockInspectionRepo) Create(ctx context.Context, insp *model.Inspection) error {
	insp.ID = 1
	m.created = insp
	return nil
}

func (m *mockInspectionRepo) CreateDisease(ctx context.Context, d *model.InspectionDisease) error {
	d.ID = 1
	m.createdDisease = d
	return nil
}

func (m *mockInspectionRepo) Delete(ctx context.Context, inspectionID int64) error {
	m.deletedID = inspectionID
	return nil
}

func (m *mockInspectionRepo) DeleteDisease(ctx context.Context, diseaseID, inspectionID int64) error {
	m.deletedDiseaseID = diseaseID
	return nil
}

func (m *mockInspectionRepo) GetByID(ctx context.Context, inspectionID, hiveID int64) (*model.Inspection, error) {
	if m.inspection == nil || m.inspection.ID != inspectionID || m.inspection.HiveID != hiveID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.inspection, nil
}

func (m *mockInspectionRepo) GetDiseaseByID(ctx context.Context, diseaseID, inspectionID int64) (*model.InspectionDisease, error) {
	if m.disease == nil || m.disease.ID != diseaseID || m.disease.InspectionID != inspectionID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.disease, nil
}

func (m *mockInspectionRepo) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Inspection, error) {
	return m.inspections, nil
}

func (m *mockInspectionRepo) ListDiseasesByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionDisease, error) {
	if m.disease != nil && m.disease.InspectionID == inspectionID {
		return []*model.InspectionDisease{m.disease}, nil
	}
	return []*model.InspectionDisease{}, nil
}

func (m *mockInspectionRepo) LastInspectionDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error) {
	return map[int64]*time.Time{}, nil
}

func (m *mockInspectionRepo) ListDiseasesByInspectionIDs(ctx context.Context, ids []int64) ([]*model.InspectionDisease, error) {
	return []*model.InspectionDisease{}, nil
}

func (m *mockInspectionRepo) Update(ctx context.Context, insp *model.Inspection) error {
	m.updated = insp
	return nil
}

func newTestInspectionService() (*InspectionService, *mockApiaryMembershipReader, *mockInspectionHiveReader, *mockInspectionRepo) {
	apiaryMock := &mockApiaryMembershipReader{}
	hiveMock := &mockInspectionHiveReader{}
	inspMock := &mockInspectionRepo{}
	svc := NewInspectionService(apiaryMock, hiveMock, inspMock)
	return svc, apiaryMock, hiveMock, inspMock
}

var baseParams = InspectionParams{
	InspectedAt: time.Now(),
	QueenStatus: "seen",
}

func TestCreateInspection_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	insp, err := svc.Create(context.Background(), 1, 1, 10, baseParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if insp.HiveID != 10 || insp.InspectedBy != 1 {
		t.Errorf("unexpected inspection: %+v", insp)
	}
	if inspMock.created == nil {
		t.Error("expected Create to be called")
	}
}

func TestCreateInspection_InspectedAtRequired(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Create(context.Background(), 1, 1, 10, InspectionParams{})
	if !errors.Is(err, ErrInspectedAtRequired) {
		t.Errorf("expected ErrInspectedAtRequired, got %v", err)
	}
}

func TestCreateInspection_InvalidQueenStatus(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Create(context.Background(), 1, 1, 10, InspectionParams{
		InspectedAt: time.Now(),
		QueenStatus: "unknown",
	})
	if !errors.Is(err, ErrInvalidQueenStatus) {
		t.Errorf("expected ErrInvalidQueenStatus, got %v", err)
	}
}

func TestCreateInspection_InvalidBroodPattern(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Create(context.Background(), 1, 1, 10, InspectionParams{
		InspectedAt:  time.Now(),
		BroodPattern: "bad_value",
	})
	if !errors.Is(err, ErrInvalidBroodPattern) {
		t.Errorf("expected ErrInvalidBroodPattern, got %v", err)
	}
}

func TestCreateInspection_InvalidAggressiveness(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Create(context.Background(), 1, 1, 10, InspectionParams{
		InspectedAt:    time.Now(),
		Aggressiveness: "very_calm",
	})
	if !errors.Is(err, ErrInvalidAggressiveness) {
		t.Errorf("expected ErrInvalidAggressiveness, got %v", err)
	}
}

func TestCreateInspection_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Create(context.Background(), 1, 99, 10, baseParams)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestCreateInspection_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}

	_, err := svc.Create(context.Background(), 1, 1, 99, baseParams)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestGetInspection_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	insp, err := svc.Get(context.Background(), 1, 1, 10, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if insp.ID != 5 {
		t.Errorf("expected ID 5, got %d", insp.ID)
	}
}

func TestGetInspection_NotFound(t *testing.T) {
	svc, apiaryMock, hiveMock, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	_, err := svc.Get(context.Background(), 1, 1, 10, 99)
	if !errors.Is(err, ErrInspectionNotFound) {
		t.Errorf("expected ErrInspectionNotFound, got %v", err)
	}
}

func TestGetInspection_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Get(context.Background(), 1, 99, 10, 5)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestGetInspection_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}

	_, err := svc.Get(context.Background(), 1, 1, 99, 5)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestListInspections_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspections = []*model.Inspection{
		{ID: 1, HiveID: 10},
		{ID: 2, HiveID: 10},
	}

	list, err := svc.List(context.Background(), 1, 1, 10, 20, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 inspections, got %d", len(list))
	}
}

func TestListInspections_Empty(t *testing.T) {
	svc, apiaryMock, hiveMock, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	list, err := svc.List(context.Background(), 1, 1, 10, 20, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestListInspections_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.List(context.Background(), 1, 99, 10, 20, 0)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestUpdateInspection_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	updated, err := svc.Update(context.Background(), 1, 1, 10, 5, InspectionParams{
		InspectedAt: time.Now(),
		QueenStatus: "seen",
		Notes:       "updated",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.QueenStatus != "seen" || updated.Notes != "updated" {
		t.Errorf("unexpected state: %+v", updated)
	}
}

func TestUpdateInspection_NotFound(t *testing.T) {
	svc, apiaryMock, hiveMock, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	_, err := svc.Update(context.Background(), 1, 1, 10, 99, baseParams)
	if !errors.Is(err, ErrInspectionNotFound) {
		t.Errorf("expected ErrInspectionNotFound, got %v", err)
	}
}

func TestUpdateInspection_ValidationError(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.Update(context.Background(), 1, 1, 10, 5, InspectionParams{})
	if !errors.Is(err, ErrInspectedAtRequired) {
		t.Errorf("expected ErrInspectedAtRequired, got %v", err)
	}
}

func TestDeleteInspection_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	if err := svc.Delete(context.Background(), 1, 1, 10, 5); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inspMock.deletedID != 5 {
		t.Errorf("expected deleted ID 5, got %d", inspMock.deletedID)
	}
}

func TestDeleteInspection_NotFound(t *testing.T) {
	svc, apiaryMock, hiveMock, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	err := svc.Delete(context.Background(), 1, 1, 10, 99)
	if !errors.Is(err, ErrInspectionNotFound) {
		t.Errorf("expected ErrInspectionNotFound, got %v", err)
	}
}

func TestDeleteInspection_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	err := svc.Delete(context.Background(), 1, 99, 10, 5)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestDeleteInspection_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}

	err := svc.Delete(context.Background(), 1, 1, 99, 5)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestAddDisease_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	d, err := svc.AddDisease(context.Background(), 1, 1, 10, 5, "nosema", "confirmed")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.Disease != "nosema" || d.InspectionID != 5 {
		t.Errorf("unexpected disease: %+v", d)
	}
}

func TestAddDisease_InvalidDisease(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.AddDisease(context.Background(), 1, 1, 10, 5, "plague", "")
	if !errors.Is(err, ErrInvalidDisease) {
		t.Errorf("expected ErrInvalidDisease, got %v", err)
	}
}

func TestAddDisease_InspectionNotFound(t *testing.T) {
	svc, apiaryMock, hiveMock, _ := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	_, err := svc.AddDisease(context.Background(), 1, 1, 10, 99, "nosema", "")
	if !errors.Is(err, ErrInspectionNotFound) {
		t.Errorf("expected ErrInspectionNotFound, got %v", err)
	}
}

func TestAddDisease_ApiaryNotFound(t *testing.T) {
	svc, _, _, _ := newTestInspectionService()

	_, err := svc.AddDisease(context.Background(), 1, 99, 10, 5, "nosema", "")
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestRemoveDisease_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}
	inspMock.disease = &model.InspectionDisease{ID: 3, InspectionID: 5}

	if err := svc.RemoveDisease(context.Background(), 1, 1, 10, 5, 3); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inspMock.deletedDiseaseID != 3 {
		t.Errorf("expected deleted disease ID 3, got %d", inspMock.deletedDiseaseID)
	}
}

func TestRemoveDisease_DiseaseNotFound(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock := newTestInspectionService()
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	err := svc.RemoveDisease(context.Background(), 1, 1, 10, 5, 99)
	if !errors.Is(err, ErrDiseaseNotFound) {
		t.Errorf("expected ErrDiseaseNotFound, got %v", err)
	}
}
