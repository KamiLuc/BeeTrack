package repository

import (
	"context"
	"errors"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type ApiaryRepository struct {
	db *gorm.DB
}

func NewApiaryRepository(db *gorm.DB) *ApiaryRepository {
	return &ApiaryRepository{db: db}
}

// Create inserts a new apiary and adds the owner as a member in a single transaction.
func (r *ApiaryRepository) Create(ctx context.Context, a *model.Apiary, ownerRole string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
		member := &model.ApiaryMember{
			ApiaryID: a.ID,
			UserID:   a.OwnerUserID,
			Role:     ownerRole,
		}
		return tx.Create(member).Error
	})
}

// ListByUserID returns all apiaries the user belongs to, ordered by creation date descending.
func (r *ApiaryRepository) ListByUserID(ctx context.Context, userID int64) ([]model.ApiaryMembership, error) {
	type row struct {
		model.Apiary
		HiveCount       int
		UserRole        string
		LastInspectedAt *time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("apiaries a").
		Select("a.*, am.role AS user_role, " +
			"(SELECT COUNT(*) FROM hives h WHERE h.apiary_id = a.id) AS hive_count, " +
			"(SELECT MAX(i.inspected_at) FROM inspections i JOIN hives h ON h.id = i.hive_id WHERE h.apiary_id = a.id) AS last_inspected_at").
		Joins("JOIN apiary_members am ON am.apiary_id = a.id").
		Where("am.user_id = ?", userID).
		Order("a.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	memberships := make([]model.ApiaryMembership, len(rows))
	for i, row := range rows {
		a := row.Apiary
		memberships[i] = model.ApiaryMembership{Apiary: &a, HiveCount: row.HiveCount, UserRole: row.UserRole, LastInspectedAt: row.LastInspectedAt}
	}
	return memberships, nil
}

// GetMembership returns the apiary and the user's role in it; returns gorm.ErrRecordNotFound if not a member.
func (r *ApiaryRepository) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	type row struct {
		model.Apiary
		UserRole string
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("apiaries a").
		Select("a.*, am.role AS user_role").
		Joins("JOIN apiary_members am ON am.apiary_id = a.id").
		Where("a.id = ? AND am.user_id = ?", apiaryID, userID).
		First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}
	if err != nil {
		return nil, "", err
	}
	a := result.Apiary
	return &a, result.UserRole, nil
}

// Update saves changes to an apiary's editable fields.
func (r *ApiaryRepository) Update(ctx context.Context, a *model.Apiary) error {
	return r.db.WithContext(ctx).
		Model(a).
		Updates(map[string]any{
			"name":       a.Name,
			"lat":        a.Lat,
			"lng":        a.Lng,
			"grid_rows":  a.GridRows,
			"grid_cols":  a.GridCols,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// DeepCopy creates a new apiary owned by ownerID, copying all hives, hive diseases,
// inspections, and inspection diseases from the source apiary. Members, invitations,
// and inspection images are not copied.
func (r *ApiaryRepository) DeepCopy(ctx context.Context, sourceID, ownerID int64, newName string) (*model.Apiary, error) {
	var result *model.Apiary
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var source model.Apiary
		if err := tx.Where("id = ?", sourceID).First(&source).Error; err != nil {
			return err
		}

		newApiary := &model.Apiary{
			OwnerUserID: ownerID,
			Name:        newName,
			Lat:         source.Lat,
			Lng:         source.Lng,
			GridRows:    source.GridRows,
			GridCols:    source.GridCols,
		}
		if err := tx.Create(newApiary).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.ApiaryMember{ApiaryID: newApiary.ID, UserID: ownerID, Role: "owner"}).Error; err != nil {
			return err
		}

		var hives []*model.Hive
		if err := tx.Where("apiary_id = ?", sourceID).Find(&hives).Error; err != nil {
			return err
		}
		if len(hives) == 0 {
			result = newApiary
			return nil
		}

		sourceHiveIDs := make([]int64, len(hives))
		for i, h := range hives {
			sourceHiveIDs[i] = h.ID
		}

		var hiveDiseases []*model.HiveDisease
		if err := tx.Where("hive_id IN ?", sourceHiveIDs).Find(&hiveDiseases).Error; err != nil {
			return err
		}
		hiveDiseasesByHiveID := make(map[int64][]*model.HiveDisease, len(hiveDiseases))
		for _, d := range hiveDiseases {
			hiveDiseasesByHiveID[d.HiveID] = append(hiveDiseasesByHiveID[d.HiveID], d)
		}

		var inspections []*model.Inspection
		if err := tx.Where("hive_id IN ?", sourceHiveIDs).Find(&inspections).Error; err != nil {
			return err
		}
		inspsByHiveID := make(map[int64][]*model.Inspection, len(inspections))
		for _, insp := range inspections {
			inspsByHiveID[insp.HiveID] = append(inspsByHiveID[insp.HiveID], insp)
		}

		var inspDiseasesByInspID map[int64][]*model.InspectionDisease
		if len(inspections) > 0 {
			sourceInspIDs := make([]int64, len(inspections))
			for i, insp := range inspections {
				sourceInspIDs[i] = insp.ID
			}
			var inspDiseases []*model.InspectionDisease
			if err := tx.Where("inspection_id IN ?", sourceInspIDs).Find(&inspDiseases).Error; err != nil {
				return err
			}
			inspDiseasesByInspID = make(map[int64][]*model.InspectionDisease, len(inspDiseases))
			for _, d := range inspDiseases {
				inspDiseasesByInspID[d.InspectionID] = append(inspDiseasesByInspID[d.InspectionID], d)
			}
		}

		for _, h := range hives {
			newHive := &model.Hive{
				ApiaryID:        newApiary.ID,
				Name:            h.Name,
				Type:            h.Type,
				Active:          h.Active,
				Frames:          h.Frames,
				ReadyForHarvest: h.ReadyForHarvest,
				Queenless:       h.Queenless,
				GridRow:         h.GridRow,
				GridCol:         h.GridCol,
			}
			if err := tx.Create(newHive).Error; err != nil {
				return err
			}
			for _, d := range hiveDiseasesByHiveID[h.ID] {
				if err := tx.Create(&model.HiveDisease{HiveID: newHive.ID, Disease: d.Disease}).Error; err != nil {
					return err
				}
			}
			for _, insp := range inspsByHiveID[h.ID] {
				newInsp := &model.Inspection{
					HiveID:                newHive.ID,
					InspectedBy:           insp.InspectedBy,
					InspectedAt:           insp.InspectedAt,
					QueenStatus:           insp.QueenStatus,
					BroodPattern:          insp.BroodPattern,
					FramesBrood:           insp.FramesBrood,
					FramesFeed:            insp.FramesFeed,
					FramesPollen:          insp.FramesPollen,
					QueenCellsCount:       insp.QueenCellsCount,
					Aggressiveness:        insp.Aggressiveness,
					FramesAddedFoundation: insp.FramesAddedFoundation,
					FramesAddedDrawn:      insp.FramesAddedDrawn,
					FramesAddedBrood:      insp.FramesAddedBrood,
					FramesAddedFeed:       insp.FramesAddedFeed,
					QueenAdded:            insp.QueenAdded,
					Notes:                 insp.Notes,
				}
				if err := tx.Create(newInsp).Error; err != nil {
					return err
				}
				for _, d := range inspDiseasesByInspID[insp.ID] {
					if err := tx.Create(&model.InspectionDisease{
						InspectionID: newInsp.ID,
						Disease:      d.Disease,
						Notes:        d.Notes,
					}).Error; err != nil {
						return err
					}
				}
			}
		}

		result = newApiary
		return nil
	})
	return result, err
}

// Delete soft-deletes an apiary by ID.
func (r *ApiaryRepository) Delete(ctx context.Context, apiaryID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", apiaryID).
		Delete(&model.Apiary{}).Error
}
