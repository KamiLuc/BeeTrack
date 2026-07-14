package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"gorm.io/gorm"
)

var (
	ErrDuplicateHiveName   = errors.New("a hive with this name already exists in this apiary")
	ErrHiveDiseaseNotFound = errors.New("hive disease not found")
	ErrHiveNotFound        = errors.New("hive not found")
	ErrHiveTypeTooLong     = fmt.Errorf("type must be at most %d characters", validation.Small.MaxLength())
	ErrInvalidGridPosition = errors.New("grid position out of apiary bounds")
	ErrPositionOccupied    = errors.New("grid position already occupied")
	ErrSameApiary          = errors.New("source and target apiary are the same")
	ErrTargetApiaryFull    = errors.New("target apiary has no free space")
)

var validHiveDiseases = map[string]bool{
	"american_foulbrood": true, "chalkbrood": true, "dwv": true,
	"european_foulbrood": true, "laying_workers": true, "nosema": true, "varroa": true,
}

type ApiaryMembershipReader interface {
	GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error)
}

type HiveRepository interface {
	Create(ctx context.Context, h *model.Hive) error
	CreateDisease(ctx context.Context, d *model.HiveDisease) error
	Delete(ctx context.Context, hiveID int64) error
	DeleteDisease(ctx context.Context, diseaseID, hiveID int64) error
	ExistsByName(ctx context.Context, apiaryID int64, name string, excludeHiveID int64) (bool, error)
	GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error)
	GetDiseaseByID(ctx context.Context, diseaseID, hiveID int64) (*model.HiveDisease, error)
	IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error)
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
	ListDiseasesByHiveID(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error)
	ListDiseasesByHiveIDs(ctx context.Context, ids []int64) ([]*model.HiveDisease, error)
	Move(ctx context.Context, hiveID int64, row, col int) error
	Relocate(ctx context.Context, hiveID, newApiaryID int64, row, col int) error
	Update(ctx context.Context, h *model.Hive) error
}

type HiveService struct {
	apiaries ApiaryMembershipReader
	hives    HiveRepository
}

func NewHiveService(apiaries ApiaryMembershipReader, hives HiveRepository) *HiveService {
	return &HiveService{apiaries: apiaries, hives: hives}
}

func (s *HiveService) List(ctx context.Context, userID, apiaryID int64) ([]*model.Hive, error) {
	_, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	hives, err := s.hives.ListByApiaryID(ctx, apiaryID)
	if err != nil {
		return nil, fmt.Errorf("list hives: %w", err)
	}

	return hives, nil
}

// Update modifies a hive's name, type, and status fields. The caller must be an apiary member.
func (s *HiveService) Update(ctx context.Context, userID, apiaryID, hiveID int64, name, hiveType string, active, readyForHarvest, queenless bool) (*model.Hive, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if validation.TooLong(name, validation.Small) {
		return nil, ErrNameTooLong
	}
	if validation.TooLong(hiveType, validation.Small) {
		return nil, ErrHiveTypeTooLong
	}

	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	hive, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get hive: %w", err)
	}

	duplicate, err := s.hives.ExistsByName(ctx, apiaryID, name, hiveID)
	if err != nil {
		return nil, fmt.Errorf("check hive name: %w", err)
	}
	if duplicate {
		return nil, ErrDuplicateHiveName
	}

	hive.Name = name
	hive.Type = hiveType
	hive.Active = active
	hive.ReadyForHarvest = readyForHarvest
	hive.Queenless = queenless

	if err := s.hives.Update(ctx, hive); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateHiveName
		}
		return nil, fmt.Errorf("update hive: %w", err)
	}

	return hive, nil
}

func (s *HiveService) Delete(ctx context.Context, userID, apiaryID, hiveID int64) error {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApiaryNotFound
		}
		return fmt.Errorf("get apiary: %w", err)
	}

	if _, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHiveNotFound
		}
		return fmt.Errorf("get hive: %w", err)
	}

	if err := s.hives.Delete(ctx, hiveID); err != nil {
		return fmt.Errorf("delete hive: %w", err)
	}

	return nil
}

func (s *HiveService) Get(ctx context.Context, userID, apiaryID, hiveID int64) (*model.Hive, error) {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	hive, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get hive: %w", err)
	}

	return hive, nil
}

func (s *HiveService) Move(ctx context.Context, userID, apiaryID, hiveID int64, gridRow, gridCol int) (*model.Hive, error) {
	apiary, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	hive, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get hive: %w", err)
	}

	if gridRow < 0 || gridRow >= apiary.GridRows || gridCol < 0 || gridCol >= apiary.GridCols {
		return nil, ErrInvalidGridPosition
	}

	if hive.GridRow != gridRow || hive.GridCol != gridCol {
		occupied, err := s.hives.IsPositionOccupied(ctx, apiaryID, gridRow, gridCol)
		if err != nil {
			return nil, fmt.Errorf("check position: %w", err)
		}
		if occupied {
			return nil, ErrPositionOccupied
		}

		if err := s.hives.Move(ctx, hiveID, gridRow, gridCol); err != nil {
			return nil, fmt.Errorf("move hive: %w", err)
		}
		hive.GridRow = gridRow
		hive.GridCol = gridCol
	}

	return hive, nil
}

// DiseasesByHive returns diseases for a hive. Caller must have already verified access.
func (s *HiveService) DiseasesByHive(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error) {
	diseases, err := s.hives.ListDiseasesByHiveID(ctx, hiveID)
	if err != nil {
		return nil, fmt.Errorf("list hive diseases: %w", err)
	}
	return diseases, nil
}

