package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrHiveNotFound        = errors.New("hive not found")
	ErrInvalidGridPosition = errors.New("grid position out of apiary bounds")
	ErrPositionOccupied    = errors.New("grid position already occupied")
)

type ApiaryMembershipReader interface {
	GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error)
}

type HiveRepository interface {
	Create(ctx context.Context, h *model.Hive) error
	Delete(ctx context.Context, hiveID int64) error
	GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error)
	IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error)
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
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

func (s *HiveService) Update(ctx context.Context, userID, apiaryID, hiveID int64, name, hiveType string, active bool) (*model.Hive, error) {
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

func (s *HiveService) Add(ctx context.Context, userID, apiaryID int64, name, hiveType string, gridRow, gridCol int) (*model.Hive, error) {
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
		ApiaryID: apiaryID,
		Name:     name,
		Type:     hiveType,
		Active:   true,
		GridRow:  gridRow,
		GridCol:  gridCol,
	}

	if err := s.hives.Create(ctx, h); err != nil {
		return nil, fmt.Errorf("create hive: %w", err)
	}

	return h, nil
}
