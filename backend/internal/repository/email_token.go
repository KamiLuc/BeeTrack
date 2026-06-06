package repository

import (
	"context"
	"errors"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type EmailTokenRepository struct {
	db *gorm.DB
}

func NewEmailTokenRepository(db *gorm.DB) *EmailTokenRepository {
	return &EmailTokenRepository{db: db}
}

func (r *EmailTokenRepository) CreateVerificationToken(ctx context.Context, t *model.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *EmailTokenRepository) GetVerificationToken(ctx context.Context, token string) (*model.EmailVerificationToken, error) {
	var t model.EmailVerificationToken
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &t, nil
}

func (r *EmailTokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.EmailVerificationToken{}).Error
}

func (r *EmailTokenRepository) DeleteVerificationTokensByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.EmailVerificationToken{}).Error
}

func (r *EmailTokenRepository) CreatePasswordResetToken(ctx context.Context, t *model.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *EmailTokenRepository) GetPasswordResetToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	var t model.PasswordResetToken
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &t, nil
}

func (r *EmailTokenRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.PasswordResetToken{}).Error
}

func (r *EmailTokenRepository) DeletePasswordResetTokensByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.PasswordResetToken{}).Error
}
