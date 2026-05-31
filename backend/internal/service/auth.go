package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrInvalidEmail = errors.New("invalid email address")
	ErrWeakPassword = errors.New("password must be at least 8 characters")
)

type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
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
