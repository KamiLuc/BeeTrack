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
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/internal/validation"
	"github.com/google/uuid"
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
	ErrBatchAlreadyCertified   = errors.New("honey batch already has a live certification")
	ErrBatchHasNoPDF           = errors.New("honey batch has no lab PDF; a PDF is required to certify")
	ErrBatchLocked             = errors.New("honey batch can no longer be edited once certification has been requested")
	ErrCertificationRequestPending = errors.New("a certification request for this batch is already pending admin review")
)

// HoneyBatchRepository is the persistence interface for honey batches used by HoneyBatchService.
type HoneyBatchRepository interface {
	CreateWithCertificationRequest(ctx context.Context, b *model.HoneyBatch, certRequest *model.HoneyBatchCertificationRequest) error
	GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error)
	GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.HoneyBatch, error)
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	UpdateFields(ctx context.Context, batch *model.HoneyBatch) error
	SoftDelete(ctx context.Context, id int64) error
	HardDelete(ctx context.Context, id int64) error
}

// HoneyBatchCertificationRequestRepository is the persistence interface for
// the admin-review queue that gates blockchain_jobs creation.
type HoneyBatchCertificationRequestRepository interface {
	Create(ctx context.Context, req *model.HoneyBatchCertificationRequest) error
	GetPendingForBatch(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error)
	GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error)
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

// BlockchainJobRepository is the persistence interface for enqueuing certification jobs used by HoneyBatchService.
type BlockchainJobRepository interface {
	Create(ctx context.Context, j *model.BlockchainJob) error
	HasPendingJob(ctx context.Context, batchID int64) (bool, error)
}

// HoneyBatchService handles business logic for honey batch certification.
type HoneyBatchService struct {
	batches        HoneyBatchRepository
	certifications HoneyBatchCertificationRepository
	certRequests   HoneyBatchCertificationRequestRepository
	qrCodes        HoneyBatchQRCodeRepository
	jobs           BlockchainJobRepository
	publicBaseURL  string // base URL this service's own /verify/{token} page is reachable at; see VerificationURL.
	pdfStoragePath string
}

// NewHoneyBatchService creates a HoneyBatchService with the given dependencies. Lab PDFs are stored under pdfStoragePath.
func NewHoneyBatchService(batches HoneyBatchRepository, certifications HoneyBatchCertificationRepository, certRequests HoneyBatchCertificationRequestRepository, qrCodes HoneyBatchQRCodeRepository, jobs BlockchainJobRepository, publicBaseURL, pdfStoragePath string) *HoneyBatchService {
	return &HoneyBatchService{batches: batches, certifications: certifications, certRequests: certRequests, qrCodes: qrCodes, jobs: jobs, publicBaseURL: publicBaseURL, pdfStoragePath: pdfStoragePath}
}

// CreateBatchRequest holds the mutable fields for creating a honey batch, including the raw lab PDF upload.
type CreateBatchRequest struct {
	GatheringDate        time.Time
	AmountGrams          int64
	ProcessingMethod     string
	HoneyType            string
	PDFMimeType          string
	PDFData              []byte
	PDFFilename          string
	RequestCertification bool
}

