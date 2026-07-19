package repository

import (
	"context"
	"errors"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// HoneyBatchRepository persists honey batches.
type HoneyBatchRepository struct {
	db *gorm.DB
}

// NewHoneyBatchRepository creates a new HoneyBatchRepository backed by db.
func NewHoneyBatchRepository(db *gorm.DB) *HoneyBatchRepository {
	return &HoneyBatchRepository{db: db}
}

// Create inserts a new honey batch record.
func (r *HoneyBatchRepository) Create(ctx context.Context, b *model.HoneyBatch) error {
	return r.db.WithContext(ctx).Create(b).Error
}

// GetByID returns the non-deleted batch with the given id, or nil if not found.
func (r *HoneyBatchRepository) GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	var b model.HoneyBatch
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetByVerificationToken returns the non-deleted batch with the given
// verification token, or nil if not found. Used by the public verification path.
func (r *HoneyBatchRepository) GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error) {
	var b model.HoneyBatch
	err := r.db.WithContext(ctx).
		Where("verification_token = ? AND deleted_at IS NULL", token).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListByUserID returns non-deleted batches created by userID, ordered by
// creation time descending, with pagination.
func (r *HoneyBatchRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.HoneyBatch, error) {
	var batches []*model.HoneyBatch
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&batches).Error
	return batches, err
}

// ListByApiaryID returns non-deleted batches for apiaryID, ordered by
// creation time descending, with pagination.
func (r *HoneyBatchRepository) ListByApiaryID(ctx context.Context, apiaryID int64, limit, offset int) ([]*model.HoneyBatch, error) {
	var batches []*model.HoneyBatch
	err := r.db.WithContext(ctx).
		Where("apiary_id = ? AND deleted_at IS NULL", apiaryID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&batches).Error
	return batches, err
}

// UpdateNotes overwrites the honey_type of the given batch — the only
// user-editable field on honey_batches (see honey_batches schema, HC-DB-01).
func (r *HoneyBatchRepository) UpdateNotes(ctx context.Context, id int64, honeyType string) error {
	return r.db.WithContext(ctx).
		Model(&model.HoneyBatch{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"honey_type": honeyType,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// SoftDelete marks the batch as deleted without removing its row. The
// on-chain certification remains untouched and immutable.
func (r *HoneyBatchRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.HoneyBatch{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}
