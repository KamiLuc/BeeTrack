package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// -- mocks --

type mockUserRepo struct {
	users map[string]*model.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error {
	u.ID = int64(len(m.users) + 1)
	u.CreatedAt = time.Now()
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

type mockTokenRepo struct {
	tokens map[string]*model.RefreshToken
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{tokens: make(map[string]*model.RefreshToken)}
}

func (m *mockTokenRepo) Create(ctx context.Context, t *model.RefreshToken) error {
	m.tokens[t.Token] = t
	return nil
}

func (m *mockTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *mockTokenRepo) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	t, ok := m.tokens[token]
	if !ok {
		return nil, nil
	}
	return t, nil
}

// -- helpers --

func newTestService() (*AuthService, *mockUserRepo, *mockTokenRepo) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := NewAuthService(users, tokens, "test-secret", 15, 30)
	return svc, users, tokens
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(hash)
}

// -- Register tests --

func TestRegister_Success(t *testing.T) {
	svc, _, _ := newTestService()

	user, err := svc.Register(context.Background(), "user@example.com", "John", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", user.Email)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.Register(context.Background(), "not-an-email", "John", "password123")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.Register(context.Background(), "user@example.com", "John", "short")
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	svc, _, _ := newTestService()

	svc.Register(context.Background(), "user@example.com", "John", "password123")
	_, err := svc.Register(context.Background(), "user@example.com", "Jane", "password123")
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

// -- Login tests --

func TestLogin_Success(t *testing.T) {
	svc, users, _ := newTestService()
	users.users["user@example.com"] = &model.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: hashPassword(t, "password123"),
	}

	accessToken, refreshToken, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _, _ := newTestService()

	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, users, _ := newTestService()
	users.users["user@example.com"] = &model.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: hashPassword(t, "password123"),
	}

	_, _, err := svc.Login(context.Background(), "user@example.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

// -- Refresh tests --

func TestRefresh_Success(t *testing.T) {
	svc, _, tokens := newTestService()
	tokens.tokens["valid-token"] = &model.RefreshToken{
		UserID:    1,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	accessToken, newRefreshToken, err := svc.Refresh(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if newRefreshToken == "valid-token" {
		t.Error("expected a new refresh token, got the same one")
	}
	if _, exists := tokens.tokens["valid-token"]; exists {
		t.Error("expected old refresh token to be deleted")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc, _, _ := newTestService()

	_, _, err := svc.Refresh(context.Background(), "nonexistent-token")
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	svc, _, tokens := newTestService()
	tokens.tokens["expired-token"] = &model.RefreshToken{
		UserID:    1,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	_, _, err := svc.Refresh(context.Background(), "expired-token")
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

// -- Logout tests --

func TestLogout_Success(t *testing.T) {
	svc, _, tokens := newTestService()
	tokens.tokens["valid-token"] = &model.RefreshToken{Token: "valid-token"}

	if err := svc.Logout(context.Background(), "valid-token"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := tokens.tokens["valid-token"]; exists {
		t.Error("expected token to be deleted after logout")
	}
}
