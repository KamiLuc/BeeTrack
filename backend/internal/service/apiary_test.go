package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockApiaryRepo struct {
	created      *model.Apiary
	apiary       *model.Apiary
	memberships  []model.ApiaryMembership
	role         string
	updated      *model.Apiary
	deletedID    int64
	deepCopied   *model.Apiary
}

func (m *mockApiaryRepo) Create(ctx context.Context, a *model.Apiary, ownerRole string) error {
	a.ID = 1
	m.created = a
	return nil
}

func (m *mockApiaryRepo) ListByUserID(ctx context.Context, userID int64) ([]model.ApiaryMembership, error) {
	return m.memberships, nil
}

func (m *mockApiaryRepo) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	if m.apiary == nil || m.apiary.ID != apiaryID {
		return nil, "", gorm.ErrRecordNotFound
	}
	return m.apiary, m.role, nil
}

func (m *mockApiaryRepo) Update(ctx context.Context, a *model.Apiary) error {
	m.updated = a
	return nil
}

func (m *mockApiaryRepo) DeepCopy(ctx context.Context, sourceID, ownerID int64, newName string) (*model.Apiary, error) {
	m.deepCopied = &model.Apiary{ID: 99, OwnerUserID: ownerID, Name: newName}
	return m.deepCopied, nil
}

func (m *mockApiaryRepo) Delete(ctx context.Context, apiaryID int64) error {
	m.deletedID = apiaryID
	return nil
}

type mockHiveRelocator struct {
	hives  []*model.Hive
	moved  [][3]int // [hiveID, row, col]
}

func (m *mockHiveRelocator) ListByApiaryID(_ context.Context, _ int64) ([]*model.Hive, error) {
	return m.hives, nil
}

func (m *mockHiveRelocator) Move(_ context.Context, hiveID int64, row, col int) error {
	m.moved = append(m.moved, [3]int{int(hiveID), row, col})
	return nil
}

func newTestApiaryService() (*ApiaryService, *mockApiaryRepo, *mockHiveRelocator) {
	apiaryRepo := &mockApiaryRepo{}
	hiveRepo := &mockHiveRelocator{}
	svc := NewApiaryService(apiaryRepo, hiveRepo)
	return svc, apiaryRepo, hiveRepo
}

func TestCreateApiary_Success(t *testing.T) {
	svc, repo, _ := newTestApiaryService()

	lat, lng := 52.23, 21.01
	apiary, err := svc.Create(context.Background(), 1, "My Apiary", &lat, &lng, 3, 4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if apiary.Name != "My Apiary" {
		t.Errorf("expected name 'My Apiary', got %s", apiary.Name)
	}
	if repo.created.OwnerUserID != 1 {
		t.Errorf("expected owner user ID 1, got %d", repo.created.OwnerUserID)
	}
}

func TestCreateApiary_NoName(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	_, err := svc.Create(context.Background(), 1, "", nil, nil, 3, 4)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestCreateApiary_InvalidGridSize(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	_, err := svc.Create(context.Background(), 1, "My Apiary", nil, nil, 0, 4)
	if !errors.Is(err, ErrInvalidGridSize) {
		t.Errorf("expected ErrInvalidGridSize, got %v", err)
	}

	_, err = svc.Create(context.Background(), 1, "My Apiary", nil, nil, 3, 0)
	if !errors.Is(err, ErrInvalidGridSize) {
		t.Errorf("expected ErrInvalidGridSize, got %v", err)
	}
}

func TestCreateApiary_WithoutGPS(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	apiary, err := svc.Create(context.Background(), 1, "My Apiary", nil, nil, 3, 4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if apiary.Lat != nil || apiary.Lng != nil {
		t.Error("expected nil GPS coords")
	}
}

func TestUpdateApiary_Success(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10, Name: "Old Name", GridRows: 2, GridCols: 2}
	repo.role = "owner"

	apiary, err := svc.Update(context.Background(), 1, 10, "New Name", nil, nil, 3, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if apiary.Name != "New Name" {
		t.Errorf("expected 'New Name', got %s", apiary.Name)
	}
	if repo.updated.GridRows != 3 || repo.updated.GridCols != 5 {
		t.Errorf("unexpected grid size: %dx%d", repo.updated.GridRows, repo.updated.GridCols)
	}
}

func TestUpdateApiary_NotFound(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	_, err := svc.Update(context.Background(), 1, 99, "Name", nil, nil, 2, 2)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestUpdateApiary_Forbidden(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10}
	repo.role = "member"

	_, err := svc.Update(context.Background(), 1, 10, "Name", nil, nil, 2, 2)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateApiary_ValidationErrors(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10}
	repo.role = "owner"

	_, err := svc.Update(context.Background(), 1, 10, "", nil, nil, 2, 2)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}

	_, err = svc.Update(context.Background(), 1, 10, "Name", nil, nil, 0, 2)
	if !errors.Is(err, ErrInvalidGridSize) {
		t.Errorf("expected ErrInvalidGridSize, got %v", err)
	}
}

func TestUpdateApiary_GridTooSmall(t *testing.T) {
	svc, repo, hives := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10, GridRows: 3, GridCols: 3}
	repo.role = "owner"
	hives.hives = []*model.Hive{
		{ID: 1, GridRow: 0, GridCol: 0},
		{ID: 2, GridRow: 0, GridCol: 1},
		{ID: 3, GridRow: 1, GridCol: 0},
	}

	_, err := svc.Update(context.Background(), 1, 10, "Name", nil, nil, 1, 1)
	if !errors.Is(err, ErrGridTooSmall) {
		t.Errorf("expected ErrGridTooSmall, got %v", err)
	}
}

func TestUpdateApiary_MovesOutOfBoundsHives(t *testing.T) {
	svc, repo, hives := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10, GridRows: 3, GridCols: 3}
	repo.role = "owner"
	hives.hives = []*model.Hive{
		{ID: 1, GridRow: 0, GridCol: 0},
		{ID: 2, GridRow: 2, GridCol: 2},
	}

	_, err := svc.Update(context.Background(), 1, 10, "Name", nil, nil, 2, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hives.moved) != 1 {
		t.Fatalf("expected 1 hive moved, got %d", len(hives.moved))
	}
	movedID, movedRow, movedCol := hives.moved[0][0], hives.moved[0][1], hives.moved[0][2]
	if movedID != 2 {
		t.Errorf("expected hive 2 to be moved, got hive %d", movedID)
	}
	if movedRow >= 2 || movedCol >= 2 {
		t.Errorf("moved hive out of new bounds: row=%d col=%d", movedRow, movedCol)
	}
}

