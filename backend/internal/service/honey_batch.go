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
	ErrBatchNotCertified       = errors.New("honey batch does not have a confirmed certification yet")
)

// HoneyBatchRepository is the persistence interface for honey batches used by HoneyBatchService.
type HoneyBatchRepository interface {
	CreateWithCertificationJob(ctx context.Context, b *model.HoneyBatch, job *model.BlockchainJob) error
	GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error)
	GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error)
}

// HoneyBatchCertificationRepository is the persistence interface for
// certification history used by HoneyBatchService.
type HoneyBatchCertificationRepository interface {
	GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error)
}

// HoneyBatchQRCodeRepository is the persistence interface for QR code data used by HoneyBatchService.
type HoneyBatchQRCodeRepository interface {
	GetByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchQRCode, error)
	Create(ctx context.Context, q *model.HoneyBatchQRCode) error
}

// HoneyBatchService handles business logic for honey batch certification.
type HoneyBatchService struct {
	apiaries       ApiaryMembershipReader
	batches        HoneyBatchRepository
	certifications HoneyBatchCertificationRepository
	qrCodes        HoneyBatchQRCodeRepository
	appURL         string
}

// NewHoneyBatchService creates a HoneyBatchService with the given dependencies.
func NewHoneyBatchService(apiaries ApiaryMembershipReader, batches HoneyBatchRepository, certifications HoneyBatchCertificationRepository, qrCodes HoneyBatchQRCodeRepository, appURL string) *HoneyBatchService {
	return &HoneyBatchService{apiaries: apiaries, batches: batches, certifications: certifications, qrCodes: qrCodes, appURL: appURL}
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

// GenerateQRCodeData returns the QR code data URL for batchID, generating and persisting it on first call. Requires a confirmed certification.
func (s *HoneyBatchService) GenerateQRCodeData(ctx context.Context, batchID int64) (string, error) {
	existing, err := s.qrCodes.GetByBatchID(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("get qr code: %w", err)
	}
	if existing != nil {
		return existing.QRCodeData, nil
	}

	batch, err := s.batches.GetByID(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return "", ErrBatchNotFound
	}

	cert, err := s.certifications.GetLatestByBatchID(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("get latest certification: %w", err)
	}
	if cert == nil || cert.Status != model.CertificationStatusConfirmed {
		return "", ErrBatchNotCertified
	}

	data := s.appURL + "/verify/" + batch.VerificationToken
	if err := s.qrCodes.Create(ctx, &model.HoneyBatchQRCode{BatchID: batchID, QRCodeData: data}); err != nil {
		return "", fmt.Errorf("create qr code: %w", err)
	}
	return data, nil
}
