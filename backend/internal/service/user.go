package service

import (
	"context"
	"fmt"
)

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) UpdateName(ctx context.Context, userID int64, name string) error {
	if name == "" {
		return ErrNameRequired
	}

	if err := s.users.UpdateName(ctx, userID, name); err != nil {
		return fmt.Errorf("update name: %w", err)
	}

	return nil
}
