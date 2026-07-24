package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockListingImageStore struct {
	images    []model.ListingImage
	image     *model.ListingImage
	created   *model.ListingImage
	deletedID int64
}

func (m *mockListingImageStore) CreateImage(ctx context.Context, img *model.ListingImage) error {
	img.ID = 1
	m.created = img
	return nil
}

func (m *mockListingImageStore) GetImageByID(ctx context.Context, imageID, listingID int64) (*model.ListingImage, error) {
	if m.image == nil || m.image.ID != imageID || m.image.ListingID != listingID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.image, nil
}

func (m *mockListingImageStore) ListImagesByListingID(ctx context.Context, listingID int64) ([]model.ListingImage, error) {
	return m.images, nil
}

func (m *mockListingImageStore) DeleteImage(ctx context.Context, imageID int64) error {
	m.deletedID = imageID
	return nil
}

type mockListingReader struct {
	listing *model.Listing
	updated *model.Listing
}

func (m *mockListingReader) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	if m.listing == nil || m.listing.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return m.listing, nil
}

func (m *mockListingReader) Update(ctx context.Context, l *model.Listing) error {
	m.updated = l
	return nil
}

func newTestListingImageService(t *testing.T) (*ListingImageService, *mockListingReader, *mockListingImageStore, string) {
	t.Helper()
	dir := t.TempDir()
	reader := &mockListingReader{}
	store := &mockListingImageStore{}
	svc := NewListingImageService(reader, store, dir)
	return svc, reader, store, dir
}

func TestListingImageUpload_Success(t *testing.T) {
	svc, reader, store, dir := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}

	img, err := svc.Upload(context.Background(), 3, 5, "image/jpeg", []byte{0xFF, 0xD8, 0xFF})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store.created == nil {
		t.Fatal("expected Create to be called")
	}
	if filepath.Ext(img.ImageURL) != ".jpg" {
		t.Errorf("expected .jpg filename, got %q", img.ImageURL)
	}
	if _, err := os.Stat(filepath.Join(dir, img.ImageURL)); err != nil {
		t.Errorf("expected file written to disk: %v", err)
	}
}

func TestListingImageUpload_ResetsRejectedListingToPending(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reason := "blurry photo"
	reader.listing = &model.Listing{
		ID:              5,
		UserID:          3,
		Status:          model.ListingStatusRejected,
		RejectionReason: &reason,
	}

	_, err := svc.Upload(context.Background(), 3, 5, "image/jpeg", []byte{0xFF, 0xD8, 0xFF})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.updated == nil {
		t.Fatal("expected the listing to be updated")
	}
	if reader.updated.Status != model.ListingStatusPending {
		t.Errorf("expected status pending, got %q", reader.updated.Status)
	}
	if reader.updated.RejectionReason != nil {
		t.Errorf("expected rejection reason cleared, got %v", *reader.updated.RejectionReason)
	}
}

func TestListingImageUpload_LeavesPendingListingAlone(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3, Status: model.ListingStatusPending}

	_, err := svc.Upload(context.Background(), 3, 5, "image/jpeg", []byte{0xFF, 0xD8, 0xFF})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.updated != nil {
		t.Error("expected no update for an already-pending listing")
	}
}

func TestListingImageUpload_DisplayOrderIncrements(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}

	img, err := svc.Upload(context.Background(), 3, 5, "image/png", []byte{0x89, 0x50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// store starts with zero existing images, so display_order should be 0.
	if img.DisplayOrder != 0 {
		t.Errorf("expected display_order 0, got %d", img.DisplayOrder)
	}
}

func TestListingImageUpload_InvalidType(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}

	_, err := svc.Upload(context.Background(), 3, 5, "image/gif", []byte{1})
	if !errors.Is(err, ErrInvalidImageType) {
		t.Errorf("expected ErrInvalidImageType, got %v", err)
	}
}

func TestListingImageUpload_TooLarge(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}

	_, err := svc.Upload(context.Background(), 3, 5, "image/jpeg", make([]byte, MaxImageBytes+1))
	if !errors.Is(err, ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge, got %v", err)
	}
}

func TestListingImageUpload_ListingNotFound(t *testing.T) {
	svc, _, _, _ := newTestListingImageService(t)

	_, err := svc.Upload(context.Background(), 3, 5, "image/jpeg", []byte{1})
	if !errors.Is(err, ErrListingNotFound) {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingImageUpload_NotOwner(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}

	_, err := svc.Upload(context.Background(), 99, 5, "image/jpeg", []byte{1})
	if !errors.Is(err, ErrNotListingOwner) {
		t.Errorf("expected ErrNotListingOwner, got %v", err)
	}
}

func TestListingImageUpload_MaxImages(t *testing.T) {
	svc, reader, store, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}
	store.images = make([]model.ListingImage, maxListingImages)

	_, err := svc.Upload(context.Background(), 3, 5, "image/jpeg", []byte{0xFF})
	if !errors.Is(err, ErrListingImageMaxImages) {
		t.Errorf("expected ErrListingImageMaxImages, got %v", err)
	}
}

