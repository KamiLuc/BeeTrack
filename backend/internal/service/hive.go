package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrHiveDiseaseNotFound = errors.New("hive disease not found")
	ErrHiveNotFound        = errors.New("hive not found")
	ErrInvalidGridPosition = errors.New("grid position out of apiary bounds")
	ErrPositionOccupied    = errors.New("grid position already occupied")
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
	GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error)
	GetDiseaseByID(ctx context.Context, diseaseID, hiveID int64) (*model.HiveDisease, error)
	IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error)
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
	ListDiseasesByHiveID(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error)
	ListDiseasesByHiveIDs(ctx context.Context, ids []int64) ([]*model.HiveDisease, error)
	Move(ctx context.Context, hiveID int64, row, col int) error
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

func (s *HiveService) Update(ctx context.Context, userID, apiaryID, hiveID int64, name, hiveType string, active, readyForHarvest, queenless bool) (*model.Hive, error) {
	if name == "" {
		return nil, ErrNameRequired
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

	hive.Name = name
	hive.Type = hiveType
	hive.Active = active
	hive.ReadyForHarvest = readyForHarvest
	hive.Queenless = queenless

	if err := s.hives.Update(ctx, hive); err != nil {
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

func (s *HiveService) Add(ctx context.Context, userID, apiaryID int64, name, hiveType string, active, queenless, readyForHarvest bool, gridRow, gridCol int) (*model.Hive, error) {
	if name == "" {
		return nil, ErrNameRequired
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
		return nil, fmt.Errorf("create hive: %w", err)
	}

	return h, nil
}
