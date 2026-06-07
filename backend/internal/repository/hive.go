package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type HiveRepository struct {
	db *gorm.DB
}

func NewHiveRepository(db *gorm.DB) *HiveRepository {
	return &HiveRepository{db: db}
}

func (r *HiveRepository) Create(ctx context.Context, h *model.Hive) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *HiveRepository) IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Hive{}).
		Where("apiary_id = ? AND grid_row = ? AND grid_col = ?", apiaryID, row, col).
		Count(&count).Error
	return count > 0, err
}

func (r *HiveRepository) ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error) {
	var hives []*model.Hive
	err := r.db.WithContext(ctx).
		Where("apiary_id = ?", apiaryID).
		Order("grid_row ASC, grid_col ASC").
		Find(&hives).Error
	return hives, err
}

func (r *HiveRepository) GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error) {
	var h model.Hive
	err := r.db.WithContext(ctx).
		Where("id = ? AND apiary_id = ?", hiveID, apiaryID).
		First(&h).Error
	return &h, err
}

func (r *HiveRepository) Update(ctx context.Context, h *model.Hive) error {
	return r.db.WithContext(ctx).
		Model(h).
		Updates(map[string]any{
			"active":            h.Active,
			"frames":            h.Frames,
			"name":              h.Name,
			"queenless":         h.Queenless,
			"ready_for_harvest": h.ReadyForHarvest,
			"type":              h.Type,
			"updated_at":        gorm.Expr("NOW()"),
		}).Error
}

// AddFrames atomically increments the frame count of a hive by delta.
func (r *HiveRepository) AddFrames(ctx context.Context, hiveID int64, delta int) error {
	return r.db.WithContext(ctx).
		Model(&model.Hive{}).
		Where("id = ?", hiveID).
		Updates(map[string]any{
			"frames":     gorm.Expr("frames + ?", delta),
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *HiveRepository) Delete(ctx context.Context, hiveID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", hiveID).
		Delete(&model.Hive{}).Error
}

func (r *HiveRepository) Move(ctx context.Context, hiveID int64, row, col int) error {
	return r.db.WithContext(ctx).
		Model(&model.Hive{}).
		Where("id = ?", hiveID).
		Updates(map[string]any{
			"grid_row":   row,
			"grid_col":   col,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// Relocate moves a hive to a different apiary at the given grid position.
func (r *HiveRepository) Relocate(ctx context.Context, hiveID, newApiaryID int64, row, col int) error {
	return r.db.WithContext(ctx).
		Model(&model.Hive{}).
		Where("id = ?", hiveID).
		Updates(map[string]any{
			"apiary_id":  newApiaryID,
			"grid_row":   row,
			"grid_col":   col,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// CreateDisease inserts a disease record linked to a hive.
func (r *HiveRepository) CreateDisease(ctx context.Context, d *model.HiveDisease) error {
	return r.db.WithContext(ctx).Create(d).Error
}

// DeleteDisease removes the disease with the given id that belongs to hiveID.
func (r *HiveRepository) DeleteDisease(ctx context.Context, diseaseID, hiveID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND hive_id = ?", diseaseID, hiveID).
		Delete(&model.HiveDisease{}).Error
}

// GetDiseaseByID returns the hive disease with the given id.
func (r *HiveRepository) GetDiseaseByID(ctx context.Context, diseaseID, hiveID int64) (*model.HiveDisease, error) {
	var d model.HiveDisease
	err := r.db.WithContext(ctx).
		Where("id = ? AND hive_id = ?", diseaseID, hiveID).
		First(&d).Error
	return &d, err
}

// ListDiseasesByHiveID returns all diseases for a hive ordered by id.
func (r *HiveRepository) ListDiseasesByHiveID(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error) {
	var diseases []*model.HiveDisease
	err := r.db.WithContext(ctx).
		Where("hive_id = ?", hiveID).
		Order("id ASC").
		Find(&diseases).Error
	return diseases, err
}

// ListDiseasesByHiveIDs returns all diseases for the given set of hive IDs.
func (r *HiveRepository) ListDiseasesByHiveIDs(ctx context.Context, ids []int64) ([]*model.HiveDisease, error) {
	if len(ids) == 0 {
		return []*model.HiveDisease{}, nil
	}
	var diseases []*model.HiveDisease
	err := r.db.WithContext(ctx).
		Where("hive_id IN ?", ids).
		Order("hive_id ASC, id ASC").
		Find(&diseases).Error
	return diseases, err
}
