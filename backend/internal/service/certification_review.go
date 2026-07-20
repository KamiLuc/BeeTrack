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
)

// CertificationRequestStore is the persistence interface for the admin
// certification-request review queue.
type CertificationRequestStore interface {
	ListPending(ctx context.Context, limit, offset int) ([]*model.HoneyBatchCertificationRequestDetail, int64, error)
	GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertificationRequestDetail, error)
	Approve(ctx context.Context, id, reviewerID int64) (*model.BlockchainJob, error)
	Reject(ctx context.Context, id, reviewerID int64, reason string) error
}

type CertificationReviewService struct {
	requests CertificationRequestStore
}

func NewCertificationReviewService(requests CertificationRequestStore) *CertificationReviewService {
	return &CertificationReviewService{requests: requests}
}

func (s *CertificationReviewService) ListPending(ctx context.Context, limit, offset int) ([]*model.HoneyBatchCertificationRequestDetail, int64, error) {
	return s.requests.ListPending(ctx, limit, offset)
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
