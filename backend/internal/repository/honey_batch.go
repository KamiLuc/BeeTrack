package repository

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/beetrack/backend/internal/blockchain"
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

// CreateWithCertificationRequest inserts b and, if certRequest is non-nil,
// certRequest, in a single transaction. CanonicalMetadataHash requires b's
// id, which doesn't exist until after insert, so the hash is computed and
// persisted here (between the insert and the request insert) rather than by
// the caller — this is the only way to keep the batch, its hash, and its
// certification request atomic together.
func (r *HoneyBatchRepository) CreateWithCertificationRequest(ctx context.Context, b *model.HoneyBatch, certRequest *model.HoneyBatchCertificationRequest) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(b).Error; err != nil {
			return err
		}

		hash := blockchain.CanonicalMetadataHash(b)
		b.MetadataHash = hex.EncodeToString(hash[:])
		if err := tx.Model(b).Update("metadata_hash", b.MetadataHash).Error; err != nil {
			return err
		}

		if certRequest != nil {
			certRequest.BatchID = b.ID
			if err := tx.Create(certRequest).Error; err != nil {
				return err
			}
		}
		return nil
	})
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

// GetByIDIgnoringDeletion returns the batch with the given id regardless of
// soft-delete status. Used by the worker: on-chain certification is
// immutable and intentionally unaffected by a soft delete.
func (r *HoneyBatchRepository) GetByIDIgnoringDeletion(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	var b model.HoneyBatch
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error
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

// CountByUserID returns the number of non-deleted batches created by userID.
func (r *HoneyBatchRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.HoneyBatch{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// UpdateFields overwrites a batch's gathering date, amount, processing
// method, honey type, PDF fields, and recomputed metadata hash — the fields
// editable before certification has ever been requested (see
// HoneyBatchService.UpdateBatch). Always writes the PDF columns from batch,
// so callers that didn't change the PDF must pass its current values through
// unchanged.
func (r *HoneyBatchRepository) UpdateFields(ctx context.Context, batch *model.HoneyBatch) error {
	return r.db.WithContext(ctx).
		Model(&model.HoneyBatch{}).
		Where("id = ? AND deleted_at IS NULL", batch.ID).
		Updates(map[string]any{
			"gathering_date":    batch.GatheringDate,
			"amount_grams":      batch.AmountGrams,
			"processing_method": batch.ProcessingMethod,
			"honey_type":        batch.HoneyType,
			"lab_pdf_url":       batch.LabPDFURL,
			"pdf_filename":      batch.PDFFilename,
			"pdf_file_hash":     batch.PDFFileHash,
			"metadata_hash":     batch.MetadataHash,
			"updated_at":        gorm.Expr("NOW()"),
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

// HardDelete permanently removes the batch row. Only safe for a batch that
// was never certified on-chain — cascades (ON DELETE CASCADE) wipe its
// honey_batch_certification_requests, blockchain_jobs, honey_batch_certifications,
// and honey_batch_qr_codes rows along with it.
func (r *HoneyBatchRepository) HardDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.HoneyBatch{}, id).Error
}
