package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
)

var (
	ErrInvalidGridSize = errors.New("grid rows and cols must be at least 1")
	ErrNameRequired    = errors.New("name is required")
)

type ApiaryRepository interface {
	Create(ctx context.Context, a *model.Apiary, ownerRole string) error
}

type ApiaryService struct {
	apiaries ApiaryRepository
}

func NewApiaryService(apiaries ApiaryRepository) *ApiaryService {
	return &ApiaryService{apiaries: apiaries}
}

func (s *ApiaryService) Create(ctx context.Context, ownerID int64, name string, lat, lng *float64, gridRows, gridCols int) (*model.Apiary, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if gridRows < 1 || gridCols < 1 {
		return nil, ErrInvalidGridSize
	}

	a := &model.Apiary{
		OwnerUserID: ownerID,
		Name:        name,
		Lat:         lat,
		Lng:         lng,
		GridRows:    gridRows,
		GridCols:    gridCols,
	}

	if err := s.apiaries.Create(ctx, a, "owner"); err != nil {
		return nil, fmt.Errorf("create apiary: %w", err)
	}

	return a, nil
}
