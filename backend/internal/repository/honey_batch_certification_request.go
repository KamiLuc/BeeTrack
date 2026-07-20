package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// HoneyBatchCertificationRequestRepository persists the admin-review queue
// that gates blockchain_jobs creation for honey batch certification.
type HoneyBatchCertificationRequestRepository struct {
	db *gorm.DB
}

func NewHoneyBatchCertificationRequestRepository(db *gorm.DB) *HoneyBatchCertificationRequestRepository {
	return &HoneyBatchCertificationRequestRepository{db: db}
}

func (r *HoneyBatchCertificationRequestRepository) Create(ctx context.Context, req *model.HoneyBatchCertificationRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *HoneyBatchCertificationRequestRepository) GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertificationRequest, error) {
	var req model.HoneyBatchCertificationRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// GetPendingForBatch returns batchID's pending request, or nil if none — used
// for the idempotency check before creating a new one.
func (r *HoneyBatchCertificationRequestRepository) GetPendingForBatch(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error) {
	var req model.HoneyBatchCertificationRequest
	err := r.db.WithContext(ctx).
		Where("batch_id = ? AND status = ?", batchID, model.CertificationRequestStatusPending).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *HoneyBatchCertificationRequestRepository) GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error) {
	var req model.HoneyBatchCertificationRequest
	err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("created_at DESC").
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// ListPending returns pending requests oldest-first, for the admin queue,
// along with the total pending count.
func (r *HoneyBatchCertificationRequestRepository) ListPending(ctx context.Context, limit, offset int) ([]*model.HoneyBatchCertificationRequest, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.HoneyBatchCertificationRequest{}).
		Where("status = ?", model.CertificationRequestStatusPending).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var reqs []*model.HoneyBatchCertificationRequest
	err := r.db.WithContext(ctx).
		Where("status = ?", model.CertificationRequestStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reqs).Error
	return reqs, total, err
}

// Approve transitions the request to approved and creates the blockchain_jobs
// row the existing worker will pick up, atomically, so a request is never
// left approved without its job (or vice versa). Fails if the request isn't
// currently pending.
func (r *HoneyBatchCertificationRequestRepository) Approve(ctx context.Context, id, reviewerID int64) (*model.BlockchainJob, error) {
	var job model.BlockchainJob
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var req model.HoneyBatchCertificationRequest
		if err := tx.Where("id = ? AND status = ?", id, model.CertificationRequestStatusPending).
			First(&req).Error; err != nil {
			return err
		}

		job = model.BlockchainJob{
			BatchID:     req.BatchID,
			JobType:     "certify",
			Status:      model.CertificationStatusQueued,
			NextRetryAt: time.Now(),
		}
		if err := tx.Create(&job).Error; err != nil {
			return err
		}

		return tx.Model(&model.HoneyBatchCertificationRequest{}).Where("id = ?", id).
			Updates(map[string]any{
				"status":            model.CertificationRequestStatusApproved,
				"reviewed_by":       reviewerID,
				"reviewed_at":       gorm.Expr("NOW()"),
				"blockchain_job_id": job.ID,
			}).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("certification request not pending: %w", err)
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *HoneyBatchCertificationRequestRepository) Reject(ctx context.Context, id, reviewerID int64, reason string) error {
	result := r.db.WithContext(ctx).Model(&model.HoneyBatchCertificationRequest{}).
		Where("id = ? AND status = ?", id, model.CertificationRequestStatusPending).
		Updates(map[string]any{
			"status":           model.CertificationRequestStatusRejected,
			"rejection_reason": reason,
			"reviewed_by":      reviewerID,
			"reviewed_at":      gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