// validateCreateBatchRequest checks req's fields. The lab PDF is only
// required when RequestCertification is true — a batch can be created
// without one and certified later, at which point a PDF must be attached.
// If a PDF is provided at all, it's always validated regardless of the flag.
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
		if req.RequestCertification {
			return ErrPDFRequired
		}
		return nil
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
// new honey batch. If req.RequestCertification is true, a pending
// certification request is created in the same transaction as the batch
// insert, awaiting admin approval before any blockchain_jobs row (and thus
// any chain interaction) exists. If false, the batch is created with no
// certification attempt at all; the owner can request one later.
func (s *HoneyBatchService) CreateBatch(ctx context.Context, userID int64, req CreateBatchRequest) (*model.HoneyBatch, error) {
	if err := validateCreateBatchRequest(req); err != nil {
		return nil, err
	}

	var filename, pdfHashHex string
	if len(req.PDFData) > 0 {
		if err := os.MkdirAll(s.pdfStoragePath, 0o755); err != nil {
			return nil, fmt.Errorf("create pdf storage dir: %w", err)
		}
		filename = uuid.New().String() + ".pdf"
		pdfPath := filepath.Join(s.pdfStoragePath, filename)
		if err := os.WriteFile(pdfPath, req.PDFData, 0o644); err != nil {
			return nil, fmt.Errorf("write pdf: %w", err)
		}

		pdfHash, err := blockchain.SHA256File(pdfPath)
		if err != nil {
			_ = os.Remove(pdfPath)
			return nil, fmt.Errorf("hash pdf: %w", err)
		}
		pdfHashHex = hex.EncodeToString(pdfHash[:])
	}

	token, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("generate verification token: %w", err)
	}

	var pdfFilename string
	if filename != "" {
		pdfFilename = req.PDFFilename
	}

	batch := &model.HoneyBatch{
		UserID:            userID,
		VerificationToken: token.String(),
		GatheringDate:     req.GatheringDate,
		AmountGrams:       req.AmountGrams,
		ProcessingMethod:  req.ProcessingMethod,
		HoneyType:         req.HoneyType,
		LabPDFURL:         filename,
		PDFFilename:       pdfFilename,
		PDFFileHash:       pdfHashHex,
	}

	var certRequest *model.HoneyBatchCertificationRequest
	if req.RequestCertification {
		certRequest = &model.HoneyBatchCertificationRequest{
			RequestedBy: userID,
			Status:      model.CertificationRequestStatusPending,
		}
	}

	if err := s.batches.CreateWithCertificationRequest(ctx, batch, certRequest); err != nil {
		if filename != "" {
			_ = os.Remove(s.FilePath(filename))
		}
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
// CertificationRequest is the latest admin-review request regardless of
// status (nil if none was ever created) — distinct from Certification,
// which only reflects state once a request has been approved and the
// worker has taken over.
type BatchVerification struct {
	Batch                 *model.HoneyBatch
	Certification         *model.HoneyBatchCertification
	CertificationRequest  *model.HoneyBatchCertificationRequest
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

// withPendingFallback returns cert unchanged if non-nil. Otherwise, since a
// queued/in-flight job has no certification row until the worker claims it
// (see RetryCertification's HasPendingJob check), it checks for one and
// synthesizes a "queued" placeholder so owner-facing reads (GetBatch,
// ListBatches) reflect an in-progress attempt instead of looking identical
// to "never certified".
func (s *HoneyBatchService) withPendingFallback(ctx context.Context, batchID int64, cert *model.HoneyBatchCertification) (*model.HoneyBatchCertification, error) {
	if cert != nil {
		return cert, nil
	}
	pending, err := s.jobs.HasPendingJob(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("check pending job: %w", err)
	}
	if !pending {
		return nil, nil
	}
	return &model.HoneyBatchCertification{Status: model.CertificationStatusQueued}, nil
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
	cert, err = s.withPendingFallback(ctx, batch.ID, cert)
	if err != nil {
		return nil, err
	}
	certRequest, err := s.certRequests.GetLatestByBatchID(ctx, batch.ID)
	if err != nil {
		return nil, fmt.Errorf("get latest certification request: %w", err)
	}
	return &BatchVerification{Batch: batch, Certification: cert, CertificationRequest: certRequest}, nil
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
		cert, err = s.withPendingFallback(ctx, b.ID, cert)
		if err != nil {
			return nil, 0, err
		}
		certRequest, err := s.certRequests.GetLatestByBatchID(ctx, b.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("get latest certification request: %w", err)
		}
		items[i] = BatchVerification{Batch: b, Certification: cert, CertificationRequest: certRequest}
	}
	return items, total, nil
}

// UpdateBatchRequest holds the editable fields for an existing batch,
// available before certification is ever requested. PDFData is optional —
// when non-empty it replaces the batch's lab PDF; when empty and RemovePDF
// is true, the existing PDF (if any) is cleared; when empty and RemovePDF is
// false, the existing PDF is left untouched. PDFData takes precedence over
// RemovePDF if both are set.
type UpdateBatchRequest struct {
	GatheringDate    time.Time
	AmountGrams      int64
	ProcessingMethod string
	HoneyType        string
	PDFData          []byte
	PDFMimeType      string
	PDFFilename      string
	RemovePDF        bool
}

// hasCertificationAttempt reports whether batchID has any certification
// history, a queued/in-flight job, or a pending review request — once true,
// the batch is locked from further edits, since its metadata hash may
// already be part of a submitted, on-chain, or awaiting-approval attempt.
func (s *HoneyBatchService) hasCertificationAttempt(ctx context.Context, batchID int64) (bool, error) {
	cert, err := s.certifications.GetLatestByBatchID(ctx, batchID)
	if err != nil {
		return false, fmt.Errorf("get latest certification: %w", err)
	}
	if cert != nil {
		return true, nil
	}
	pending, err := s.jobs.HasPendingJob(ctx, batchID)
	if err != nil {
		return false, fmt.Errorf("check pending job: %w", err)
	}
	if pending {
		return true, nil
	}
	pendingRequest, err := s.certRequests.GetPendingForBatch(ctx, batchID)
	if err != nil {
		return false, fmt.Errorf("check pending certification request: %w", err)
	}
	return pendingRequest != nil, nil
}

// UpdateBatch validates and overwrites a batch's gathering date, amount,
// processing method, and honey type — the fields editable before
// certification is ever requested. Recomputes metadata_hash so a later
// certification always hashes the batch's current field values, never
// stale ones from creation time. Rejects with ErrBatchLocked once any
// certification attempt exists (see hasCertificationAttempt).
func (s *HoneyBatchService) UpdateBatch(ctx context.Context, userID, batchID int64, req UpdateBatchRequest) (*model.HoneyBatch, error) {
	if req.AmountGrams <= 0 || req.AmountGrams > maxAmountGrams {
		return nil, ErrInvalidAmount
	}
	if req.HoneyType == "" {
		return nil, ErrHoneyTypeRequired
	}
	if validation.TooLong(req.HoneyType, validation.Small) {
		return nil, ErrHoneyTypeTooLong
	}
	if !model.IsValidProcessingMethod(req.ProcessingMethod) {
		return nil, ErrInvalidProcessingMethod
	}
	if len(req.PDFData) > 0 {
		if req.PDFMimeType != "application/pdf" {
			return nil, ErrInvalidPDFType
		}
		if len(req.PDFData) > maxLabPDFBytes {
			return nil, ErrPDFTooLarge
		}
	}

	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return nil, err
	}

	locked, err := s.hasCertificationAttempt(ctx, batch.ID)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, ErrBatchLocked
	}

	oldPDFFilename := batch.LabPDFURL
	pdfChanged := len(req.PDFData) > 0 || req.RemovePDF
	switch {
	case len(req.PDFData) > 0:
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
		batch.LabPDFURL = filename
		batch.PDFFilename = req.PDFFilename
		batch.PDFFileHash = hex.EncodeToString(pdfHash[:])
	case req.RemovePDF:
		batch.LabPDFURL = ""
		batch.PDFFilename = ""
		batch.PDFFileHash = ""
	}

	batch.GatheringDate = req.GatheringDate
	batch.AmountGrams = req.AmountGrams
	batch.ProcessingMethod = req.ProcessingMethod
	batch.HoneyType = req.HoneyType
	metadataHash := blockchain.CanonicalMetadataHash(batch)
	batch.MetadataHash = hex.EncodeToString(metadataHash[:])

	if err := s.batches.UpdateFields(ctx, batch); err != nil {
		if len(req.PDFData) > 0 {
			_ = os.Remove(s.FilePath(batch.LabPDFURL))
		}
		return nil, fmt.Errorf("update batch: %w", err)
	}
	if pdfChanged && oldPDFFilename != "" {
		_ = os.Remove(s.FilePath(oldPDFFilename))
	}
	return batch, nil
}

