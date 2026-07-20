package service

import (
	"context"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
)

// UserService handles business logic for user profile fields not owned by AuthService.
type UserService struct {
	users UserRepository
}

// NewUserService creates a UserService with the given dependencies.
func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Me(ctx context.Context, userID int64) (*model.User, error) {
	return s.users.GetByID(ctx, userID)
}

// UpdateName validates and persists a user's display name.
func (s *UserService) UpdateName(ctx context.Context, userID int64, name string) error {
	if name == "" {
		return ErrNameRequired
	}
	if validation.TooLong(name, validation.Small) {
		return ErrNameTooLong
	}

	if err := s.users.UpdateName(ctx, userID, name); err != nil {
		return fmt.Errorf("update name: %w", err)
	}

	return nil
}
