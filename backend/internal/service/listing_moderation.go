package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"gorm.io/gorm"
)

// MinRejectionReasonLength rules out an empty click-through rejection.
const MinRejectionReasonLength = 3

var (
	ErrRejectionReasonRequired = errors.New("rejection reason is required")
	ErrRejectionReasonTooShort = fmt.Errorf("rejection reason must be at least %d characters", MinRejectionReasonLength)
	ErrRejectionReasonTooLong  = fmt.Errorf("rejection reason must be at most %d characters", validation.Large.MaxLength())
	ErrListingNotPending       = errors.New("listing is not pending review")
	ErrListingNotApproved      = errors.New("listing is not approved")
	ErrListingNotRemoved       = errors.New("listing is not removed")
)

// validateRejectionReason is shared by listing and certification-request rejection.
func validateRejectionReason(reason string) error {
	trimmed := strings.TrimSpace(reason)
	switch {
	case trimmed == "":
		return ErrRejectionReasonRequired
	case len([]rune(trimmed)) < MinRejectionReasonLength:
		return ErrRejectionReasonTooShort
	case validation.TooLong(trimmed, validation.Large):
		return ErrRejectionReasonTooLong
	default:
		return nil
	}
}

// ListingModerationStore is the persistence interface for the admin listing review queue.
type ListingModerationStore interface {
	ListReview(ctx context.Context, status, keyword, sortDir string, limit, offset int) ([]*model.Listing, int64, error)
	GetByIDForReview(ctx context.Context, id int64) (*model.Listing, error)
	Approve(ctx context.Context, id, reviewerID int64) error
	Reject(ctx context.Context, id, reviewerID int64, reason string) error
	Remove(ctx context.Context, id, reviewerID int64, reason string) error
	Restore(ctx context.Context, id, reviewerID int64) error
}

// CertificationRequestReader looks up a honey batch's latest certification
// request, so the admin listing detail view can link a HONEY listing to its
// certification review.
type CertificationRequestReader interface {
	GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error)
}

type ListingModerationService struct {
	listings     ListingModerationStore
	certRequests CertificationRequestReader
}

func NewListingModerationService(listings ListingModerationStore, certRequests CertificationRequestReader) *ListingModerationService {
	return &ListingModerationService{listings: listings, certRequests: certRequests}
}

// List returns listings for the admin review queue, optionally filtered by status
// (empty means all statuses) and keyword (title or owner email), sorted by
// submission date via sortDir ("asc"/"desc").
func (s *ListingModerationService) List(ctx context.Context, status, keyword, sortDir string, limit, offset int) ([]*model.Listing, int64, error) {
	return s.listings.ListReview(ctx, status, keyword, sortDir, limit, offset)
}

func (s *ListingModerationService) Get(ctx context.Context, listingID int64) (*model.Listing, error) {
	l, err := s.listings.GetByIDForReview(ctx, listingID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrListingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if l.HoneyBatchID != nil {
		req, err := s.certRequests.GetLatestByBatchID(ctx, *l.HoneyBatchID)
		if err != nil {
			return nil, fmt.Errorf("get certification request: %w", err)
		}
		if req != nil {
			l.CertificationRequestID = &req.ID
			l.CertificationRequestStatus = req.Status
		}
	}
	return l, nil
}

func (s *ListingModerationService) pendingListing(ctx context.Context, listingID int64) (*model.Listing, error) {
	l, err := s.Get(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if l.Status != model.ListingStatusPending {
		return nil, ErrListingNotPending
	}
	return l, nil
}

func (s *ListingModerationService) Approve(ctx context.Context, reviewerID, listingID int64) error {
	l, err := s.pendingListing(ctx, listingID)
	if err != nil {
		return err
	}
	if len(l.Images) == 0 {
		return ErrListingPhotoRequired
	}
	if err := s.listings.Approve(ctx, listingID, reviewerID); err != nil {
		return fmt.Errorf("approve listing: %w", err)
	}
	return nil
}

func (s *ListingModerationService) Reject(ctx context.Context, reviewerID, listingID int64, reason string) error {
	if err := validateRejectionReason(reason); err != nil {
		return err
	}
	if _, err := s.pendingListing(ctx, listingID); err != nil {
		return err
	}
	if err := s.listings.Reject(ctx, listingID, reviewerID, reason); err != nil {
		return fmt.Errorf("reject listing: %w", err)
	}
	return nil
}

// Remove takes a live (approved) listing off the marketplace, recording who did it and why.
func (s *ListingModerationService) Remove(ctx context.Context, reviewerID, listingID int64, reason string) error {
	if err := validateRejectionReason(reason); err != nil {
		return err
	}
	l, err := s.Get(ctx, listingID)
	if err != nil {
		return err
	}
	if l.Status != model.ListingStatusApproved {
		return ErrListingNotApproved
	}
	if err := s.listings.Remove(ctx, listingID, reviewerID, reason); err != nil {
		return fmt.Errorf("remove listing: %w", err)
	}
	return nil
}

// Restore brings a removed listing back onto the marketplace as approved.
func (s *ListingModerationService) Restore(ctx context.Context, reviewerID, listingID int64) error {
	l, err := s.Get(ctx, listingID)
	if err != nil {
		return err
	}
	if l.Status != model.ListingStatusRemoved {
		return ErrListingNotRemoved
	}
	if err := s.listings.Restore(ctx, listingID, reviewerID); err != nil {
		return fmt.Errorf("restore listing: %w", err)
	}
	return nil
}
