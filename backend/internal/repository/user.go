package repository

import (
	"context"
	"errors"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &u, nil
}

func (r *UserRepository) UpdateName(ctx context.Context, userID int64, name string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"name":       name,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}
