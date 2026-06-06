package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailNotVerified        = errors.New("email address not verified")
	ErrEmailTaken              = errors.New("email already registered")
	ErrInvalidEmail            = errors.New("invalid email address")
	ErrInvalidPassword         = errors.New("invalid email or password")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrInvalidResetToken       = errors.New("invalid or expired password reset token")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrTokenExpired            = errors.New("refresh token expired")
	ErrWeakPassword            = errors.New("password must be at least 8 characters")
)

type EmailTokenRepository interface {
	CreateVerificationToken(ctx context.Context, t *model.EmailVerificationToken) error
	DeleteVerificationToken(ctx context.Context, token string) error
	DeleteVerificationTokensByUserID(ctx context.Context, userID int64) error
	GetVerificationToken(ctx context.Context, token string) (*model.EmailVerificationToken, error)
	CreatePasswordResetToken(ctx context.Context, t *model.PasswordResetToken) error
	DeletePasswordResetToken(ctx context.Context, token string) error
	DeletePasswordResetTokensByUserID(ctx context.Context, userID int64) error
	GetPasswordResetToken(ctx context.Context, token string) (*model.PasswordResetToken, error)
}

type Mailer interface {
	SendPasswordResetEmail(ctx context.Context, to, name, resetURL, lang string) error
	SendVerificationEmail(ctx context.Context, to, name, verificationURL, lang string) error
}

type TokenRepository interface {
	Create(ctx context.Context, t *model.RefreshToken) error
	DeleteByToken(ctx context.Context, token string) error
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	SetVerified(ctx context.Context, userID int64) error
	UpdateName(ctx context.Context, userID int64, name string) error
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
}

type AuthService struct {
	accessTTLMin   int
	apiURL         string
	appURL         string
	emailTokens    EmailTokenRepository
	jwtSecret      string
	mailer         Mailer
	refreshTTLDays int
	tokens         TokenRepository
	users          UserRepository
}

func NewAuthService(
	users UserRepository,
	tokens TokenRepository,
	emailTokens EmailTokenRepository,
	mailer Mailer,
	apiURL string,
	appURL string,
	jwtSecret string,
	accessTTLMin int,
	refreshTTLDays int,
) *AuthService {
	return &AuthService{
		accessTTLMin:   accessTTLMin,
		apiURL:         apiURL,
		appURL:         appURL,
		emailTokens:    emailTokens,
		jwtSecret:      jwtSecret,
		mailer:         mailer,
		refreshTTLDays: refreshTTLDays,
		tokens:         tokens,
		users:          users,
	}
}

// ForgotPassword initiates a password reset for the given email. Always returns nil to
// avoid leaking whether the email is registered.
func (s *AuthService) ForgotPassword(ctx context.Context, email, lang string) error {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("forgot password: %w", err)
	}
	if user == nil {
		return nil
	}

	if err := s.emailTokens.DeletePasswordResetTokensByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("clear old reset tokens: %w", err)
	}

	rawToken, err := token.NewRefreshToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	rt := &model.PasswordResetToken{
		UserID:    user.ID,
		Token:     rawToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := s.emailTokens.CreatePasswordResetToken(ctx, rt); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	resetURL := s.apiURL + "/api/v1/auth/reset-password-form?token=" + rawToken + "&lang=" + lang
	if err := s.mailer.SendPasswordResetEmail(ctx, user.Email, user.Name, resetURL, lang); err != nil {
		log.Printf("failed to send password reset email to %s: %v", user.Email, err)
	}

	return nil
}

// Login authenticates a user and returns a token pair. Returns ErrEmailNotVerified if
// the account has not been verified yet.
func (s *AuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("login: %w", err)
	}
	if user == nil {
		return "", "", ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidPassword
	}

	if !user.Verified {
		return "", "", ErrEmailNotVerified
	}

	accessToken, err = token.NewAccessToken(user.ID, s.jwtSecret, s.accessTTLMin)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = token.NewRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().AddDate(0, 0, s.refreshTTLDays),
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Logout revokes the refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokens.DeleteByToken(ctx, refreshToken)
}

// Refresh exchanges a refresh token for a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	rt, err := s.tokens.GetByToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("refresh: %w", err)
	}
	if rt == nil {
		return "", "", ErrInvalidRefreshToken
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", ErrTokenExpired
	}

	if err := s.tokens.DeleteByToken(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("delete old refresh token: %w", err)
	}

	accessToken, err = token.NewAccessToken(rt.UserID, s.jwtSecret, s.accessTTLMin)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err = token.NewRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	newRT := &model.RefreshToken{
		UserID:    rt.UserID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().AddDate(0, 0, s.refreshTTLDays),
	}
	if err := s.tokens.Create(ctx, newRT); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// Register creates a new unverified user account and sends a verification email in the
// requested language.
func (s *AuthService) Register(ctx context.Context, email, name, password, lang string) (*model.User, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
		Verified:     false,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := s.sendVerificationEmail(ctx, u, lang); err != nil {
		log.Printf("failed to send verification email to %s: %v", u.Email, err)
	}

	return u, nil
}

// ResendVerification sends a new verification email. Always returns nil to avoid
// leaking whether the email is registered or already verified.
func (s *AuthService) ResendVerification(ctx context.Context, email, lang string) error {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("resend verification: %w", err)
	}
	if user == nil || user.Verified {
		return nil
	}

	if err := s.sendVerificationEmail(ctx, user, lang); err != nil {
		log.Printf("failed to resend verification email to %s: %v", user.Email, err)
	}

	return nil
}

// ResetPassword validates the reset token and updates the user's password.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	rt, err := s.emailTokens.GetPasswordResetToken(ctx, rawToken)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if rt == nil || time.Now().After(rt.ExpiresAt) {
		return ErrInvalidResetToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, rt.UserID, string(hash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := s.emailTokens.DeletePasswordResetTokensByUserID(ctx, rt.UserID); err != nil {
		return fmt.Errorf("delete reset tokens: %w", err)
	}

	return nil
}

// VerifyEmail validates the verification token and marks the user's account as verified.
func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	vt, err := s.emailTokens.GetVerificationToken(ctx, rawToken)
	if err != nil {
		return fmt.Errorf("verify email: %w", err)
	}
	if vt == nil || time.Now().After(vt.ExpiresAt) {
		return ErrInvalidVerificationToken
	}

	if err := s.users.SetVerified(ctx, vt.UserID); err != nil {
		return fmt.Errorf("set verified: %w", err)
	}

	if err := s.emailTokens.DeleteVerificationToken(ctx, rawToken); err != nil {
		return fmt.Errorf("delete verification token: %w", err)
	}

	return nil
}

func (s *AuthService) sendVerificationEmail(ctx context.Context, user *model.User, lang string) error {
	if err := s.emailTokens.DeleteVerificationTokensByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("clear old verification tokens: %w", err)
	}

	rawToken, err := token.NewRefreshToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	vt := &model.EmailVerificationToken{
		UserID:    user.ID,
		Token:     rawToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.emailTokens.CreateVerificationToken(ctx, vt); err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}

	verificationURL := s.apiURL + "/api/v1/auth/verify-email?token=" + rawToken + "&lang=" + lang
	return s.mailer.SendVerificationEmail(ctx, user.Email, user.Name, verificationURL, lang)
}
