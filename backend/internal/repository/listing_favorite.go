package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListingFavoriteRepository persists users' favorited listings.
type ListingFavoriteRepository struct {
	db *gorm.DB
}

// NewListingFavoriteRepository creates a new ListingFavoriteRepository backed by db.
func NewListingFavoriteRepository(db *gorm.DB) *ListingFavoriteRepository {
	return &ListingFavoriteRepository{db: db}
}

// Add inserts a favorite, doing nothing if the user already favorited the listing.
func (r *ListingFavoriteRepository) Add(ctx context.Context, f *model.ListingFavorite) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(f).Error
}

// Remove deletes the favorite linking userID and listingID, if any.
func (r *ListingFavoriteRepository) Remove(ctx context.Context, userID, listingID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND listing_id = ?", userID, listingID).
		Delete(&model.ListingFavorite{}).Error
}

// ListListingsByUserID returns the listings favorited by userID, most recently favorited first.
// Hidden listings are excluded unless the user owns them.
func (r *ListingFavoriteRepository) ListListingsByUserID(ctx context.Context, userID int64) ([]*model.Listing, error) {
	type row struct {
		model.Listing
		ApiaryName string `gorm:"column:apiary_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("listing_favorites f").
		Select("l.*, a.name AS apiary_name").
		Joins("JOIN listings l ON l.id = f.listing_id").
		Joins("LEFT JOIN apiaries a ON a.id = l.apiary_id").
		Where("f.user_id = ? AND (l.is_hidden = FALSE OR l.user_id = ?)", userID, userID).
		Order("f.created_at DESC").
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
