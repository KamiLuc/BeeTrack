package repository

import (
	"context"
	"errors"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// HoneyBatchQRCodeRepository persists generated QR code data for honey batches.
type HoneyBatchQRCodeRepository struct {
	db *gorm.DB
}

// NewHoneyBatchQRCodeRepository creates a new HoneyBatchQRCodeRepository backed by db.
func NewHoneyBatchQRCodeRepository(db *gorm.DB) *HoneyBatchQRCodeRepository {
	return &HoneyBatchQRCodeRepository{db: db}
}

// GetByBatchID returns the QR code row for batchID, or nil if none has been
// generated yet.
func (r *HoneyBatchQRCodeRepository) GetByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchQRCode, error) {
	var q model.HoneyBatchQRCode
	err := r.db.WithContext(ctx).Where("batch_id = ?", batchID).First(&q).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Create inserts a new QR code row.
func (r *HoneyBatchQRCodeRepository) Create(ctx context.Context, q *model.HoneyBatchQRCode) error {
	return r.db.WithContext(ctx).Create(q).Error
}
