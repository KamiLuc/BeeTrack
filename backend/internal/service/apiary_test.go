package service

import (
	"context"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

type mockApiaryRepo struct {
	created *model.Apiary
}

func (m *mockApiaryRepo) Create(ctx context.Context, a *model.Apiary, ownerRole string) error {
	a.ID = 1
	m.created = a
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
