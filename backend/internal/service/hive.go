package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrInvalidGridPosition = errors.New("grid position out of apiary bounds")
	ErrPositionOccupied    = errors.New("grid position already occupied")
)

type ApiaryMembershipReader interface {
	GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error)
}

type HiveRepository interface {
	Create(ctx context.Context, h *model.Hive) error
	IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error)
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
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
