package worker

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/model"
)

const maxAttempts = 3

// JobRepository is the durable job queue interface used by BlockchainWorker.
type JobRepository interface {
	ClaimNext(ctx context.Context) (*model.BlockchainJob, error)
	MarkSubmitting(ctx context.Context, id, certificationID int64) error
	MarkSubmitted(ctx context.Context, id int64) error
	MarkConfirmed(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, lastErr string, nextRetryAt time.Time) error
}

// CertificationRepository is the certification history interface used by BlockchainWorker.
type CertificationRepository interface {
	GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error)
	Create(ctx context.Context, c *model.HoneyBatchCertification) error
	UpdateStatus(ctx context.Context, id int64, status model.CertificationStatus, fields map[string]any) error
}

// BatchReader is the batch-lookup interface used by BlockchainWorker.
type BatchReader interface {
	GetByIDIgnoringDeletion(ctx context.Context, id int64) (*model.HoneyBatch, error)
}

// CertifyWriter signs and broadcasts certify() transactions.
type CertifyWriter interface {
	CertifyBatch(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (string, error)
}

// CertificationReader reads on-chain certification records.
type CertificationReader interface {
	GetCertification(ctx context.Context, batchID int64) (*blockchain.CertificationRecord, error)
}

// BlockchainWorker processes the durable blockchain_jobs queue: it's the
// only code path that ever calls CertifyWriter or CertificationReader.
type BlockchainWorker struct {
	jobs            JobRepository
	certifications  CertificationRepository
	batches         BatchReader
	writer          CertifyWriter
	reader          CertificationReader
	chainID         int
	contractAddress string
}

// NewBlockchainWorker creates a BlockchainWorker with the given dependencies.
func NewBlockchainWorker(
	jobs JobRepository,
	certifications CertificationRepository,
	batches BatchReader,
	writer CertifyWriter,
	reader CertificationReader,
	chainID int,
	contractAddress string,
) *BlockchainWorker {
	return &BlockchainWorker{
		jobs:            jobs,
		certifications:  certifications,
		batches:         batches,
		writer:          writer,
		reader:          reader,
		chainID:         chainID,
		contractAddress: contractAddress,
	}
}

// ProcessNextJob claims and processes one runnable job. Returns processed=false
// if no job was currently runnable.
func (w *BlockchainWorker) ProcessNextJob(ctx context.Context) (processed bool, err error) {
	job, err := w.jobs.ClaimNext(ctx)
	if err != nil {
		return false, fmt.Errorf("claim next job: %w", err)
	}
	if job == nil {
		return false, nil
	}

	// A previous worker may have crashed after broadcasting but before
	// marking this job done — if the batch already has a live
	// certification, there's nothing left for this job to do.
	latest, err := w.certifications.GetLatestByBatchID(ctx, job.BatchID)
	if err != nil {
		return true, fmt.Errorf("get latest certification: %w", err)
	}
	if latest != nil && latest.Status.IsLive() {
		if err := w.jobs.MarkConfirmed(ctx, job.ID); err != nil {
			return true, fmt.Errorf("mark job confirmed (idempotency skip): %w", err)
		}
		return true, nil
	}

	batch, err := w.batches.GetByIDIgnoringDeletion(ctx, job.BatchID)
	if err != nil {
		return true, fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return true, w.failJob(ctx, job, nil, errors.New("batch not found"))
	}

	pdfHash, err := decodeHash32(batch.PDFFileHash)
	if err != nil {
		return true, w.failJob(ctx, job, nil, fmt.Errorf("decode pdf hash: %w", err))
	}
	metadataHash, err := decodeHash32(batch.MetadataHash)
	if err != nil {
		return true, w.failJob(ctx, job, nil, fmt.Errorf("decode metadata hash: %w", err))
	}

	cert := &model.HoneyBatchCertification{
		BatchID:         job.BatchID,
		ChainID:         w.chainID,
		ContractAddress: w.contractAddress,
		Status:          model.CertificationStatusSubmitting,
	}
	if err := w.certifications.Create(ctx, cert); err != nil {
		return true, fmt.Errorf("create certification: %w", err)
	}
	if err := w.jobs.MarkSubmitting(ctx, job.ID, cert.ID); err != nil {
		return true, fmt.Errorf("mark job submitting: %w", err)
	}

	txHash, err := w.writer.CertifyBatch(ctx, batch.ID, pdfHash, metadataHash)
	if errors.Is(err, blockchain.ErrAlreadyCertified) {
		return true, w.handleAlreadyCertified(ctx, job, cert, batch.ID)
	}
	if err != nil {
		return true, w.failJob(ctx, job, &cert.ID, err)
	}

	if err := w.certifications.UpdateStatus(ctx, cert.ID, model.CertificationStatusSubmitted, map[string]any{
		"transaction_hash": txHash,
	}); err != nil {
		return true, fmt.Errorf("mark certification submitted: %w", err)
	}
	if err := w.jobs.MarkSubmitted(ctx, job.ID); err != nil {
		return true, fmt.Errorf("mark job submitted: %w", err)
	}
	return true, nil
}

// handleAlreadyCertified treats the contract's "already certified" revert as
// success (HC-BE-25's application-level idempotency safety net): the batch
// is certified, just not by this attempt, so there's no transaction of our
// own to track — the certification is marked confirmed immediately.
func (w *BlockchainWorker) handleAlreadyCertified(ctx context.Context, job *model.BlockchainJob, cert *model.HoneyBatchCertification, batchID int64) error {
	fields := map[string]any{}
	if record, err := w.reader.GetCertification(ctx, batchID); err == nil {
		fields["confirmation_timestamp"] = record.Timestamp
	}
	if err := w.certifications.UpdateStatus(ctx, cert.ID, model.CertificationStatusConfirmed, fields); err != nil {
		return fmt.Errorf("mark certification confirmed (already certified): %w", err)
	}
	if err := w.jobs.MarkConfirmed(ctx, job.ID); err != nil {
		return fmt.Errorf("mark job confirmed (already certified): %w", err)
	}
	return nil
}

// failJob records cause against job and certID (if a certification row was
// already created for this attempt), scheduling a retry with exponential
// backoff until maxAttempts is reached, after which next_retry_at is pushed
// far into the future so ClaimNext stops picking it up automatically — the
// owner must then explicitly retry (HC-BE-24c).
func (w *BlockchainWorker) failJob(ctx context.Context, job *model.BlockchainJob, certID *int64, cause error) error {
	attempt := job.AttemptCount + 1
	var nextRetryAt time.Time
	if attempt >= maxAttempts {
		nextRetryAt = time.Now().AddDate(100, 0, 0)
	} else {
		nextRetryAt = time.Now().Add(backoffDuration(attempt))
	}

	if err := w.jobs.MarkFailed(ctx, job.ID, cause.Error(), nextRetryAt); err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}
	if certID != nil {
		if err := w.certifications.UpdateStatus(ctx, *certID, model.CertificationStatusFailed, nil); err != nil {
			return fmt.Errorf("mark certification failed: %w", err)
		}
	}
	return cause
}

// backoffDuration returns 1s, 2s, 4s, 8s (capped) for attempt 1, 2, 3, 4+.
func backoffDuration(attempt int) time.Duration {
	d := time.Second
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= 8*time.Second {
			return 8 * time.Second
		}
	}
	return d
}

func decodeHash32(s string) ([32]byte, error) {
	var out [32]byte
	b, err := hex.DecodeString(s)
	if err != nil {
		return out, err
	}
	if len(b) != 32 {
		return out, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	copy(out[:], b)
	return out, nil
}
