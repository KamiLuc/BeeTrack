package repository

import (
	"context"
	"errors"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// HoneyBatchCertificationRepository persists the append-only certification
// history for honey batches.
type HoneyBatchCertificationRepository struct {
	db *gorm.DB
}

// NewHoneyBatchCertificationRepository creates a new HoneyBatchCertificationRepository backed by db.
func NewHoneyBatchCertificationRepository(db *gorm.DB) *HoneyBatchCertificationRepository {
	return &HoneyBatchCertificationRepository{db: db}
}

// Create inserts a new certification attempt row.
func (r *HoneyBatchCertificationRepository) Create(ctx context.Context, c *model.HoneyBatchCertification) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// GetLatestByBatchID returns the most recent certification attempt for
// batchID, or nil if the batch has none yet.
func (r *HoneyBatchCertificationRepository) GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error) {
	var c model.HoneyBatchCertification
	err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("created_at DESC").
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListByBatchID returns the full certification history for batchID, most
// recent first.
func (r *HoneyBatchCertificationRepository) ListByBatchID(ctx context.Context, batchID int64) ([]*model.HoneyBatchCertification, error) {
	var certs []*model.HoneyBatchCertification
	err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("created_at DESC").
		Find(&certs).Error
	return certs, err
}

// UpdateStatus transitions the certification with the given id to status,
// applying any additional column updates (e.g. transaction_hash, block_number,
// gas_used, confirmation_timestamp) in the same statement.
func (r *HoneyBatchCertificationRepository) UpdateStatus(ctx context.Context, id int64, status model.CertificationStatus, fields map[string]any) error {
	updates := map[string]any{"status": status}
	for k, v := range fields {
		updates[k] = v
	}
	return r.db.WithContext(ctx).
		Model(&model.HoneyBatchCertification{}).
		Where("id = ?", id).
		Updates(updates).Error
}