// DeleteBatch removes a batch owned by userID. A batch that was never
// approved for on-chain certification (no certification request ever reached
// blockchain_jobs) is hard-deleted — its lab PDF is removed from disk and the
// row's cascading deletes wipe any dangling certification_requests, so it
// can't linger in the admin panel pointing at a PDF that no longer exists.
// A batch that was approved (certification in flight or already confirmed
// on-chain) is only soft-deleted, preserving that audit trail untouched.
func (s *HoneyBatchService) DeleteBatch(ctx context.Context, userID, batchID int64) error {
	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return err
	}

	certRequest, err := s.certRequests.GetLatestByBatchID(ctx, batchID)
	if err != nil {
		return fmt.Errorf("get latest certification request: %w", err)
	}
	everApprovedForCertification := certRequest != nil && certRequest.BlockchainJobID != nil

	if !everApprovedForCertification {
		if err := s.batches.HardDelete(ctx, batchID); err != nil {
			return fmt.Errorf("delete batch: %w", err)
		}
		if batch.PDFFilename != "" {
			_ = os.Remove(s.FilePath(batch.PDFFilename))
		}
		return nil
	}

	if err := s.batches.SoftDelete(ctx, batchID); err != nil {
		return fmt.Errorf("delete batch: %w", err)
	}
	return nil
}

