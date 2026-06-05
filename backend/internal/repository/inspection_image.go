package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type InspectionImageRepository struct {
	db *gorm.DB
}

// NewInspectionImageRepository creates a new InspectionImageRepository backed by db.
func NewInspectionImageRepository(db *gorm.DB) *InspectionImageRepository {
	return &InspectionImageRepository{db: db}
}

// Create inserts a new inspection image record.
func (r *InspectionImageRepository) Create(ctx context.Context, img *model.InspectionImage) error {
	return r.db.WithContext(ctx).Create(img).Error
}

// GetByID returns the image with the given id that belongs to inspectionID.
func (r *InspectionImageRepository) GetByID(ctx context.Context, imageID, inspectionID int64) (*model.InspectionImage, error) {
	var img model.InspectionImage
	err := r.db.WithContext(ctx).
		Where("id = ? AND inspection_id = ?", imageID, inspectionID).
		First(&img).Error
	return &img, err
}

// ListByInspectionID returns all images for the given inspectionID ordered by id.
func (r *InspectionImageRepository) ListByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error) {
	var images []*model.InspectionImage
	err := r.db.WithContext(ctx).
		Where("inspection_id = ?", inspectionID).
		Order("id ASC").
		Find(&images).Error
	return images, err
}

// Delete removes the image with the given id.
func (r *InspectionImageRepository) Delete(ctx context.Context, imageID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", imageID).
		Delete(&model.InspectionImage{}).Error
}

// ListByInspectionIDForCleanup returns all images for the given inspectionID (used for file cleanup before cascade delete).
func (r *InspectionImageRepository) ListByInspectionIDForCleanup(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error) {
	return r.ListByInspectionID(ctx, inspectionID)
}
