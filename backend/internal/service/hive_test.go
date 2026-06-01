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
	occupied bool
	created  *model.Hive
	hives    []*model.Hive
}

func (m *mockHiveRepo) Create(ctx context.Context, h *model.Hive) error {
	h.ID = 1
	m.created = h
	return nil
}

func (m *mockHiveRepo) IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error) {
	return m.occupied, nil
}

func (m *mockHiveRepo) ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error) {
	return m.hives, nil
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

func TestAddHive_PositionOccupied(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.occupied = true

	_, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", 0, 0)
	if !errors.Is(err, ErrPositionOccupied) {
		t.Errorf("expected ErrPositionOccupied, got %v", err)
	}
}
