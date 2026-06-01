package service

import (
	"context"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockApiaryMembershipReader struct {
	apiary *model.Apiary
	role   string
}

func (m *mockApiaryMembershipReader) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	if m.apiary == nil || m.apiary.ID != apiaryID {
		return nil, "", gorm.ErrRecordNotFound
	}
	return m.apiary, m.role, nil
}

type mockHiveRepo struct {
	occupied  bool
	created   *model.Hive
	hive      *model.Hive
	hives     []*model.Hive
	updated   *model.Hive
	deletedID int64
}

func (m *mockHiveRepo) Create(ctx context.Context, h *model.Hive) error {
	h.ID = 1
	m.created = h
	return nil
}

func (m *mockHiveRepo) GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error) {
	if m.hive == nil || m.hive.ID != hiveID || m.hive.ApiaryID != apiaryID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.hive, nil
}

func (m *mockHiveRepo) IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error) {
	return m.occupied, nil
}

func (m *mockHiveRepo) ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error) {
	return m.hives, nil
}

func (m *mockHiveRepo) Move(ctx context.Context, hiveID int64, row, col int) error {
	if m.hive != nil && m.hive.ID == hiveID {
		m.hive.GridRow = row
		m.hive.GridCol = col
	}
	return nil
}

func (m *mockHiveRepo) Update(ctx context.Context, h *model.Hive) error {
	m.updated = h
	return nil
}

func (m *mockHiveRepo) Delete(ctx context.Context, hiveID int64) error {
	m.deletedID = hiveID
	return nil
}

func newTestHiveService() (*HiveService, *mockApiaryMembershipReader, *mockHiveRepo) {
	apiaryMock := &mockApiaryMembershipReader{}
	hiveMock := &mockHiveRepo{}
	svc := NewHiveService(apiaryMock, hiveMock)
	return svc, apiaryMock, hiveMock
}

func TestListHives_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hives = []*model.Hive{
		{ID: 1, Name: "Hive A", GridRow: 0, GridCol: 0},
		{ID: 2, Name: "Hive B", GridRow: 1, GridCol: 2},
	}

	hives, err := svc.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hives) != 2 {
		t.Fatalf("expected 2 hives, got %d", len(hives))
	}
}

func TestListHives_Empty(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	hives, err := svc.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hives) != 0 {
		t.Errorf("expected empty list, got %d hives", len(hives))
	}
}

func TestListHives_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.List(context.Background(), 1, 99)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestAddHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	apiaryMock.role = "owner"

	hive, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.Name != "Hive A" {
		t.Errorf("expected name 'Hive A', got %s", hive.Name)
	}
	if !hiveMock.created.Active {
		t.Error("expected hive to be active by default")
	}
}

func TestAddHive_MemberCanAdd(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	apiaryMock.role = "member"

	_, err := svc.Add(context.Background(), 2, 1, "Hive B", "langstroth", 1, 1)
	if err != nil {
		t.Fatalf("members should be allowed to add hives, got %v", err)
	}
}

func TestAddHive_NoName(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Add(context.Background(), 1, 1, "", "langstroth", 0, 0)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestAddHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Add(context.Background(), 1, 99, "Hive A", "langstroth", 0, 0)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestAddHive_OutOfBounds(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	cases := [][2]int{{-1, 0}, {0, -1}, {3, 0}, {0, 4}}
	for _, c := range cases {
		_, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", c[0], c[1])
		if !errors.Is(err, ErrInvalidGridPosition) {
			t.Errorf("row=%d col=%d: expected ErrInvalidGridPosition, got %v", c[0], c[1], err)
		}
	}
}

func TestUpdateHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, Name: "Old", Type: "langstroth", Active: true}

	hive, err := svc.Update(context.Background(), 1, 1, 10, "New Name", "top_bar", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.Name != "New Name" || hive.Type != "top_bar" || hive.Active {
		t.Errorf("unexpected hive state: %+v", hive)
	}
}

func TestUpdateHive_NoName(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Update(context.Background(), 1, 1, 10, "", "langstroth", true)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestUpdateHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Update(context.Background(), 1, 99, 10, "Name", "langstroth", true)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestUpdateHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Update(context.Background(), 1, 1, 99, "Name", "langstroth", true)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestDeleteHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	if err := svc.Delete(context.Background(), 1, 1, 10); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hiveMock.deletedID != 10 {
		t.Errorf("expected deleted ID 10, got %d", hiveMock.deletedID)
	}
}

func TestDeleteHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	err := svc.Delete(context.Background(), 1, 99, 10)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestDeleteHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	err := svc.Delete(context.Background(), 1, 1, 99)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestMoveHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}

	hive, err := svc.Move(context.Background(), 1, 1, 10, 2, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.GridRow != 2 || hive.GridCol != 3 {
		t.Errorf("expected position (2,3), got (%d,%d)", hive.GridRow, hive.GridCol)
	}
}

func TestMoveHive_SamePosition_NoOp(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 1, GridCol: 1}
	hiveMock.occupied = true

	_, err := svc.Move(context.Background(), 1, 1, 10, 1, 1)
	if err != nil {
		t.Fatalf("moving to same position should not error, got %v", err)
	}
}

func TestMoveHive_OutOfBounds(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}

	_, err := svc.Move(context.Background(), 1, 1, 10, 3, 0)
	if !errors.Is(err, ErrInvalidGridPosition) {
		t.Errorf("expected ErrInvalidGridPosition, got %v", err)
	}
}

func TestMoveHive_PositionOccupied(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}
	hiveMock.occupied = true

	_, err := svc.Move(context.Background(), 1, 1, 10, 1, 1)
	if !errors.Is(err, ErrPositionOccupied) {
		t.Errorf("expected ErrPositionOccupied, got %v", err)
	}
}

func TestMoveHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Move(context.Background(), 1, 99, 10, 0, 0)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestMoveHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Move(context.Background(), 1, 1, 99, 0, 0)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestAddHive_PositionOccupied(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.occupied = true

	_, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", 0, 0)
	if !errors.Is(err, ErrPositionOccupied) {
		t.Errorf("expected ErrPositionOccupied, got %v", err)
	}
}
