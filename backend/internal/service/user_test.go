package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

type mockUpdateNameRepo struct {
	updatedID   int64
	updatedName string
}

func (m *mockUpdateNameRepo) Create(ctx context.Context, u *model.User) error { return nil }
func (m *mockUpdateNameRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}
func (m *mockUpdateNameRepo) SetVerified(ctx context.Context, userID int64) error { return nil }
func (m *mockUpdateNameRepo) UpdateName(ctx context.Context, userID int64, name string) error {
	m.updatedID = userID
	m.updatedName = name
	return nil
}
func (m *mockUpdateNameRepo) UpdatePassword(ctx context.Context, userID int64, hash string) error {
	return nil
}

func newTestUserService() (*UserService, *mockUpdateNameRepo) {
	repo := &mockUpdateNameRepo{}
	svc := NewUserService(repo)
	return svc, repo
}

func TestUpdateName_Success(t *testing.T) {
	svc, repo := newTestUserService()

	if err := svc.UpdateName(context.Background(), 42, "Alice"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedID != 42 {
		t.Errorf("expected userID 42, got %d", repo.updatedID)
	}
	if repo.updatedName != "Alice" {
		t.Errorf("expected name 'Alice', got %s", repo.updatedName)
	}
}

func TestUpdateName_EmptyName(t *testing.T) {
	svc, _ := newTestUserService()

	err := svc.UpdateName(context.Background(), 1, "")
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestUpdateName_TooLong(t *testing.T) {
	svc, _ := newTestUserService()

	err := svc.UpdateName(context.Background(), 1, strings.Repeat("a", 51))
	if !errors.Is(err, ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong, got %v", err)
	}
}
