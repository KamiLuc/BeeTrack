package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrApiaryNotFound  = errors.New("apiary not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidGridSize = errors.New("grid rows and cols must be at least 1")
	ErrNameRequired    = errors.New("name is required")
)

type ApiaryRepository interface {
	Create(ctx context.Context, a *model.Apiary, ownerRole string) error
	Delete(ctx context.Context, apiaryID int64) error
	GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.ApiaryMembership, error)
	Update(ctx context.Context, a *model.Apiary) error
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

func (s *ApiaryService) List(ctx context.Context, userID int64) ([]model.ApiaryMembership, error) {
	memberships, err := s.apiaries.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list apiaries: %w", err)
	}
	return memberships, nil
}

func (s *ApiaryService) Update(ctx context.Context, userID, apiaryID int64, name string, lat, lng *float64, gridRows, gridCols int) (*model.Apiary, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if gridRows < 1 || gridCols < 1 {
		return nil, ErrInvalidGridSize
	}

	apiary, role, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}
	if role != "owner" {
		return nil, ErrForbidden
	}

	apiary.Name = name
	apiary.Lat = lat
	apiary.Lng = lng
	apiary.GridRows = gridRows
	apiary.GridCols = gridCols

	if err := s.apiaries.Update(ctx, apiary); err != nil {
		return nil, fmt.Errorf("update apiary: %w", err)
	}

	return apiary, nil
}

func (s *ApiaryService) Delete(ctx context.Context, userID, apiaryID int64) error {
	_, role, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApiaryNotFound
		}
		return fmt.Errorf("get apiary: %w", err)
	}
	if role != "owner" {
		return ErrForbidden
	}

	if err := s.apiaries.Delete(ctx, apiaryID); err != nil {
		return fmt.Errorf("delete apiary: %w", err)
	}

	return nil
}
