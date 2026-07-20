package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxAmountGrams = 100_000_000

const maxLabPDFBytes = 10 * 1024 * 1024

var (
	ErrInvalidAmount           = errors.New("amount_grams must be greater than 0 and at most 100,000,000")
	ErrHoneyTypeRequired       = errors.New("honey_type is required")
	ErrHoneyTypeTooLong        = fmt.Errorf("honey_type must be at most %d characters", validation.Small.MaxLength())
	ErrInvalidProcessingMethod = errors.New("invalid processing_method")
	ErrPDFRequired             = errors.New("lab PDF file is required")
	ErrInvalidPDFType          = errors.New("lab PDF must have content type application/pdf")
	ErrPDFTooLarge             = fmt.Errorf("lab PDF exceeds %d MB limit", maxLabPDFBytes/(1024*1024))
	ErrBatchNotFound           = errors.New("honey batch not found")
	ErrBatchNotCertified       = errors.New("honey batch does not have a confirmed certification yet")
)

// HoneyBatchRepository is the persistence interface for honey batches used by HoneyBatchService.
type HoneyBatchRepository interface {
	CreateWithCertificationJob(ctx context.Context, b *model.HoneyBatch, job *model.BlockchainJob) error
	GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error)
	GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.HoneyBatch, error)
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	UpdateNotes(ctx context.Context, id int64, honeyType string) error
	SoftDelete(ctx context.Context, id int64) error
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
	pdfStoragePath string
}

// NewHoneyBatchService creates a HoneyBatchService with the given dependencies. Lab PDFs are stored under pdfStoragePath.
func NewHoneyBatchService(apiaries ApiaryMembershipReader, batches HoneyBatchRepository, certifications HoneyBatchCertificationRepository, qrCodes HoneyBatchQRCodeRepository, appURL, pdfStoragePath string) *HoneyBatchService {
	return &HoneyBatchService{apiaries: apiaries, batches: batches, certifications: certifications, qrCodes: qrCodes, appURL: appURL, pdfStoragePath: pdfStoragePath}
}

// CreateBatchRequest holds the mutable fields for creating a honey batch, including the raw lab PDF upload.
type CreateBatchRequest struct {
	GatheringDate        time.Time
	AmountGrams          int64
	ProcessingMethod     string
	HoneyType            string
	PDFMimeType          string
	PDFData              []byte
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
	if len(req.PDFData) == 0 {
		return ErrPDFRequired
	}
	if req.PDFMimeType != "application/pdf" {
		return ErrInvalidPDFType
	}
	if len(req.PDFData) > maxLabPDFBytes {
		return ErrPDFTooLarge
	}
	return nil
}

// CreateBatch validates req, stores and hashes the lab PDF, and persists a
// new honey batch. If req.RequestCertification is true, a "certify" job is
// enqueued in the same transaction as the batch insert — no blockchain call
// happens here. If false, the batch is created with no certification
// attempt at all; the owner can request one later.
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

	if err := os.MkdirAll(s.pdfStoragePath, 0o755); err != nil {
		return nil, fmt.Errorf("create pdf storage dir: %w", err)
	}
	filename := uuid.New().String() + ".pdf"
	pdfPath := filepath.Join(s.pdfStoragePath, filename)
	if err := os.WriteFile(pdfPath, req.PDFData, 0o644); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}

	pdfHash, err := blockchain.SHA256File(pdfPath)
	if err != nil {
		_ = os.Remove(pdfPath)
		return nil, fmt.Errorf("hash pdf: %w", err)
	}

	token, err := uuid.NewRandom()
	if err != nil {
		_ = os.Remove(pdfPath)
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
		LabPDFURL:         filename,
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
		_ = os.Remove(pdfPath)
		return nil, fmt.Errorf("create honey batch: %w", err)
	}

	return batch, nil
}

