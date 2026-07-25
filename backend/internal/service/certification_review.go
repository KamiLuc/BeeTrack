package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrCertificationRequestNotFound   = errors.New("certification request not found")
	ErrCertificationRequestNotPending = errors.New("certification request is not pending review")
	// ErrCertificationRequestNotFailed guards Delete — only a request whose
	// blockchain attempt definitively failed/reverted can be cleared this
	// way, so a pending, in-flight, or confirmed request is never lost.
	ErrCertificationRequestNotFailed = errors.New("certification request's blockchain attempt did not fail")
)

// CertificationRequestStore is the persistence interface for the admin
// certification-request review queue.
type CertificationRequestStore interface {
	List(ctx context.Context, status, keyword, sortDir string, limit, offset int) ([]*model.HoneyBatchCertificationRequestDetail, int64, error)
	GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertificationRequestDetail, error)
	Approve(ctx context.Context, id, reviewerID int64) (*model.BlockchainJob, error)
	Reject(ctx context.Context, id, reviewerID int64, reason string) error
	Delete(ctx context.Context, id int64) error
}

type CertificationReviewService struct {
	requests CertificationRequestStore
}

func NewCertificationReviewService(requests CertificationRequestStore) *CertificationReviewService {
	return &CertificationReviewService{requests: requests}
}

func (s *CertificationReviewService) List(ctx context.Context, status, keyword, sortDir string, limit, offset int) ([]*model.HoneyBatchCertificationRequestDetail, int64, error) {
	return s.requests.List(ctx, status, keyword, sortDir, limit, offset)
}

func (s *CertificationReviewService) Get(ctx context.Context, requestID int64) (*model.HoneyBatchCertificationRequestDetail, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("get certification request: %w", err)
	}
	if req == nil {
		return nil, ErrCertificationRequestNotFound
	}
	return req, nil
}

// Approve creates the blockchain_jobs row the existing worker picks up —
// nothing else in the certification pipeline changes behavior once this
// runs.
func (s *CertificationReviewService) Approve(ctx context.Context, reviewerID, requestID int64) error {
	if _, err := s.requests.Approve(ctx, requestID, reviewerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCertificationRequestNotPending
		}
		return fmt.Errorf("approve certification request: %w", err)
	}
	return nil
}

// Delete removes a certification request row, but only once its blockchain
// attempt has definitively failed or reverted — lets the admin panel clear
// retry noise (a batch's Nth failed attempt) without risking the deletion of
// a request that's still pending review, in flight, or already confirmed.
func (s *CertificationReviewService) Delete(ctx context.Context, requestID int64) error {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("get certification request: %w", err)
	}
	if req == nil {
		return ErrCertificationRequestNotFound
	}
	if req.JobStatus == nil ||
		(*req.JobStatus != model.CertificationStatusFailed && *req.JobStatus != model.CertificationStatusReverted) {
		return ErrCertificationRequestNotFailed
	}
	if err := s.requests.Delete(ctx, requestID); err != nil {
		return fmt.Errorf("delete certification request: %w", err)
	}
	return nil
}

func (s *CertificationReviewService) Reject(ctx context.Context, reviewerID, requestID int64, reason string) error {
	if err := validateRejectionReason(reason); err != nil {
		return err
	}
	if err := s.requests.Reject(ctx, requestID, reviewerID, reason); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCertificationRequestNotPending
		}
		return fmt.Errorf("reject certification request: %w", err)
	}
	return nil
}