// DiseasesForHives returns diseases grouped by hive ID. Caller must have already verified access.
func (s *HiveService) DiseasesForHives(ctx context.Context, ids []int64) (map[int64][]*model.HiveDisease, error) {
	rows, err := s.hives.ListDiseasesByHiveIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list hive diseases: %w", err)
	}
	grouped := make(map[int64][]*model.HiveDisease, len(ids))
	for _, id := range ids {
		grouped[id] = []*model.HiveDisease{}
	}
	for _, d := range rows {
		grouped[d.HiveID] = append(grouped[d.HiveID], d)
	}
	return grouped, nil
}

// AddDisease validates the disease, verifies access, and attaches it to a hive.
func (s *HiveService) AddDisease(ctx context.Context, userID, apiaryID, hiveID int64, disease string) (*model.HiveDisease, error) {
	if !validHiveDiseases[disease] {
		return nil, ErrInvalidDisease
	}
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}
	if _, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get hive: %w", err)
	}
	d := &model.HiveDisease{HiveID: hiveID, Disease: disease}
	if err := s.hives.CreateDisease(ctx, d); err != nil {
		return nil, fmt.Errorf("create hive disease: %w", err)
	}
	return d, nil
}

// RemoveDisease verifies access and deletes a hive disease record.
func (s *HiveService) RemoveDisease(ctx context.Context, userID, apiaryID, hiveID, diseaseID int64) error {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApiaryNotFound
		}
		return fmt.Errorf("get apiary: %w", err)
	}
	if _, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, apiaryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHiveNotFound
		}
		return fmt.Errorf("get hive: %w", err)
	}
	if _, err := s.hives.GetDiseaseByID(ctx, diseaseID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHiveDiseaseNotFound
		}
		return fmt.Errorf("get hive disease: %w", err)
	}
	if err := s.hives.DeleteDisease(ctx, diseaseID, hiveID); err != nil {
		return fmt.Errorf("delete hive disease: %w", err)
	}
	return nil
}

// ChangeApiary moves a hive from its current apiary to a different one, placing it at the
// first available grid position in the target apiary. Both apiaries must be accessible to userID.
func (s *HiveService) ChangeApiary(ctx context.Context, userID, srcApiaryID, hiveID, dstApiaryID int64) (*model.Hive, error) {
	if srcApiaryID == dstApiaryID {
		return nil, ErrSameApiary
	}

	if _, _, err := s.apiaries.GetMembership(ctx, srcApiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get source apiary: %w", err)
	}

	hive, err := s.hives.GetByIDAndApiaryID(ctx, hiveID, srcApiaryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get hive: %w", err)
	}

	target, _, err := s.apiaries.GetMembership(ctx, dstApiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get target apiary: %w", err)
	}

	existing, err := s.hives.ListByApiaryID(ctx, dstApiaryID)
	if err != nil {
		return nil, fmt.Errorf("list target hives: %w", err)
	}

	if len(existing) >= target.GridRows*target.GridCols {
		return nil, ErrTargetApiaryFull
	}

	duplicate, err := s.hives.ExistsByName(ctx, dstApiaryID, hive.Name, 0)
	if err != nil {
		return nil, fmt.Errorf("check hive name: %w", err)
	}
	if duplicate {
		return nil, ErrDuplicateHiveName
	}

	occupied := make(map[[2]int]bool, len(existing))
	for _, h := range existing {
		occupied[[2]int{h.GridRow, h.GridCol}] = true
	}
	row, col := firstFreeCell(target.GridRows, target.GridCols, occupied)

	if err := s.hives.Relocate(ctx, hiveID, dstApiaryID, row, col); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateHiveName
		}
		return nil, fmt.Errorf("relocate hive: %w", err)
	}

	hive.ApiaryID = dstApiaryID
	hive.GridRow = row
	hive.GridCol = col

	return hive, nil
}

// Add creates a hive at the given grid position within an apiary. The caller must be an apiary member.
func (s *HiveService) Add(ctx context.Context, userID, apiaryID int64, name, hiveType string, active, queenless, readyForHarvest bool, gridRow, gridCol int) (*model.Hive, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if validation.TooLong(name, validation.Small) {
		return nil, ErrNameTooLong
	}
	if validation.TooLong(hiveType, validation.Small) {
		return nil, ErrHiveTypeTooLong
	}

	apiary, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	if gridRow < 0 || gridRow >= apiary.GridRows || gridCol < 0 || gridCol >= apiary.GridCols {
		return nil, ErrInvalidGridPosition
	}

	occupied, err := s.hives.IsPositionOccupied(ctx, apiaryID, gridRow, gridCol)
	if err != nil {
		return nil, fmt.Errorf("check position: %w", err)
	}
	if occupied {
		return nil, ErrPositionOccupied
	}

	duplicate, err := s.hives.ExistsByName(ctx, apiaryID, name, 0)
	if err != nil {
		return nil, fmt.Errorf("check hive name: %w", err)
	}
	if duplicate {
		return nil, ErrDuplicateHiveName
	}

	h := &model.Hive{
		ApiaryID:        apiaryID,
		Name:            name,
		Type:            hiveType,
		Active:          active,
		Queenless:       queenless,
		ReadyForHarvest: readyForHarvest,
		GridRow:         gridRow,
		GridCol:         gridCol,
	}

	if err := s.hives.Create(ctx, h); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateHiveName
		}
		return nil, fmt.Errorf("create hive: %w", err)
	}

	return h, nil
}
