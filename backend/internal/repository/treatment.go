package repository

import (
	"context"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// TreatmentRepository persists treatment records.
type TreatmentRepository struct {
	db *gorm.DB
}

// NewTreatmentRepository creates a new TreatmentRepository backed by db.
func NewTreatmentRepository(db *gorm.DB) *TreatmentRepository {
	return &TreatmentRepository{db: db}
}

// Create inserts a new treatment record.
func (r *TreatmentRepository) Create(ctx context.Context, t *model.Treatment) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// GetByID returns the treatment with the given id that belongs to hiveID.
func (r *TreatmentRepository) GetByID(ctx context.Context, treatmentID, hiveID int64) (*model.Treatment, error) {
	type row struct {
		model.Treatment
		TreaterName string `gorm:"column:treated_by_name"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("treatments t").
		Select("t.*, u.name AS treated_by_name").
		Joins("LEFT JOIN users u ON u.id = t.treated_by").
		Where("t.id = ? AND t.hive_id = ?", treatmentID, hiveID).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	tr := result.Treatment
	tr.TreatedByName = result.TreaterName
	return &tr, nil
}

// CountByHiveID returns the total number of treatments for hiveID.
func (r *TreatmentRepository) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Where("hive_id = ?", hiveID).
		Count(&count).Error
	return count, err
}

// ListByHiveID returns treatments for hiveID ordered by treated_at descending with pagination.
func (r *TreatmentRepository) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Treatment, error) {
	type row struct {
		model.Treatment
		TreaterName string `gorm:"column:treated_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("treatments t").
		Select("t.*, u.name AS treated_by_name").
		Joins("LEFT JOIN users u ON u.id = t.treated_by").
		Where("t.hive_id = ?", hiveID).
		Order("t.treated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	treatments := make([]*model.Treatment, len(rows))
	for i, r := range rows {
		tr := r.Treatment
		tr.TreatedByName = r.TreaterName
		treatments[i] = &tr
	}
	return treatments, nil
}

// ListByHiveIDsAndRange returns treatments for the given hive IDs with treated_at
// in [from, to), ordered by hive ID then treated_at descending, for report generation.
func (r *TreatmentRepository) ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Treatment, error) {
	if len(hiveIDs) == 0 {
		return []*model.Treatment{}, nil
	}
	type row struct {
		model.Treatment
		TreaterName string `gorm:"column:treated_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("treatments t").
		Select("t.*, u.name AS treated_by_name").
		Joins("LEFT JOIN users u ON u.id = t.treated_by").
		Where("t.hive_id IN ? AND t.treated_at >= ? AND t.treated_at < ?", hiveIDs, from, to).
		Order("t.hive_id ASC, t.treated_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	treatments := make([]*model.Treatment, len(rows))
	for i, r := range rows {
		tr := r.Treatment
		tr.TreatedByName = r.TreaterName
		treatments[i] = &tr
	}
	return treatments, nil
}

// Update persists all mutable fields of t.
func (r *TreatmentRepository) Update(ctx context.Context, t *model.Treatment) error {
	return r.db.WithContext(ctx).
		Model(t).
		Updates(map[string]any{
			"treated_at":    t.TreatedAt,
			"medicine_name": t.MedicineName,
			"dose":          t.Dose,
			"notes":         t.Notes,
			"updated_at":    gorm.Expr("NOW()"),
		}).Error
}

// BulkCreate inserts multiple treatment records in a single transaction.
func (r *TreatmentRepository) BulkCreate(ctx context.Context, treatments []*model.Treatment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, t := range treatments {
			if err := tx.Create(t).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes the treatment with the given id.
func (r *TreatmentRepository) Delete(ctx context.Context, treatmentID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", treatmentID).
		Delete(&model.Treatment{}).Error
}

// DistinctMedicineNames returns the distinct medicine names userID has previously
// entered, most recently used first.
func (r *TreatmentRepository) DistinctMedicineNames(ctx context.Context, userID int64) ([]string, error) {
	names := []string{}
	err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Select("medicine_name").
		Where("treated_by = ? AND medicine_name <> ''", userID).
		Group("medicine_name").
		Order("MAX(treated_at) DESC").
		Pluck("medicine_name", &names).Error
	return names, err
}

// DistinctDoses returns the distinct doses userID has previously entered, most
// recently used first.
func (r *TreatmentRepository) DistinctDoses(ctx context.Context, userID int64) ([]string, error) {
	doses := []string{}
	err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Select("dose").
		Where("treated_by = ? AND dose <> ''", userID).
		Group("dose").
		Order("MAX(treated_at) DESC").
		Pluck("dose", &doses).Error
	return doses, err
}
