package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type ApiaryRepository struct {
	db *gorm.DB
}

func NewApiaryRepository(db *gorm.DB) *ApiaryRepository {
	return &ApiaryRepository{db: db}
}

func (r *ApiaryRepository) Create(ctx context.Context, a *model.Apiary, ownerRole string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
		member := &model.ApiaryMember{
			ApiaryID: a.ID,
			UserID:   a.OwnerUserID,
			Role:     ownerRole,
		}
		return tx.Create(member).Error
	})
}
