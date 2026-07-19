package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxAmountGrams = 100_000_000

var (
	ErrInvalidAmount           = errors.New("amount_grams must be greater than 0 and at most 100,000,000")
	ErrHoneyTypeRequired       = errors.New("honey_type is required")
	ErrHoneyTypeTooLong        = fmt.Errorf("honey_type must be at most %d characters", validation.Small.MaxLength())
	ErrInvalidProcessingMethod = errors.New("invalid processing_method")
	ErrPDFRequired             = errors.New("lab PDF file is required")
	ErrBatchNotFound           = errors.New("honey batch not found")
)

// HoneyBatchRepository is the persistence interface for honey batches used by HoneyBatchService.
type HoneyBatchRepository interface {
	CreateWithCertificationJob(ctx context.Context, b *model.HoneyBatch, job *model.BlockchainJob) error
	GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error)
}

// HoneyBatchCertificationRepository is the persistence interface for
// certification history used by HoneyBatchService.
type HoneyBatchCertificationRepository interface {
	GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error)
}

// HoneyBatchService handles business logic for honey batch certification.
type HoneyBatchService struct {
	apiaries       ApiaryMembershipReader
	batches        HoneyBatchRepository
	certifications HoneyBatchCertificationRepository
}

// NewHoneyBatchService creates a HoneyBatchService with the given dependencies.
func NewHoneyBatchService(apiaries ApiaryMembershipReader, batches HoneyBatchRepository, certifications HoneyBatchCertificationRepository) *HoneyBatchService {
	return &HoneyBatchService{apiaries: apiaries, batches: batches, certifications: certifications}
}

// CreateBatchRequest holds the mutable fields for creating a honey batch.
// PDFFilePath is a local path to the already-saved lab PDF (used for
// hashing); LabPDFURL is the URL persisted on the batch record.
type CreateBatchRequest struct {
	GatheringDate        time.Time
	AmountGrams          int64
	ProcessingMethod     string
	HoneyType            string
	PDFFilePath          string
	LabPDFURL            string
	RequestCertification bool
}

func validateCreateBatchRequest(req CreateBatchRequest) error {
	if req.AmountGrams <= 0 || req.AmountGrams > maxAmountGrams {
		return ErrInvalidAmount
	}
	if req.HoneyType == "" {
		return ErrHoneyTypeRequired
	}
	if validation.TooLong(req.HoneyType, validation.Small) {
		return ErrHoneyTypeTooLong
	}
	if !model.IsValidProcessingMethod(req.ProcessingMethod) {
		return ErrInvalidProcessingMethod
	}
	if req.PDFFilePath == "" {
		return ErrPDFRequired
	}
	return nil
}

// CreateBatch validates req, hashes the lab PDF, and persists a new honey
// batch. If req.RequestCertification is true, a "certify" job is enqueued in
// the same transaction as the batch insert — no blockchain call happens
// here. If false, the batch is created with no certification attempt at
// all; the owner can request one later.
func (s *HoneyBatchService) CreateBatch(ctx context.Context, userID, apiaryID int64, req CreateBatchRequest) (*model.HoneyBatch, error) {
	if err := validateCreateBatchRequest(req); err != nil {
		return nil, err
	}
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	pdfHash, err := blockchain.SHA256File(req.PDFFilePath)
	if err != nil {
		return nil, fmt.Errorf("hash pdf: %w", err)
	}

	token, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("generate verification token: %w", err)
	}

	batch := &model.HoneyBatch{
		UserID:            userID,
		ApiaryID:          apiaryID,
		VerificationToken: token.String(),
		GatheringDate:     req.GatheringDate,
		AmountGrams:       req.AmountGrams,
		ProcessingMethod:  req.ProcessingMethod,
		HoneyType:         req.HoneyType,
		LabPDFURL:         req.LabPDFURL,
		PDFFileHash:       hex.EncodeToString(pdfHash[:]),
	}

	var job *model.BlockchainJob
	if req.RequestCertification {
		job = &model.BlockchainJob{
			JobType:     "certify",
			Status:      model.CertificationStatusQueued,
			NextRetryAt: time.Now(),
		}
	}

	if err := s.batches.CreateWithCertificationJob(ctx, batch, job); err != nil {
		return nil, fmt.Errorf("create honey batch: %w", err)
	}

	return batch, nil
}

// BatchVerification is a batch plus its current certification state.
// Certification is nil if the batch has never had certification requested.
type BatchVerification struct {
	Batch         *model.HoneyBatch
	Certification *model.HoneyBatchCertification
}

// GetBatchWithVerification looks up a batch by its public verification token
// and returns it together with its latest certification (if any). It only
// reads from the DB — kept fresh by the worker's confirmation loop — never
// making a live blockchain call itself.
func (s *HoneyBatchService) GetBatchWithVerification(ctx context.Context, token string) (*BatchVerification, error) {
	batch, err := s.batches.GetByVerificationToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}

	cert, err := s.certifications.GetLatestByBatchID(ctx, batch.ID)
	if err != nil {
		return nil, fmt.Errorf("get latest certification: %w", err)
	}

	return &BatchVerification{Batch: batch, Certification: cert}, nil
}
