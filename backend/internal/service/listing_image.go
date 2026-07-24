package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beetrack/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrListingImageNotFound  = errors.New("listing image not found")
	ErrListingImageMaxImages = errors.New("a listing may have at most 3 images")
	ErrListingImageLastPhoto = errors.New("a listing must keep at least one photo")
)

// listingAllowedMIME maps accepted upload content types to file extensions.
var listingAllowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// listingExtMIME maps stored file extensions back to content types for serving.
var listingExtMIME = map[string]string{
	".jpg":  "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

// ListingImageReader fetches listings for ownership checks and persists the
// re-review reset triggered by a photo change.
type ListingImageReader interface {
	GetByID(ctx context.Context, id int64) (*model.Listing, error)
	Update(ctx context.Context, l *model.Listing) error
}

// ListingImageStore is the persistence interface for listing images.
type ListingImageStore interface {
	CreateImage(ctx context.Context, img *model.ListingImage) error
	GetImageByID(ctx context.Context, imageID, listingID int64) (*model.ListingImage, error)
	ListImagesByListingID(ctx context.Context, listingID int64) ([]model.ListingImage, error)
	DeleteImage(ctx context.Context, imageID int64) error
}

// ListingImageService handles storing and retrieving listing photos on disk.
type ListingImageService struct {
	listings    ListingImageReader
	images      ListingImageStore
	storagePath string
}

// NewListingImageService creates a ListingImageService that stores files under storagePath.
func NewListingImageService(listings ListingImageReader, images ListingImageStore, storagePath string) *ListingImageService {
	return &ListingImageService{listings: listings, images: images, storagePath: storagePath}
}

// FilePath returns the absolute path for a stored image filename.
func (s *ListingImageService) FilePath(filename string) string {
	return filepath.Join(s.storagePath, filename)
}

// ContentType returns the MIME type for a stored image based on its extension.
func (s *ListingImageService) ContentType(filename string) string {
	if mime, ok := listingExtMIME[filepath.Ext(filename)]; ok {
		return mime
	}
	return "application/octet-stream"
}

// ownedListing fetches a listing and verifies it belongs to userID.
func (s *ListingImageService) ownedListing(ctx context.Context, userID, listingID int64) (*model.Listing, error) {
	l, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if l.UserID != userID {
		return nil, ErrNotListingOwner
	}
	return l, nil
}

// Upload validates ownership and the file, writes it to disk, and creates the DB record.
func (s *ListingImageService) Upload(ctx context.Context, userID, listingID int64, mimeType string, data []byte) (*model.ListingImage, error) {
	if len(data) > MaxImageBytes {
		return nil, ErrImageTooLarge
	}
	ext, ok := listingAllowedMIME[mimeType]
	if !ok {
		return nil, ErrInvalidImageType
	}
	l, err := s.ownedListing(ctx, userID, listingID)
	if err != nil {
		return nil, err
	}
	existing, err := s.images.ListImagesByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	if len(existing) >= maxListingImages {
		return nil, ErrListingImageMaxImages
	}
	if err := os.MkdirAll(s.storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	filename := uuid.New().String() + ext
	if err := os.WriteFile(filepath.Join(s.storagePath, filename), data, 0o644); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}
	img := &model.ListingImage{
		ListingID:    listingID,
		ImageURL:     filename,
		DisplayOrder: len(existing),
	}
	if err := s.images.CreateImage(ctx, img); err != nil {
		_ = os.Remove(filepath.Join(s.storagePath, filename))
		return nil, fmt.Errorf("create image record: %w", err)
	}
	if err := s.resetToPendingIfReviewed(ctx, l); err != nil {
		return nil, fmt.Errorf("reset listing to pending: %w", err)
	}
	return img, nil
}

// GetFile returns the stored image record for serving, without an ownership check (public).
func (s *ListingImageService) GetFile(ctx context.Context, listingID, imageID int64) (*model.ListingImage, error) {
	img, err := s.images.GetImageByID(ctx, imageID, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingImageNotFound
		}
		return nil, fmt.Errorf("get image: %w", err)
	}
	return img, nil
}

// Delete verifies ownership, removes the file from disk, and deletes the DB record.
// A listing's last remaining photo cannot be deleted — every listing must keep at least one.
func (s *ListingImageService) Delete(ctx context.Context, userID, listingID, imageID int64) error {
	l, err := s.ownedListing(ctx, userID, listingID)
	if err != nil {
		return err
	}
	img, err := s.images.GetImageByID(ctx, imageID, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingImageNotFound
		}
		return fmt.Errorf("get image: %w", err)
	}
	existing, err := s.images.ListImagesByListingID(ctx, listingID)
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}
	if len(existing) <= 1 {
		return ErrListingImageLastPhoto
	}
	_ = os.Remove(filepath.Join(s.storagePath, img.ImageURL))
	if err := s.images.DeleteImage(ctx, imageID); err != nil {
		return fmt.Errorf("delete image record: %w", err)
	}
	if err := s.resetToPendingIfReviewed(ctx, l); err != nil {
		return fmt.Errorf("reset listing to pending: %w", err)
	}
	return nil
}

// resetToPendingIfReviewed sends a listing whose photos just changed back
// into the moderation queue if it had already been approved or rejected —
// the reviewer hasn't seen this photo yet. A listing still pending review
// is left alone.
func (s *ListingImageService) resetToPendingIfReviewed(ctx context.Context, l *model.Listing) error {
	if l.Status != model.ListingStatusApproved && l.Status != model.ListingStatusRejected {
		return nil
	}
	l.Status = model.ListingStatusPending
	l.RejectionReason = nil
	l.ReviewedBy = nil
	l.ReviewedAt = nil
	return s.listings.Update(ctx, l)
}
