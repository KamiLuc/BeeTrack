package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// ListingFilter holds optional criteria for searching listings.
type ListingFilter struct {
	Category    string
	Keyword     string
	PriceMin    *float64
	PriceMax    *float64
	PostedAfter *string
	OwnerUserID *int64
	Limit       int
	Offset      int
}

// ListingRepository persists marketplace listings and their images.
type ListingRepository struct {
	db *gorm.DB
}

// NewListingRepository creates a new ListingRepository backed by db.
func NewListingRepository(db *gorm.DB) *ListingRepository {
	return &ListingRepository{db: db}
}

// Create inserts a listing and its images in a single transaction.
func (r *ListingRepository) Create(ctx context.Context, l *model.Listing) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(l).Error; err != nil {
			return err
		}
		for i := range l.Images {
			l.Images[i].ListingID = l.ID
			if err := tx.Create(&l.Images[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID returns the listing with the given id, including its images and apiary name.
func (r *ListingRepository) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	type row struct {
		model.Listing
		ApiaryName string `gorm:"column:apiary_name"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("listings l").
		Select("l.*, a.name AS apiary_name").
		Joins("LEFT JOIN apiaries a ON a.id = l.apiary_id").
		Where("l.id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	listing := result.Listing
	listing.ApiaryName = result.ApiaryName
	images, err := r.listImages(ctx, id)
	if err != nil {
		return nil, err
	}
	listing.Images = images
	return &listing, nil
}

// Count returns the number of listings matching the filter.
func (r *ListingRepository) Count(ctx context.Context, f ListingFilter) (int64, error) {
	var count int64
	err := r.applyFilter(r.db.WithContext(ctx).Model(&model.Listing{}), f).
		Count(&count).Error
	return count, err
}

// List returns listings matching the filter, ordered by created_at descending with pagination.
func (r *ListingRepository) List(ctx context.Context, f ListingFilter) ([]*model.Listing, error) {
	type row struct {
		model.Listing
		ApiaryName string `gorm:"column:apiary_name"`
	}
	var rows []row
	q := r.db.WithContext(ctx).
		Table("listings l").
		Select("l.*, a.name AS apiary_name").
		Joins("LEFT JOIN apiaries a ON a.id = l.apiary_id")
	q = r.applyFilterAliased(q, f)
	err := q.Order("l.created_at DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	listings := make([]*model.Listing, len(rows))
	for i, row := range rows {
		l := row.Listing
		l.ApiaryName = row.ApiaryName
		listings[i] = &l
	}
	return listings, nil
}

// Update persists all mutable fields of l.
func (r *ListingRepository) Update(ctx context.Context, l *model.Listing) error {
	return r.db.WithContext(ctx).
		Model(l).
		Updates(map[string]any{
			"title":         l.Title,
			"description":   l.Description,
			"category":      l.Category,
			"price":         l.Price,
			"quantity":      l.Quantity,
			"address":       l.Address,
			"apiary_id":     l.ApiaryID,
			"contact_phone": l.ContactPhone,
			"contact_email": l.ContactEmail,
			"updated_at":    gorm.Expr("NOW()"),
		}).Error
}

// SetHidden toggles the is_hidden flag of the listing with the given id.
func (r *ListingRepository) SetHidden(ctx context.Context, id int64, hidden bool) error {
	return r.db.WithContext(ctx).
		Model(&model.Listing{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_hidden":  hidden,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// AddImages inserts the given images.
func (r *ListingRepository) AddImages(ctx context.Context, images []model.ListingImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&images).Error
}

// CreateImage inserts a single listing image.
func (r *ListingRepository) CreateImage(ctx context.Context, img *model.ListingImage) error {
	return r.db.WithContext(ctx).Create(img).Error
}

// GetImageByID returns the image with the given id that belongs to listingID.
func (r *ListingRepository) GetImageByID(ctx context.Context, imageID, listingID int64) (*model.ListingImage, error) {
	var img model.ListingImage
	err := r.db.WithContext(ctx).
		Where("id = ? AND listing_id = ?", imageID, listingID).
		First(&img).Error
	return &img, err
}

// ListImagesByListingID returns the images for listingID ordered by display_order.
func (r *ListingRepository) ListImagesByListingID(ctx context.Context, listingID int64) ([]model.ListingImage, error) {
	return r.listImages(ctx, listingID)
}

// DeleteImage removes the image with the given id.
func (r *ListingRepository) DeleteImage(ctx context.Context, imageID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", imageID).
		Delete(&model.ListingImage{}).Error
}

// DeleteImages removes all images belonging to listingID.
func (r *ListingRepository) DeleteImages(ctx context.Context, listingID int64) error {
	return r.db.WithContext(ctx).
		Where("listing_id = ?", listingID).
		Delete(&model.ListingImage{}).Error
}

// Delete removes the listing with the given id; its images cascade in the database.
func (r *ListingRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.Listing{}).Error
}

// listImages returns the images for listingID ordered by display_order.
func (r *ListingRepository) listImages(ctx context.Context, listingID int64) ([]model.ListingImage, error) {
	var images []model.ListingImage
	err := r.db.WithContext(ctx).
		Where("listing_id = ?", listingID).
		Order("display_order ASC").
		Find(&images).Error
	return images, err
}

// applyFilter adds the filter's WHERE clauses to an unaliased listings query.
func (r *ListingRepository) applyFilter(q *gorm.DB, f ListingFilter) *gorm.DB {
	return r.buildFilter(q, f, "")
}

// applyFilterAliased adds the filter's WHERE clauses to a query using the "l" alias.
func (r *ListingRepository) applyFilterAliased(q *gorm.DB, f ListingFilter) *gorm.DB {
	return r.buildFilter(q, f, "l.")
}

// buildFilter applies shared listing filter conditions using the given column prefix.
func (r *ListingRepository) buildFilter(q *gorm.DB, f ListingFilter, p string) *gorm.DB {
	if f.OwnerUserID != nil {
		q = q.Where(p+"user_id = ?", *f.OwnerUserID)
	} else {
		q = q.Where(p + "is_hidden = FALSE")
	}
	if f.Category != "" {
		q = q.Where(p+"category = ?", f.Category)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("("+p+"title ILIKE ? OR "+p+"description ILIKE ?)", like, like)
	}
	if f.PriceMin != nil {
		q = q.Where(p+"price >= ?", *f.PriceMin)
	}
	if f.PriceMax != nil {
		q = q.Where(p+"price <= ?", *f.PriceMax)
	}
	if f.PostedAfter != nil {
		q = q.Where(p+"created_at >= ?", *f.PostedAfter)
	}
	return q
}