// FilePath returns the absolute path for a stored lab PDF filename.
func (s *HoneyBatchService) FilePath(filename string) string {
	return filepath.Join(s.pdfStoragePath, filename)
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

// ownedBatch fetches a batch and verifies it belongs to userID.
func (s *HoneyBatchService) ownedBatch(ctx context.Context, userID, batchID int64) (*model.HoneyBatch, error) {
	batch, err := s.batches.GetByID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("get batch: %w", err)
	}
	if batch == nil || batch.UserID != userID {
		return nil, ErrBatchNotFound
	}
	return batch, nil
}

// GetBatch returns a single batch owned by userID, together with its latest certification (if any).
func (s *HoneyBatchService) GetBatch(ctx context.Context, userID, batchID int64) (*BatchVerification, error) {
	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return nil, err
	}
	cert, err := s.certifications.GetLatestByBatchID(ctx, batch.ID)
	if err != nil {
		return nil, fmt.Errorf("get latest certification: %w", err)
	}
	return &BatchVerification{Batch: batch, Certification: cert}, nil
}

// ListBatches returns a paginated slice of userID's batches (each paired with its latest certification) and the total count.
func (s *HoneyBatchService) ListBatches(ctx context.Context, userID int64, limit, offset int) ([]BatchVerification, int64, error) {
	total, err := s.batches.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count batches: %w", err)
	}
	batches, err := s.batches.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list batches: %w", err)
	}

	items := make([]BatchVerification, len(batches))
	for i, b := range batches {
		cert, err := s.certifications.GetLatestByBatchID(ctx, b.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("get latest certification: %w", err)
		}
		items[i] = BatchVerification{Batch: b, Certification: cert}
	}
	return items, total, nil
}

// UpdateHoneyType validates and overwrites a batch's honey_type — the only user-editable field.
func (s *HoneyBatchService) UpdateHoneyType(ctx context.Context, userID, batchID int64, honeyType string) (*model.HoneyBatch, error) {
	if honeyType == "" {
		return nil, ErrHoneyTypeRequired
	}
	if validation.TooLong(honeyType, validation.Small) {
		return nil, ErrHoneyTypeTooLong
	}
	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return nil, err
	}
	if err := s.batches.UpdateNotes(ctx, batchID, honeyType); err != nil {
		return nil, fmt.Errorf("update honey type: %w", err)
	}
	batch.HoneyType = honeyType
	return batch, nil
}

// DeleteBatch soft-deletes a batch owned by userID. The on-chain certification, if any, is untouched.
func (s *HoneyBatchService) DeleteBatch(ctx context.Context, userID, batchID int64) error {
	if _, err := s.ownedBatch(ctx, userID, batchID); err != nil {
		return err
	}
	if err := s.batches.SoftDelete(ctx, batchID); err != nil {
		return fmt.Errorf("delete batch: %w", err)
	}
	return nil
}

// GetBatchPDF returns the absolute file path of a batch's lab PDF, verifying ownership.
func (s *HoneyBatchService) GetBatchPDF(ctx context.Context, userID, batchID int64) (string, error) {
	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return "", err
	}
	return s.FilePath(batch.LabPDFURL), nil
}

// GetBatchPDFByToken returns the absolute file path of a batch's lab PDF via its public verification token. Requires a confirmed certification.
func (s *HoneyBatchService) GetBatchPDFByToken(ctx context.Context, token string) (string, error) {
	batch, err := s.batches.GetByVerificationToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return "", ErrBatchNotFound
	}
	cert, err := s.certifications.GetLatestByBatchID(ctx, batch.ID)
	if err != nil {
		return "", fmt.Errorf("get latest certification: %w", err)
	}
	if cert == nil || cert.Status != model.CertificationStatusConfirmed {
		return "", ErrBatchNotCertified
	}
	return s.FilePath(batch.LabPDFURL), nil
}

// GenerateQRCodeDataByToken resolves a public verification token to its batch and returns its QR code data (see GenerateQRCodeData).
func (s *HoneyBatchService) GenerateQRCodeDataByToken(ctx context.Context, token string) (string, error) {
	batch, err := s.batches.GetByVerificationToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return "", ErrBatchNotFound
	}
	return s.GenerateQRCodeData(ctx, batch.ID)
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
