package repository

import (
	"context"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// HarvestRepository persists harvest records.
type HarvestRepository struct {
	db *gorm.DB
}

// NewHarvestRepository creates a new HarvestRepository backed by db.
func NewHarvestRepository(db *gorm.DB) *HarvestRepository {
	return &HarvestRepository{db: db}
}

// Create inserts a new harvest record.
func (r *HarvestRepository) Create(ctx context.Context, h *model.Harvest) error {
	return r.db.WithContext(ctx).Create(h).Error
}

// GetByID returns the harvest with the given id that belongs to hiveID.
func (r *HarvestRepository) GetByID(ctx context.Context, harvestID, hiveID int64) (*model.Harvest, error) {
	type row struct {
		model.Harvest
		HarvesterName string `gorm:"column:harvested_by_name"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("harvests h").
		Select("h.*, u.name AS harvested_by_name").
		Joins("LEFT JOIN users u ON u.id = h.harvested_by").
		Where("h.id = ? AND h.hive_id = ?", harvestID, hiveID).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	h := result.Harvest
	h.HarvestedByName = result.HarvesterName
	return &h, nil
}

// CountByHiveID returns the total number of harvests for hiveID.
func (r *HarvestRepository) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Harvest{}).
		Where("hive_id = ?", hiveID).
		Count(&count).Error
	return count, err
}

// ListByHiveID returns harvests for hiveID ordered by harvested_at descending with pagination.
func (r *HarvestRepository) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Harvest, error) {
	type row struct {
		model.Harvest
		HarvesterName string `gorm:"column:harvested_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("harvests h").
		Select("h.*, u.name AS harvested_by_name").
		Joins("LEFT JOIN users u ON u.id = h.harvested_by").
		Where("h.hive_id = ?", hiveID).
		Order("h.harvested_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	harvests := make([]*model.Harvest, len(rows))
	for i, r := range rows {
		h := r.Harvest
		h.HarvestedByName = r.HarvesterName
		harvests[i] = &h
	}
	return harvests, nil
}

// ListByHiveIDsAndRange returns harvests for the given hive IDs with harvested_at
// in [from, to), ordered by hive ID then harvested_at descending, for report generation.
func (r *HarvestRepository) ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Harvest, error) {
	if len(hiveIDs) == 0 {
		return []*model.Harvest{}, nil
	}
	type row struct {
		model.Harvest
		HarvesterName string `gorm:"column:harvested_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("harvests h").
		Select("h.*, u.name AS harvested_by_name").
		Joins("LEFT JOIN users u ON u.id = h.harvested_by").
		Where("h.hive_id IN ? AND h.harvested_at >= ? AND h.harvested_at < ?", hiveIDs, from, to).
		Order("h.hive_id ASC, h.harvested_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	harvests := make([]*model.Harvest, len(rows))
	for i, r := range rows {
		hv := r.Harvest
		hv.HarvestedByName = r.HarvesterName
		harvests[i] = &hv
	}
	return harvests, nil
}

// Update persists all mutable fields of h.
func (r *HarvestRepository) Update(ctx context.Context, h *model.Harvest) error {
	return r.db.WithContext(ctx).
		Model(h).
		Updates(map[string]any{
			"harvested_at": h.HarvestedAt,
			"frames":       h.Frames,
			"half_frames":  h.HalfFrames,
			"kilograms":    h.Kilograms,
			"notes":        h.Notes,
			"updated_at":   gorm.Expr("NOW()"),
		}).Error
}

// Delete removes the harvest with the given id.
func (r *HarvestRepository) Delete(ctx context.Context, harvestID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", harvestID).
		Delete(&model.Harvest{}).Error
}
