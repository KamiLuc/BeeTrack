package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// ErrCannotFavoriteOwnListing is returned when a user tries to favorite their own listing.
var ErrCannotFavoriteOwnListing = errors.New("cannot favorite your own listing")

// FavoriteListingReader fetches listings for favorite visibility checks.
type FavoriteListingReader interface {
	GetByID(ctx context.Context, id int64) (*model.Listing, error)
}

// ListingFavoriteStore is the persistence interface for listing favorites.
type ListingFavoriteStore interface {
	Add(ctx context.Context, f *model.ListingFavorite) error
	Remove(ctx context.Context, userID, listingID int64) error
	Exists(ctx context.Context, userID, listingID int64) (bool, error)
	ListListingsByUserID(ctx context.Context, userID int64) ([]*model.Listing, error)
}

// ListingFavoriteService handles business logic for favoriting listings.
type ListingFavoriteService struct {
	favorites ListingFavoriteStore
	listings  FavoriteListingReader
}

// NewListingFavoriteService creates a ListingFavoriteService with the given dependencies.
func NewListingFavoriteService(favorites ListingFavoriteStore, listings FavoriteListingReader) *ListingFavoriteService {
	return &ListingFavoriteService{favorites: favorites, listings: listings}
}

// Add saves a listing to the user's favorites after verifying it is visible to them.
func (s *ListingFavoriteService) Add(ctx context.Context, userID, listingID int64) error {
	l, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return fmt.Errorf("get listing: %w", err)
	}
	if l.IsHidden && l.UserID != userID {
		return ErrListingNotFound
	}
	if l.UserID == userID {
		return ErrCannotFavoriteOwnListing
	}
	if err := s.favorites.Add(ctx, &model.ListingFavorite{UserID: userID, ListingID: listingID}); err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

// Remove deletes a listing from the user's favorites (no-op if it wasn't favorited).
func (s *ListingFavoriteService) Remove(ctx context.Context, userID, listingID int64) error {
	if err := s.favorites.Remove(ctx, userID, listingID); err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

// IsFavorite reports whether the user has favorited the given listing.
func (s *ListingFavoriteService) IsFavorite(ctx context.Context, userID, listingID int64) (bool, error) {
	exists, err := s.favorites.Exists(ctx, userID, listingID)
	if err != nil {
		return false, fmt.Errorf("check favorite: %w", err)
	}
	return exists, nil
}

// List returns the listings the user has favorited, most recently favorited first.
func (s *ListingFavoriteService) List(ctx context.Context, userID int64) ([]*model.Listing, error) {
	listings, err := s.favorites.ListListingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	return listings, nil
}
