package repository

import (
	"context"
	"errors"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlockchainJobRepository persists the durable blockchain job queue.
type BlockchainJobRepository struct {
	db *gorm.DB
}

// NewBlockchainJobRepository creates a new BlockchainJobRepository backed by db.
func NewBlockchainJobRepository(db *gorm.DB) *BlockchainJobRepository {
	return &BlockchainJobRepository{db: db}
}

// Create inserts a new job.
func (r *BlockchainJobRepository) Create(ctx context.Context, j *model.BlockchainJob) error {
	return r.db.WithContext(ctx).Create(j).Error
}

// HasPendingJob reports whether batchID already has a job that's still in
// flight (queued, submitting, submitted, or pending_confirmation) — used to
// stop RetryCertification from enqueuing a second job on top of one the
// worker hasn't claimed/resolved yet.
func (r *BlockchainJobRepository) HasPendingJob(ctx context.Context, batchID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("batch_id = ? AND status IN ?", batchID, []model.CertificationStatus{
			model.CertificationStatusQueued,
			model.CertificationStatusSubmitting,
			model.CertificationStatusSubmitted,
			model.CertificationStatusPendingConfirmation,
		}).
		Count(&count).Error
	return count > 0, err
}

// ClaimNext atomically selects and claims one runnable job (queued or failed,
// due for retry), so multiple worker instances can run against the same queue
// safely. Claiming immediately transitions the job to "submitting" within the
// same transaction that holds the row lock, so a concurrent ClaimNext can
// never observe (and re-claim) the same job. Returns nil, nil if no job is
// currently runnable.
func (r *BlockchainJobRepository) ClaimNext(ctx context.Context) (*model.BlockchainJob, error) {
	var job model.BlockchainJob
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status IN (?, ?) AND next_retry_at <= NOW()",
				model.CertificationStatusQueued, model.CertificationStatusFailed).
			Order("created_at ASC").
			Limit(1).
			First(&job).Error
		if err != nil {
			return err
		}
		return tx.Model(&job).Updates(map[string]any{
			"status":     model.CertificationStatusSubmitting,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job.Status = model.CertificationStatusSubmitting
	return &job, nil
}

// MarkSubmitting links the job to the certification row created for this
// attempt, keeping its status at "submitting".
func (r *BlockchainJobRepository) MarkSubmitting(ctx context.Context, id, certificationID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":           model.CertificationStatusSubmitting,
			"certification_id": certificationID,
			"updated_at":       gorm.Expr("NOW()"),
		}).Error
}

// MarkSubmitted transitions the job to "submitted" after a successful broadcast.
func (r *BlockchainJobRepository) MarkSubmitted(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     model.CertificationStatusSubmitted,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// MarkFailed increments the job's attempt count, records lastErr, and
// schedules the next retry at nextRetryAt.
func (r *BlockchainJobRepository) MarkFailed(ctx context.Context, id int64, lastErr string, nextRetryAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        model.CertificationStatusFailed,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"last_error":    lastErr,
			"next_retry_at": nextRetryAt,
			"updated_at":    gorm.Expr("NOW()"),
		}).Error
}

// MarkConfirmed transitions the job directly to "confirmed", bypassing
// submitted/pending_confirmation — used only when the worker discovers the
// batch was already certified by a different attempt, so this job has
// nothing left to broadcast or track.
func (r *BlockchainJobRepository) MarkConfirmed(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     model.CertificationStatusConfirmed,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// MarkPendingConfirmation transitions the job to "pending_confirmation" —
// mined but not yet past RequiredConfirmations.
func (r *BlockchainJobRepository) MarkPendingConfirmation(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     model.CertificationStatusPendingConfirmation,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// MarkReverted transitions the job to "reverted" — terminal, not retried,
// since a revert is a semantic on-chain rejection, not a transient failure.
func (r *BlockchainJobRepository) MarkReverted(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     model.CertificationStatusReverted,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// SweepStuckSubmitting resets jobs stuck in "submitting" for longer than
// olderThan back to "failed" with an immediate next_retry_at. A job can get
// stuck here if the worker crashes between claiming it and recording the
// outcome of its writer call — ClaimNext never picks up "submitting" jobs on
// its own, so without this sweep such a job would be stranded forever.
// Returns the number of jobs reset.
func (r *BlockchainJobRepository) SweepStuckSubmitting(ctx context.Context, olderThan time.Duration) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.BlockchainJob{}).
		Where("status = ? AND updated_at < ?", model.CertificationStatusSubmitting, time.Now().Add(-olderThan)).
		Updates(map[string]any{
			"status":        model.CertificationStatusFailed,
			"last_error":    "reset by stuck-submitting sweep",
			"next_retry_at": time.Now(),
			"updated_at":    gorm.Expr("NOW()"),
		})
	return result.RowsAffected, result.Error
}

// ListPendingConfirmation returns jobs whose certification has been broadcast
// but not yet confirmed on-chain, for the confirmation-polling loop.
func (r *BlockchainJobRepository) ListPendingConfirmation(ctx context.Context) ([]*model.BlockchainJob, error) {
	var jobs []*model.BlockchainJob
	err := r.db.WithContext(ctx).
		Where("status IN (?, ?)", model.CertificationStatusSubmitted, model.CertificationStatusPendingConfirmation).
		Find(&jobs).Error
	return jobs, err
}

// ListStuckConfirming returns submitted/pending_confirmation jobs that haven't
// changed state in longer than olderThan — e.g. a transaction dropped from the
// mempool or an RPC endpoint that stopped returning it, which PollSubmittedJobs
// alone would otherwise wait on forever.
func (r *BlockchainJobRepository) ListStuckConfirming(ctx context.Context, olderThan time.Duration) ([]*model.BlockchainJob, error) {
	var jobs []*model.BlockchainJob
	err := r.db.WithContext(ctx).
		Where("status IN (?, ?) AND updated_at < ?",
			model.CertificationStatusSubmitted, model.CertificationStatusPendingConfirmation,
			time.Now().Add(-olderThan)).
		Find(&jobs).Error
	return jobs, err
}
