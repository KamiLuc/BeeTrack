package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"gorm.io/gorm"
)

// maxApiaryGridDimension is the largest number of rows or columns an apiary grid may have.
const maxApiaryGridDimension = 25

var (
	ErrApiaryNotFound   = errors.New("apiary not found")
	ErrForbidden        = errors.New("forbidden")
	ErrGridTooSmall     = errors.New("grid is too small to fit all existing hives")
	ErrInvalidGPS       = errors.New("lat must be between -90 and 90, lng must be between -180 and 180")
	ErrInvalidGridSize  = errors.New("grid rows and cols must be at least 1")
	ErrGridSizeTooLarge = fmt.Errorf("grid rows and cols must be at most %d", maxApiaryGridDimension)
	ErrNameRequired     = errors.New("name is required")
	ErrNameTooLong      = fmt.Errorf("name must be at most %d characters", validation.Small.MaxLength())
)

// validGPS reports whether lat/lng are unset or within valid GPS coordinate bounds.
func validGPS(lat, lng *float64) bool {
	if lat != nil && (*lat < -90 || *lat > 90) {
		return false
	}
	if lng != nil && (*lng < -180 || *lng > 180) {
		return false
	}
	return true
}

type ApiaryRepository interface {
	Create(ctx context.Context, a *model.Apiary, ownerRole string) error
	DeepCopy(ctx context.Context, sourceID, ownerID int64, newName string) (*model.Apiary, error)
	Delete(ctx context.Context, apiaryID int64) error
	GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.ApiaryMembership, error)
	Update(ctx context.Context, a *model.Apiary) error
}

type HiveRelocator interface {
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
	Move(ctx context.Context, hiveID int64, row, col int) error
}

type ApiaryService struct {
	apiaries ApiaryRepository
	hives    HiveRelocator
}

func NewApiaryService(apiaries ApiaryRepository, hives HiveRelocator) *ApiaryService {
	return &ApiaryService{apiaries: apiaries, hives: hives}
}

// Create creates a new apiary and assigns the given user as its owner.
func (s *ApiaryService) Create(ctx context.Context, ownerID int64, name string, lat, lng *float64, gridRows, gridCols int) (*model.Apiary, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if validation.TooLong(name, validation.Small) {
		return nil, ErrNameTooLong
	}
	if gridRows < 1 || gridCols < 1 {
		return nil, ErrInvalidGridSize
	}
	if gridRows > maxApiaryGridDimension || gridCols > maxApiaryGridDimension {
		return nil, ErrGridSizeTooLarge
	}
	if !validGPS(lat, lng) {
		return nil, ErrInvalidGPS
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

// List returns all apiaries the given user is a member of, along with their role in each.
func (s *ApiaryService) List(ctx context.Context, userID int64) ([]model.ApiaryMembership, error) {
	memberships, err := s.apiaries.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list apiaries: %w", err)
	}
	return memberships, nil
}

// Update updates an apiary's fields; returns ErrForbidden if the user is not the owner.
// If the new grid is smaller than the number of existing hives, ErrGridTooSmall is returned.
// Hives outside the new bounds are automatically relocated to free cells within the new grid.
func (s *ApiaryService) Update(ctx context.Context, userID, apiaryID int64, name string, lat, lng *float64, gridRows, gridCols int) (*model.Apiary, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if validation.TooLong(name, validation.Small) {
		return nil, ErrNameTooLong
	}
	if gridRows < 1 || gridCols < 1 {
		return nil, ErrInvalidGridSize
	}
	if gridRows > maxApiaryGridDimension || gridCols > maxApiaryGridDimension {
		return nil, ErrGridSizeTooLarge
	}
	if !validGPS(lat, lng) {
		return nil, ErrInvalidGPS
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

	hives, err := s.hives.ListByApiaryID(ctx, apiaryID)
	if err != nil {
		return nil, fmt.Errorf("list hives: %w", err)
	}

	if gridRows*gridCols < len(hives) {
		return nil, ErrGridTooSmall
	}

	occupied := make(map[[2]int]bool, len(hives))
	for _, h := range hives {
		if h.GridRow < gridRows && h.GridCol < gridCols {
			occupied[[2]int{h.GridRow, h.GridCol}] = true
		}
	}

	for _, h := range hives {
		if h.GridRow >= gridRows || h.GridCol >= gridCols {
			row, col := firstFreeCell(gridRows, gridCols, occupied)
			if err := s.hives.Move(ctx, h.ID, row, col); err != nil {
				return nil, fmt.Errorf("move hive: %w", err)
			}
			occupied[[2]int{row, col}] = true
		}
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

func firstFreeCell(rows, cols int, occupied map[[2]int]bool) (row, col int) {
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if !occupied[[2]int{r, c}] {
				return r, c
			}
		}
	}
	return 0, 0
}

// Copy creates a deep copy of an apiary the user is a member of. The copy is owned by the
// requesting user and includes all hives, hive diseases, inspections, and inspection diseases.
// Members, invitations, and inspection images are not copied. newName is used as-is if non-empty;
// otherwise falls back to the source name suffixed with " (copy)".
func (s *ApiaryService) Copy(ctx context.Context, userID, apiaryID int64, newName string) (*model.Apiary, error) {
	apiary, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	if newName != "" && validation.TooLong(newName, validation.Small) {
		return nil, ErrNameTooLong
	}
	if newName == "" {
		newName = apiary.Name + " (copy)"
	}

	result, err := s.apiaries.DeepCopy(ctx, apiaryID, userID, newName)
	if err != nil {
		return nil, fmt.Errorf("copy apiary: %w", err)
	}
	return result, nil
}

// Delete deletes an apiary; returns ErrForbidden if the user is not the owner.
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
