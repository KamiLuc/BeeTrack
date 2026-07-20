package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrRejectionReasonRequired = errors.New("rejection reason is required")
	ErrListingNotPending       = errors.New("listing is not pending review")
)

// ListingModerationStore is the persistence interface for the admin listing review queue.
type ListingModerationStore interface {
	ListPendingReview(ctx context.Context, limit, offset int) ([]*model.Listing, int64, error)
	GetByID(ctx context.Context, id int64) (*model.Listing, error)
	Approve(ctx context.Context, id, reviewerID int64) error
	Reject(ctx context.Context, id, reviewerID int64, reason string) error
}

type ListingModerationService struct {
	listings ListingModerationStore
}

func NewListingModerationService(listings ListingModerationStore) *ListingModerationService {
	return &ListingModerationService{listings: listings}
}

func (s *ListingModerationService) ListPending(ctx context.Context, limit, offset int) ([]*model.Listing, int64, error) {
	return s.listings.ListPendingReview(ctx, limit, offset)
}

func (s *ListingModerationService) Get(ctx context.Context, listingID int64) (*model.Listing, error) {
	l, err := s.listings.GetByID(ctx, listingID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrListingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get listing: %w", err)
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
	if _, err := s.pendingListing(ctx, listingID); err != nil {
		return err
	}
	if err := s.listings.Approve(ctx, listingID, reviewerID); err != nil {
		return fmt.Errorf("approve listing: %w", err)
	}
	return nil
}

func (s *ListingModerationService) Reject(ctx context.Context, reviewerID, listingID int64, reason string) error {
	if reason == "" {
		return ErrRejectionReasonRequired
	}
	if _, err := s.pendingListing(ctx, listingID); err != nil {
		return err
	}
	if err := s.listings.Reject(ctx, listingID, reviewerID, reason); err != nil {
		return fmt.Errorf("reject listing: %w", err)
	}
	return nil
}
