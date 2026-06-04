package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type InspectionRepository struct {
	db *gorm.DB
}

// NewInspectionRepository creates a new InspectionRepository backed by db.
func NewInspectionRepository(db *gorm.DB) *InspectionRepository {
	return &InspectionRepository{db: db}
}

// Create inserts a new inspection record.
func (r *InspectionRepository) Create(ctx context.Context, insp *model.Inspection) error {
	return r.db.WithContext(ctx).Create(insp).Error
}

// GetByID returns the inspection with the given id that belongs to hiveID.
func (r *InspectionRepository) GetByID(ctx context.Context, inspectionID, hiveID int64) (*model.Inspection, error) {
	var insp model.Inspection
	err := r.db.WithContext(ctx).
		Where("id = ? AND hive_id = ?", inspectionID, hiveID).
		First(&insp).Error
	return &insp, err
}

// ListByHiveID returns inspections for hiveID ordered by inspected_at descending with pagination.
func (r *InspectionRepository) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Inspection, error) {
	var inspections []*model.Inspection
	err := r.db.WithContext(ctx).
		Where("hive_id = ?", hiveID).
		Order("inspected_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&inspections).Error
	return inspections, err
}

// Update persists all mutable fields of insp.
func (r *InspectionRepository) Update(ctx context.Context, insp *model.Inspection) error {
	return r.db.WithContext(ctx).
		Model(insp).
		Updates(map[string]any{
			"inspected_at":             insp.InspectedAt,
			"queen_status":             insp.QueenStatus,
			"brood_pattern":            insp.BroodPattern,
			"frames_honey":             insp.FramesHoney,
			"frames_pollen":            insp.FramesPollen,
			"varroa_count":             insp.VarroaCount,
			"queen_cells_count":        insp.QueenCellsCount,
			"aggressiveness":           insp.Aggressiveness,
			"frames_added_foundation":  insp.FramesAddedFoundation,
			"frames_added_drawn":       insp.FramesAddedDrawn,
			"queen_added":              insp.QueenAdded,
			"notes":                    insp.Notes,
			"updated_at":               gorm.Expr("NOW()"),
		}).Error
}

// Delete removes the inspection with the given id.
func (r *InspectionRepository) Delete(ctx context.Context, inspectionID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", inspectionID).
		Delete(&model.Inspection{}).Error
}

// CreateDisease inserts a new disease record linked to an inspection.
func (r *InspectionRepository) CreateDisease(ctx context.Context, d *model.InspectionDisease) error {
	return r.db.WithContext(ctx).Create(d).Error
}

// DeleteDisease removes the disease with the given id that belongs to inspectionID.
func (r *InspectionRepository) DeleteDisease(ctx context.Context, diseaseID, inspectionID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND inspection_id = ?", diseaseID, inspectionID).
		Delete(&model.InspectionDisease{}).Error
}

// GetDiseaseByID returns the disease with the given id that belongs to inspectionID.
func (r *InspectionRepository) GetDiseaseByID(ctx context.Context, diseaseID, inspectionID int64) (*model.InspectionDisease, error) {
	var d model.InspectionDisease
	err := r.db.WithContext(ctx).
		Where("id = ? AND inspection_id = ?", diseaseID, inspectionID).
		First(&d).Error
	return &d, err
}

// ListDiseasesByInspectionID returns all diseases for the given inspectionID ordered by id.
func (r *InspectionRepository) ListDiseasesByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionDisease, error) {
	var diseases []*model.InspectionDisease
	err := r.db.WithContext(ctx).
		Where("inspection_id = ?", inspectionID).
		Order("id ASC").
		Find(&diseases).Error
	return diseases, err
}

// ListDiseasesByInspectionIDs returns all diseases for the given set of inspection IDs.
func (r *InspectionRepository) ListDiseasesByInspectionIDs(ctx context.Context, ids []int64) ([]*model.InspectionDisease, error) {
	if len(ids) == 0 {
		return []*model.InspectionDisease{}, nil
	}
	var diseases []*model.InspectionDisease
	err := r.db.WithContext(ctx).
		Where("inspection_id IN ?", ids).
		Order("inspection_id ASC, id ASC").
		Find(&diseases).Error
	return diseases, err
}
