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
	ErrImageNotFound      = errors.New("image not found")
	ErrInvalidImageType   = errors.New("unsupported image type; allowed: image/jpeg, image/png, image/webp")
	ErrImageTooLarge      = errors.New("image exceeds 10 MB limit")
	ErrMaxImagesReached   = errors.New("inspection already has the maximum of 6 images")
)

const maxImageBytes = 10 * 1024 * 1024
const maxImagesPerInspection = 6

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// InspectionImageRepository is the persistence interface for inspection images.
type InspectionImageRepository interface {
	Create(ctx context.Context, img *model.InspectionImage) error
	GetByID(ctx context.Context, imageID, inspectionID int64) (*model.InspectionImage, error)
	ListByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error)
	ListByInspectionIDForCleanup(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error)
	Delete(ctx context.Context, imageID int64) error
}

// InspectionImageService handles storing and retrieving inspection photos.
type InspectionImageService struct {
	apiaries    ApiaryMembershipReader
	hives       InspectionHiveReader
	inspections InspectionRepository
	images      InspectionImageRepository
	storagePath string
}

// NewInspectionImageService creates an InspectionImageService that stores files under storagePath.
func NewInspectionImageService(
	apiaries ApiaryMembershipReader,
	hives InspectionHiveReader,
	inspections InspectionRepository,
	images InspectionImageRepository,
	storagePath string,
) *InspectionImageService {
	return &InspectionImageService{
		apiaries:    apiaries,
		hives:       hives,
		inspections: inspections,
		images:      images,
		storagePath: storagePath,
	}
}

// FilePath returns the absolute path for a stored image filename.
func (s *InspectionImageService) FilePath(filename string) string {
	return filepath.Join(s.storagePath, filename)
}

// Upload validates access, writes the file to disk, and creates the DB record.
func (s *InspectionImageService) Upload(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64, mimeType string, data []byte) (*model.InspectionImage, error) {
	if len(data) > maxImageBytes {
		return nil, ErrImageTooLarge
	}
	ext, ok := allowedMIME[mimeType]
	if !ok {
		return nil, ErrInvalidImageType
	}
	if err := s.checkInspectionAccess(ctx, apiaryID, userID, hiveID, inspectionID); err != nil {
		return nil, err
	}
	existing, err := s.images.ListByInspectionID(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("count images: %w", err)
	}
	if len(existing) >= maxImagesPerInspection {
		return nil, ErrMaxImagesReached
	}
	if err := os.MkdirAll(s.storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	filename := uuid.New().String() + ext
	if err := os.WriteFile(filepath.Join(s.storagePath, filename), data, 0o644); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}
	img := &model.InspectionImage{
		InspectionID: inspectionID,
		Filename:     filename,
		MimeType:     mimeType,
	}
	if err := s.images.Create(ctx, img); err != nil {
		_ = os.Remove(filepath.Join(s.storagePath, filename))
		return nil, fmt.Errorf("create image record: %w", err)
	}
	return img, nil
}

// List returns all images for an inspection after verifying access.
func (s *InspectionImageService) List(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64) ([]*model.InspectionImage, error) {
	if err := s.checkInspectionAccess(ctx, apiaryID, userID, hiveID, inspectionID); err != nil {
		return nil, err
	}
	imgs, err := s.images.ListByInspectionID(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	return imgs, nil
}

// Delete verifies access, removes the file from disk, and deletes the DB record.
func (s *InspectionImageService) Delete(ctx context.Context, userID, apiaryID, hiveID, inspectionID, imageID int64) error {
	if err := s.checkInspectionAccess(ctx, apiaryID, userID, hiveID, inspectionID); err != nil {
		return err
	}
	img, err := s.images.GetByID(ctx, imageID, inspectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrImageNotFound
		}
		return fmt.Errorf("get image: %w", err)
	}
	_ = os.Remove(filepath.Join(s.storagePath, img.Filename))
	if err := s.images.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("delete image record: %w", err)
	}
	return nil
}

// DeleteFilesForInspection removes all image files on disk for an inspection (DB records are cleaned by CASCADE).
func (s *InspectionImageService) DeleteFilesForInspection(ctx context.Context, inspectionID int64) {
	imgs, err := s.images.ListByInspectionIDForCleanup(ctx, inspectionID)
	if err != nil {
		return
	}
	for _, img := range imgs {
		_ = os.Remove(filepath.Join(s.storagePath, img.Filename))
	}
}

func (s *InspectionImageService) checkInspectionAccess(ctx context.Context, apiaryID, userID, hiveID, inspectionID int64) error {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApiaryNotFound
		}
		return fmt.Errorf("get apiary: %w", err)
	}
	if _, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHiveNotFound
		}
		return fmt.Errorf("get hive: %w", err)
	}
	if _, err := s.inspections.GetByID(ctx, inspectionID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInspectionNotFound
		}
		return fmt.Errorf("get inspection: %w", err)
	}
	return nil
}
