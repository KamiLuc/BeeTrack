package repository

import (
	"context"
	"time"

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
	type row struct {
		model.Inspection
		InspectorName string `gorm:"column:inspected_by_name"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("inspections i").
		Select("i.*, u.name AS inspected_by_name").
		Joins("LEFT JOIN users u ON u.id = i.inspected_by").
		Where("i.id = ? AND i.hive_id = ?", inspectionID, hiveID).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	insp := result.Inspection
	insp.InspectedByName = result.InspectorName
	return &insp, nil
}

// CountByHiveID returns the total number of inspections for hiveID.
func (r *InspectionRepository) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Inspection{}).
		Where("hive_id = ?", hiveID).
		Count(&count).Error
	return count, err
}

// ListByHiveID returns inspections for hiveID ordered by inspected_at descending with pagination.
func (r *InspectionRepository) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Inspection, error) {
	type row struct {
		model.Inspection
		InspectorName string `gorm:"column:inspected_by_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("inspections i").
		Select("i.*, u.name AS inspected_by_name").
		Joins("LEFT JOIN users u ON u.id = i.inspected_by").
		Where("i.hive_id = ?", hiveID).
		Order("i.inspected_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	inspections := make([]*model.Inspection, len(rows))
	for i, r := range rows {
		insp := r.Inspection
		insp.InspectedByName = r.InspectorName
		inspections[i] = &insp
	}
	return inspections, nil
}

// Update persists all mutable fields of insp.
func (r *InspectionRepository) Update(ctx context.Context, insp *model.Inspection) error {
	return r.db.WithContext(ctx).
		Model(insp).
		Updates(map[string]any{
			"inspected_at":             insp.InspectedAt,
			"queen_status":             insp.QueenStatus,
			"brood_pattern":            insp.BroodPattern,
			"frames_brood":             insp.FramesBrood,
			"frames_honey":             insp.FramesHoney,
			"frames_pollen":            insp.FramesPollen,
			"queen_cells_count":        insp.QueenCellsCount,
			"aggressiveness":           insp.Aggressiveness,
			"frames_added_foundation":  insp.FramesAddedFoundation,
			"frames_added_drawn":       insp.FramesAddedDrawn,
			"frames_added_honey":       insp.FramesAddedHoney,
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

// LastInspectionDatesByHiveIDs returns the latest inspected_at per hive ID for the given set of hive IDs.
func (r *InspectionRepository) LastInspectionDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error) {
	if len(ids) == 0 {
		return map[int64]*time.Time{}, nil
	}
	type row struct {
		HiveID      int64
		InspectedAt time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&model.Inspection{}).
		Select("hive_id, MAX(inspected_at) AS inspected_at").
		Where("hive_id IN ?", ids).
		Group("hive_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]*time.Time, len(rows))
	for _, r := range rows {
		t := r.InspectedAt
		out[r.HiveID] = &t
	}
	return out, nil
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
