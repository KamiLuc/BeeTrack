package service

import (
	"context"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockApiaryRepo struct {
	created   *model.Apiary
	apiary    *model.Apiary
	role      string
	updated   *model.Apiary
	deletedID int64
}

func (m *mockApiaryRepo) Create(ctx context.Context, a *model.Apiary, ownerRole string) error {
	a.ID = 1
	m.created = a
	return nil
}

func (m *mockApiaryRepo) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	if m.apiary == nil {
		return nil, "", gorm.ErrRecordNotFound
	}
	return m.apiary, m.role, nil
}

func (m *mockApiaryRepo) Update(ctx context.Context, a *model.Apiary) error {
	m.updated = a
	return nil
}

func (m *mockApiaryRepo) Delete(ctx context.Context, apiaryID int64) error {
	m.deletedID = apiaryID
	return nil
}

func newTestApiaryService() (*ApiaryService, *mockApiaryRepo) {
	repo := &mockApiaryRepo{}
	svc := NewApiaryService(repo)
	return svc, repo
}

func TestCreateApiary_Success(t *testing.T) {
	svc, repo := newTestApiaryService()

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
	svc, _ := newTestApiaryService()

	_, err := svc.Create(context.Background(), 1, "", nil, nil, 3, 4)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestCreateApiary_InvalidGridSize(t *testing.T) {
	svc, _ := newTestApiaryService()

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
	svc, _ := newTestApiaryService()

	apiary, err := svc.Create(context.Background(), 1, "My Apiary", nil, nil, 3, 4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if apiary.Lat != nil || apiary.Lng != nil {
		t.Error("expected nil GPS coords")
	}
}

func TestUpdateApiary_Success(t *testing.T) {
	svc, repo := newTestApiaryService()
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
	svc, _ := newTestApiaryService()

	_, err := svc.Update(context.Background(), 1, 99, "Name", nil, nil, 2, 2)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestUpdateApiary_Forbidden(t *testing.T) {
	svc, repo := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10}
	repo.role = "member"

	_, err := svc.Update(context.Background(), 1, 10, "Name", nil, nil, 2, 2)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateApiary_ValidationErrors(t *testing.T) {
	svc, repo := newTestApiaryService()
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

func TestDeleteApiary_Success(t *testing.T) {
	svc, repo := newTestApiaryService()
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
	svc, _ := newTestApiaryService()

	err := svc.Delete(context.Background(), 1, 99)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestDeleteApiary_Forbidden(t *testing.T) {
	svc, repo := newTestApiaryService()
	repo.apiary = &model.Apiary{ID: 10}
	repo.role = "member"

	err := svc.Delete(context.Background(), 1, 10)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