// RetryCertification submits a batch owned by userID for admin certification
// review — first-time request if none was ever made, or a resubmission if
// the latest attempt failed/reverted/was rejected. Approval (by an admin,
// see CertificationReviewService) is what actually enqueues the
// blockchain_jobs row the worker picks up. Rejects with
// ErrBatchAlreadyCertified if a live certification or in-flight job already
// exists, or ErrCertificationRequestPending if a review request is still
// awaiting a decision.
func (s *HoneyBatchService) RetryCertification(ctx context.Context, userID, batchID int64) error {
	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return err
	}
	if batch.PDFFileHash == "" {
		return ErrBatchHasNoPDF
	}
	cert, err := s.certifications.GetLatestByBatchID(ctx, batch.ID)
	if err != nil {
		return fmt.Errorf("get latest certification: %w", err)
	}
	if cert != nil && cert.Status.IsLive() {
		return ErrBatchAlreadyCertified
	}
	// A queued/in-flight job may not have a certification row yet (the worker
	// only creates one once it claims the job), so the cert.IsLive() check
	// above can't catch a duplicate retry click during that window.
	if pending, err := s.jobs.HasPendingJob(ctx, batch.ID); err != nil {
		return fmt.Errorf("check pending job: %w", err)
	} else if pending {
		return ErrBatchAlreadyCertified
	}
	if pendingRequest, err := s.certRequests.GetPendingForBatch(ctx, batch.ID); err != nil {
		return fmt.Errorf("check pending certification request: %w", err)
	} else if pendingRequest != nil {
		return ErrCertificationRequestPending
	}
	req := &model.HoneyBatchCertificationRequest{
		BatchID:     batch.ID,
		RequestedBy: userID,
		Status:      model.CertificationRequestStatusPending,
	}
	if err := s.certRequests.Create(ctx, req); err != nil {
		if errors.Is(err, repository.ErrPendingRequestExists) {
			return ErrCertificationRequestPending
		}
		return fmt.Errorf("create certification request: %w", err)
	}
	return nil
}

// GetBatchPDF returns the absolute file path of a batch's lab PDF, verifying ownership.
func (s *HoneyBatchService) GetBatchPDF(ctx context.Context, userID, batchID int64) (string, error) {
	batch, err := s.ownedBatch(ctx, userID, batchID)
	if err != nil {
		return "", err
	}
	if batch.LabPDFURL == "" {
		return "", ErrBatchHasNoPDF
	}
	return s.FilePath(batch.LabPDFURL), nil
}

// GetBatchPDFForAdmin returns the absolute file path of a batch's lab PDF for
// an admin reviewer, bypassing the ownership check — callers must gate this
// with RequireAdmin themselves.
func (s *HoneyBatchService) GetBatchPDFForAdmin(ctx context.Context, batchID int64) (string, error) {
	batch, err := s.batches.GetByID(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return "", ErrBatchNotFound
	}
	if batch.LabPDFURL == "" {
		return "", ErrBatchHasNoPDF
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

	data := s.VerificationURL(batch.VerificationToken)
	if err := s.qrCodes.Create(ctx, &model.HoneyBatchQRCode{BatchID: batchID, QRCodeData: data}); err != nil {
		return "", fmt.Errorf("create qr code: %w", err)
	}
	return data, nil
}

// VerificationURL builds the public, token-scoped verification page URL for token — the same URL encoded in the batch's QR code, served directly by this backend's own VerifyPage handler (see honey_batch_verify_page.go).
func (s *HoneyBatchService) VerificationURL(token string) string {
	return s.publicBaseURL + "/verify/" + token
}