func TestUpdateApiary_NoHivesNoMove(t *testing.T) {
	svc, repo, hives := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10, GridRows: 5, GridCols: 5}
	repo.role = "owner"

	_, err := svc.Update(context.Background(), 1, 10, "Name", nil, nil, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hives.moved) != 0 {
		t.Errorf("expected no moves, got %d", len(hives.moved))
	}
}

func TestDeleteApiary_Success(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10}
	repo.role = "owner"

	err := svc.Delete(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedID != 10 {
		t.Errorf("expected deleted ID 10, got %d", repo.deletedID)
	}
}

func TestDeleteApiary_NotFound(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	err := svc.Delete(context.Background(), 1, 99)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestDeleteApiary_Forbidden(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10}
	repo.role = "member"

	err := svc.Delete(context.Background(), 1, 10)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestListApiaries_ReturnsMemberships(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	ts := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	repo.memberships = []model.ApiaryMembership{
		{Apiary: &model.Apiary{ID: 1, Name: "Alpha"}, UserRole: "owner", HiveCount: 3, LastInspectedAt: &ts},
		{Apiary: &model.Apiary{ID: 2, Name: "Beta"}, UserRole: "member", HiveCount: 0},
	}

	list, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 apiaries, got %d", len(list))
	}
	if list[0].UserRole != "owner" || list[1].UserRole != "member" {
		t.Errorf("unexpected roles: %s, %s", list[0].UserRole, list[1].UserRole)
	}
	if list[0].HiveCount != 3 {
		t.Errorf("expected HiveCount 3, got %d", list[0].HiveCount)
	}
	if list[0].LastInspectedAt == nil || !list[0].LastInspectedAt.Equal(ts) {
		t.Errorf("expected LastInspectedAt %v, got %v", ts, list[0].LastInspectedAt)
	}
	if list[1].LastInspectedAt != nil {
		t.Errorf("expected nil LastInspectedAt for Beta, got %v", list[1].LastInspectedAt)
	}
}

func TestCopyApiary_Success(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10, Name: "My Apiary"}
	repo.role = "member"

	result, err := svc.Copy(context.Background(), 1, 10, "My Apiary (copy)")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "My Apiary (copy)" {
		t.Errorf("expected name 'My Apiary (copy)', got %s", result.Name)
	}
	if repo.deepCopied.OwnerUserID != 1 {
		t.Errorf("expected owner ID 1, got %d", repo.deepCopied.OwnerUserID)
	}
}

func TestCopyApiary_NotFound(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	_, err := svc.Copy(context.Background(), 1, 99, "")
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestCopyApiary_MemberCanCopy(t *testing.T) {
	svc, repo, _ := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10, Name: "Shared Apiary"}
	repo.role = "member"

	result, err := svc.Copy(context.Background(), 2, 10, "Shared Apiary (kopia)")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.OwnerUserID != 2 {
		t.Errorf("expected copy owned by user 2, got %d", result.OwnerUserID)
	}
}

func TestListApiaries_Empty(t *testing.T) {
	svc, _, _ := newTestApiaryService()

	list, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}
