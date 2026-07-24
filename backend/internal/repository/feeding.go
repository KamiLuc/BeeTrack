package repository

import (
	"context"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// FeedingRepository persists feeding records.
type FeedingRepository struct {
	db *gorm.DB
}

// NewFeedingRepository creates a new FeedingRepository backed by db.
func NewFeedingRepository(db *gorm.DB) *FeedingRepository {
	return &FeedingRepository{db: db}
}

// Create inserts a new feeding record.
func (r *FeedingRepository) Create(ctx context.Context, f *model.Feeding) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// BulkCreate inserts multiple feeding records in a single transaction.
func (r *FeedingRepository) BulkCreate(ctx context.Context, feedings []*model.Feeding) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, f := range feedings {
			if err := tx.Create(f).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID returns the feeding with the given id that belongs to hiveID.
func (r *FeedingRepository) GetByID(ctx context.Context, feedingID, hiveID int64) (*model.Feeding, error) {
	type row struct {
		model.Feeding
		FeederName string `gorm:"column:fed_by_name"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("feedings f").
		Select("f.*, u.name AS fed_by_name").
		Joins("LEFT JOIN users u ON u.id = f.fed_by").
		Where("f.id = ? AND f.hive_id = ?", feedingID, hiveID).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	fd := result.Feeding
	fd.FedByName = result.FeederName
	return &fd, nil
}

// CountByHiveID returns the total number of feedings for hiveID.
func (r *FeedingRepository) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Feeding{}).
		Where("hive_id = ?", hiveID).
		Count(&count).Error
	return count, err
}

// ListByHiveID returns feedings for hiveID ordered by fed_at descending with pagination.
func (r *FeedingRepository) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Feeding, error) {
	type row struct {
		model.Feeding
		FeederName string `gorm:"column:fed_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("feedings f").
		Select("f.*, u.name AS fed_by_name").
		Joins("LEFT JOIN users u ON u.id = f.fed_by").
		Where("f.hive_id = ?", hiveID).
		Order("f.fed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	feedings := make([]*model.Feeding, len(rows))
	for i, r := range rows {
		fd := r.Feeding
		fd.FedByName = r.FeederName
		feedings[i] = &fd
	}
	return feedings, nil
}

// ListByHiveIDsAndRange returns feedings for the given hive IDs with fed_at in
// [from, to), ordered by hive ID then fed_at descending, for report generation.
func (r *FeedingRepository) ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Feeding, error) {
	if len(hiveIDs) == 0 {
		return []*model.Feeding{}, nil
	}
	type row struct {
		model.Feeding
		FeederName string `gorm:"column:fed_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("feedings f").
		Select("f.*, u.name AS fed_by_name").
		Joins("LEFT JOIN users u ON u.id = f.fed_by").
		Where("f.hive_id IN ? AND f.fed_at >= ? AND f.fed_at < ?", hiveIDs, from, to).
		Order("f.hive_id ASC, f.fed_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	feedings := make([]*model.Feeding, len(rows))
	for i, r := range rows {
		fd := r.Feeding
		fd.FedByName = r.FeederName
		feedings[i] = &fd
	}
	return feedings, nil
}

// Update persists all mutable fields of f.
func (r *FeedingRepository) Update(ctx context.Context, f *model.Feeding) error {
	return r.db.WithContext(ctx).
		Model(f).
		Updates(map[string]any{
			"fed_at":     f.FedAt,
			"feed_type":  f.FeedType,
			"amount":     f.Amount,
			"notes":      f.Notes,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// Delete removes the feeding with the given id.
func (r *FeedingRepository) Delete(ctx context.Context, feedingID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", feedingID).
		Delete(&model.Feeding{}).Error
}

// DistinctFeedTypes returns the distinct feed types userID has previously
// entered, most recently used first.
func (r *FeedingRepository) DistinctFeedTypes(ctx context.Context, userID int64) ([]string, error) {
	types := []string{}
	err := r.db.WithContext(ctx).
		Model(&model.Feeding{}).
		Select("feed_type").
		Where("fed_by = ? AND feed_type <> ''", userID).
		Group("feed_type").
		Order("MAX(fed_at) DESC").
		Pluck("feed_type", &types).Error
	return types, err
}

// DistinctAmounts returns the distinct amounts userID has previously entered,
// most recently used first.
func (r *FeedingRepository) DistinctAmounts(ctx context.Context, userID int64) ([]string, error) {
	amounts := []string{}
	err := r.db.WithContext(ctx).
		Model(&model.Feeding{}).
		Select("amount").
		Where("fed_by = ? AND amount <> ''", userID).
		Group("amount").
		Order("MAX(fed_at) DESC").
		Pluck("amount", &amounts).Error
	return amounts, err
}
