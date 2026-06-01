package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidPassword     = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenExpired        = errors.New("refresh token expired")
	ErrWeakPassword        = errors.New("password must be at least 8 characters")
)

type TokenRepository interface {
	Create(ctx context.Context, t *model.RefreshToken) error
	DeleteByToken(ctx context.Context, token string) error
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateName(ctx context.Context, userID int64, name string) error
}

type AuthService struct {
	accessTTLMin   int
	jwtSecret      string
	refreshTTLDays int
	tokens         TokenRepository
	users          UserRepository
}

func NewAuthService(
	users UserRepository,
	tokens TokenRepository,
	jwtSecret string,
	accessTTLMin int,
	refreshTTLDays int,
) *AuthService {
	return &AuthService{
		accessTTLMin:   accessTTLMin,
		jwtSecret:      jwtSecret,
		refreshTTLDays: refreshTTLDays,
		tokens:         tokens,
		users:          users,
	}
}

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

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokens.DeleteByToken(ctx, refreshToken)
}

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

func (s *AuthService) Register(ctx context.Context, email, name, password string) (*model.User, error) {
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
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}
