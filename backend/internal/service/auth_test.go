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

func (m *mockUserRepo) SetVerified(ctx context.Context, userID int64) error {
	for _, u := range m.users {
		if u.ID == userID {
			u.Verified = true
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) UpdateName(ctx context.Context, userID int64, name string) error {
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	for _, u := range m.users {
		if u.ID == userID {
			u.PasswordHash = passwordHash
			return nil
		}
	}
	return nil
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

type mockEmailTokenRepo struct {
	verificationTokens map[string]*model.EmailVerificationToken
	resetTokens        map[string]*model.PasswordResetToken
}

func newMockEmailTokenRepo() *mockEmailTokenRepo {
	return &mockEmailTokenRepo{
		verificationTokens: make(map[string]*model.EmailVerificationToken),
		resetTokens:        make(map[string]*model.PasswordResetToken),
	}
}

func (m *mockEmailTokenRepo) CreateVerificationToken(_ context.Context, t *model.EmailVerificationToken) error {
	m.verificationTokens[t.Token] = t
	return nil
}

func (m *mockEmailTokenRepo) GetVerificationToken(_ context.Context, token string) (*model.EmailVerificationToken, error) {
	t, ok := m.verificationTokens[token]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *mockEmailTokenRepo) DeleteVerificationToken(_ context.Context, token string) error {
	delete(m.verificationTokens, token)
	return nil
}

func (m *mockEmailTokenRepo) DeleteVerificationTokensByUserID(_ context.Context, userID int64) error {
	for k, t := range m.verificationTokens {
		if t.UserID == userID {
			delete(m.verificationTokens, k)
		}
	}
	return nil
}

func (m *mockEmailTokenRepo) CreatePasswordResetToken(_ context.Context, t *model.PasswordResetToken) error {
	m.resetTokens[t.Token] = t
	return nil
}

func (m *mockEmailTokenRepo) GetPasswordResetToken(_ context.Context, token string) (*model.PasswordResetToken, error) {
	t, ok := m.resetTokens[token]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *mockEmailTokenRepo) DeletePasswordResetToken(_ context.Context, token string) error {
	delete(m.resetTokens, token)
	return nil
}

func (m *mockEmailTokenRepo) DeletePasswordResetTokensByUserID(_ context.Context, userID int64) error {
	for k, t := range m.resetTokens {
		if t.UserID == userID {
			delete(m.resetTokens, k)
		}
	}
	return nil
}

type mockMailer struct {
	verificationsSent []string
	resetsSent        []string
}

func (m *mockMailer) SendVerificationEmail(_ context.Context, to, _, _, _ string) error {
	m.verificationsSent = append(m.verificationsSent, to)
	return nil
}

func (m *mockMailer) SendPasswordResetEmail(_ context.Context, to, _, _, _ string) error {
	m.resetsSent = append(m.resetsSent, to)
	return nil
}

// -- helpers --

func newTestService() (*AuthService, *mockUserRepo, *mockTokenRepo, *mockEmailTokenRepo, *mockMailer) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	emailTokens := newMockEmailTokenRepo()
	mailer := &mockMailer{}
	svc := NewAuthService(users, tokens, emailTokens, mailer, "http://localhost:8080", "http://localhost:5000", "test-secret", 15, 30)
	return svc, users, tokens, emailTokens, mailer
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
	svc, _, _, emailTokens, mailer := newTestService()

	user, err := svc.Register(context.Background(), "user@example.com", "John", "password123", "en")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", user.Email)
	}
	if user.Verified {
		t.Error("expected user to be unverified after registration")
	}
	if len(emailTokens.verificationTokens) != 1 {
		t.Errorf("expected 1 verification token, got %d", len(emailTokens.verificationTokens))
	}
	if len(mailer.verificationsSent) != 1 {
		t.Errorf("expected 1 verification email, got %d", len(mailer.verificationsSent))
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.Register(context.Background(), "not-an-email", "John", "password123", "en")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.Register(context.Background(), "user@example.com", "John", "short", "en")
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	svc.Register(context.Background(), "user@example.com", "John", "password123", "en")
	_, err := svc.Register(context.Background(), "user@example.com", "Jane", "password123", "en")
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

// -- Login tests --

func TestLogin_Success(t *testing.T) {
	svc, users, _, _, _ := newTestService()
	users.users["user@example.com"] = &model.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: hashPassword(t, "password123"),
		Verified:     true,
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

func TestLogin_EmailNotVerified(t *testing.T) {
	svc, users, _, _, _ := newTestService()
	users.users["user@example.com"] = &model.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: hashPassword(t, "password123"),
		Verified:     false,
	}

	_, _, err := svc.Login(context.Background(), "user@example.com", "password123")
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Errorf("expected ErrEmailNotVerified, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, users, _, _, _ := newTestService()
	users.users["user@example.com"] = &model.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: hashPassword(t, "password123"),
		Verified:     true,
	}

	_, _, err := svc.Login(context.Background(), "user@example.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

// -- Refresh tests --

func TestRefresh_Success(t *testing.T) {
	svc, _, tokens, _, _ := newTestService()
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
	svc, _, _, _, _ := newTestService()

	_, _, err := svc.Refresh(context.Background(), "nonexistent-token")
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	svc, _, tokens, _, _ := newTestService()
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
	svc, _, tokens, _, _ := newTestService()
	tokens.tokens["valid-token"] = &model.RefreshToken{Token: "valid-token"}

	if err := svc.Logout(context.Background(), "valid-token"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := tokens.tokens["valid-token"]; exists {
		t.Error("expected token to be deleted after logout")
	}
}

// -- VerifyEmail tests --

func TestVerifyEmail_Success(t *testing.T) {
	svc, users, _, emailTokens, _ := newTestService()
	users.users["user@example.com"] = &model.User{ID: 1, Email: "user@example.com", Verified: false}
	emailTokens.verificationTokens["valid-token"] = &model.EmailVerificationToken{
		UserID:    1,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := svc.VerifyEmail(context.Background(), "valid-token"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !users.users["user@example.com"].Verified {
		t.Error("expected user to be verified")
	}
	if _, exists := emailTokens.verificationTokens["valid-token"]; exists {
		t.Error("expected verification token to be deleted")
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	err := svc.VerifyEmail(context.Background(), "nonexistent-token")
	if !errors.Is(err, ErrInvalidVerificationToken) {
		t.Errorf("expected ErrInvalidVerificationToken, got %v", err)
	}
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
	svc, _, _, emailTokens, _ := newTestService()
	emailTokens.verificationTokens["expired-token"] = &model.EmailVerificationToken{
		UserID:    1,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	err := svc.VerifyEmail(context.Background(), "expired-token")
	if !errors.Is(err, ErrInvalidVerificationToken) {
		t.Errorf("expected ErrInvalidVerificationToken, got %v", err)
	}
}

// -- ForgotPassword tests --

func TestForgotPassword_Success(t *testing.T) {
	svc, users, _, emailTokens, mailer := newTestService()
	users.users["user@example.com"] = &model.User{ID: 1, Email: "user@example.com", Name: "John", Verified: true}

	if err := svc.ForgotPassword(context.Background(), "user@example.com", "en"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(emailTokens.resetTokens) != 1 {
		t.Errorf("expected 1 reset token, got %d", len(emailTokens.resetTokens))
	}
	if len(mailer.resetsSent) != 1 {
		t.Errorf("expected 1 reset email, got %d", len(mailer.resetsSent))
	}
}

func TestForgotPassword_UnknownEmail(t *testing.T) {
	svc, _, _, emailTokens, mailer := newTestService()

	if err := svc.ForgotPassword(context.Background(), "nobody@example.com", "en"); err != nil {
		t.Fatalf("expected no error for unknown email, got %v", err)
	}
	if len(emailTokens.resetTokens) != 0 {
		t.Error("expected no reset tokens for unknown email")
	}
	if len(mailer.resetsSent) != 0 {
		t.Error("expected no reset emails for unknown email")
	}
}

// -- ResetPassword tests --

func TestResetPassword_Success(t *testing.T) {
	svc, users, _, emailTokens, _ := newTestService()
	users.users["user@example.com"] = &model.User{ID: 1, Email: "user@example.com", PasswordHash: hashPassword(t, "oldpassword")}
	emailTokens.resetTokens["valid-token"] = &model.PasswordResetToken{
		UserID:    1,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := svc.ResetPassword(context.Background(), "valid-token", "newpassword123"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := emailTokens.resetTokens["valid-token"]; exists {
		t.Error("expected reset token to be deleted")
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	err := svc.ResetPassword(context.Background(), "nonexistent-token", "newpassword123")
	if !errors.Is(err, ErrInvalidResetToken) {
		t.Errorf("expected ErrInvalidResetToken, got %v", err)
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	svc, _, _, emailTokens, _ := newTestService()
	emailTokens.resetTokens["expired-token"] = &model.PasswordResetToken{
		UserID:    1,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	err := svc.ResetPassword(context.Background(), "expired-token", "newpassword123")
	if !errors.Is(err, ErrInvalidResetToken) {
		t.Errorf("expected ErrInvalidResetToken, got %v", err)
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	svc, _, _, emailTokens, _ := newTestService()
	emailTokens.resetTokens["valid-token"] = &model.PasswordResetToken{
		UserID:    1,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err := svc.ResetPassword(context.Background(), "valid-token", "short")
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}