func TestListingImageGetFile_Success(t *testing.T) {
	svc, _, store, _ := newTestListingImageService(t)
	store.image = &model.ListingImage{ID: 7, ListingID: 5, ImageURL: "abc.png"}

	img, err := svc.GetFile(context.Background(), 5, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img.ImageURL != "abc.png" {
		t.Errorf("unexpected image: %+v", img)
	}
}

func TestListingImageGetFile_NotFound(t *testing.T) {
	svc, _, _, _ := newTestListingImageService(t)

	_, err := svc.GetFile(context.Background(), 5, 7)
	if !errors.Is(err, ErrListingImageNotFound) {
		t.Errorf("expected ErrListingImageNotFound, got %v", err)
	}
}

func TestListingImageUpload_DisplayOrderWithExistingImages(t *testing.T) {
	svc, reader, store, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}
	store.images = []model.ListingImage{{ID: 1, ListingID: 5}}

	img, err := svc.Upload(context.Background(), 3, 5, "image/png", []byte{0x89, 0x50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img.DisplayOrder != 1 {
		t.Errorf("expected display_order 1, got %d", img.DisplayOrder)
	}
}

func TestListingImageFilePath(t *testing.T) {
	svc, _, _, dir := newTestListingImageService(t)

	got := svc.FilePath("abc.jpg")
	want := filepath.Join(dir, "abc.jpg")
	if got != want {
		t.Errorf("FilePath() = %q, want %q", got, want)
	}
}

func TestListingImageContentType(t *testing.T) {
	svc, _, _, _ := newTestListingImageService(t)
	cases := map[string]string{
		"a.jpg":     "image/jpeg",
		"b.png":     "image/png",
		"c.webp":    "image/webp",
		"d.unknown": "application/octet-stream",
	}
	for filename, want := range cases {
		if got := svc.ContentType(filename); got != want {
			t.Errorf("ContentType(%q) = %q, want %q", filename, got, want)
		}
	}
}

func TestListingImageDelete_Success(t *testing.T) {
	svc, reader, store, dir := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}
	filename := "todelete.jpg"
	_ = os.WriteFile(filepath.Join(dir, filename), []byte{1}, 0o644)
	store.image = &model.ListingImage{ID: 7, ListingID: 5, ImageURL: filename}
	store.images = []model.ListingImage{
		{ID: 7, ListingID: 5, ImageURL: filename},
		{ID: 8, ListingID: 5, ImageURL: "other.jpg"},
	}

	if err := svc.Delete(context.Background(), 3, 5, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.deletedID != 7 {
		t.Errorf("expected deletedID 7, got %d", store.deletedID)
	}
	if _, err := os.Stat(filepath.Join(dir, filename)); !os.IsNotExist(err) {
		t.Error("expected file removed from disk")
	}
}

func TestListingImageDelete_ResetsApprovedListingToPending(t *testing.T) {
	svc, reader, store, dir := newTestListingImageService(t)
	reviewer := int64(9)
	reviewedAt := time.Now()
	reader.listing = &model.Listing{
		ID:         5,
		UserID:     3,
		Status:     model.ListingStatusApproved,
		ReviewedBy: &reviewer,
		ReviewedAt: &reviewedAt,
	}
	filename := "todelete.jpg"
	_ = os.WriteFile(filepath.Join(dir, filename), []byte{1}, 0o644)
	store.image = &model.ListingImage{ID: 7, ListingID: 5, ImageURL: filename}
	store.images = []model.ListingImage{
		{ID: 7, ListingID: 5, ImageURL: filename},
		{ID: 8, ListingID: 5, ImageURL: "other.jpg"},
	}

	if err := svc.Delete(context.Background(), 3, 5, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.updated == nil {
		t.Fatal("expected the listing to be updated")
	}
	if reader.updated.Status != model.ListingStatusPending {
		t.Errorf("expected status pending, got %q", reader.updated.Status)
	}
	if reader.updated.ReviewedBy != nil || reader.updated.ReviewedAt != nil {
		t.Error("expected reviewer info cleared")
	}
}

func TestListingImageDelete_LastPhotoRejected(t *testing.T) {
	svc, reader, store, dir := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}
	filename := "onlyphoto.jpg"
	_ = os.WriteFile(filepath.Join(dir, filename), []byte{1}, 0o644)
	store.image = &model.ListingImage{ID: 7, ListingID: 5, ImageURL: filename}
	store.images = []model.ListingImage{{ID: 7, ListingID: 5, ImageURL: filename}}

	err := svc.Delete(context.Background(), 3, 5, 7)
	if !errors.Is(err, ErrListingImageLastPhoto) {
		t.Errorf("expected ErrListingImageLastPhoto, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, filename)); err != nil {
		t.Error("expected file to remain on disk")
	}
}

func TestListingImageDelete_NotOwner(t *testing.T) {
	svc, reader, store, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}
	store.image = &model.ListingImage{ID: 7, ListingID: 5, ImageURL: "x.jpg"}

	err := svc.Delete(context.Background(), 99, 5, 7)
	if !errors.Is(err, ErrNotListingOwner) {
		t.Errorf("expected ErrNotListingOwner, got %v", err)
	}
}

func TestListingImageDelete_NotFound(t *testing.T) {
	svc, reader, _, _ := newTestListingImageService(t)
	reader.listing = &model.Listing{ID: 5, UserID: 3}

	err := svc.Delete(context.Background(), 3, 5, 99)
	if !errors.Is(err, ErrListingImageNotFound) {
		t.Errorf("expected ErrListingImageNotFound, got %v", err)
	}
}
