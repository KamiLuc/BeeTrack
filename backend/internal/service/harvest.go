package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrHarvestNotFound        = errors.New("harvest not found")
	ErrHarvestedAtRequired    = errors.New("harvested_at is required")
	ErrHarvestFramesRequired  = errors.New("frames or half_frames must be greater than zero")
	ErrHarvestKilogramsRequired = errors.New("kilograms must be greater than zero")
)

// HarvestRepository is the persistence interface for harvests.
type HarvestRepository interface {
	Create(ctx context.Context, h *model.Harvest) error
	Delete(ctx context.Context, harvestID int64) error
	GetByID(ctx context.Context, harvestID, hiveID int64) (*model.Harvest, error)
	CountByHiveID(ctx context.Context, hiveID int64) (int64, error)
	ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Harvest, error)
	Update(ctx context.Context, h *model.Harvest) error
}

// HarvestService handles business logic for harvest records.
type HarvestService struct {
	apiaries ApiaryMembershipReader
	hives    InspectionHiveReader
	harvests HarvestRepository
}

// NewHarvestService creates a HarvestService with the given dependencies.
func NewHarvestService(apiaries ApiaryMembershipReader, hives InspectionHiveReader, harvests HarvestRepository) *HarvestService {
	return &HarvestService{apiaries: apiaries, hives: hives, harvests: harvests}
}

// HarvestParams holds the mutable fields for create and update operations.
type HarvestParams struct {
	HarvestedAt time.Time
	Frames      int
	HalfFrames  int
	Kilograms   float64
	Notes       string
}

func validateHarvestParams(p HarvestParams) error {
	if p.HarvestedAt.IsZero() {
		return ErrHarvestedAtRequired
	}
	if p.Frames == 0 && p.HalfFrames == 0 {
		return ErrHarvestFramesRequired
	}
	if p.Kilograms <= 0 {
		return ErrHarvestKilogramsRequired
	}
	return nil
}

func (s *HarvestService) checkAccess(ctx context.Context, apiaryID, userID, hiveID int64) error {
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
	return nil
}

// Create validates params, checks membership, and inserts a new harvest.
func (s *HarvestService) Create(ctx context.Context, userID, apiaryID, hiveID int64, params HarvestParams) (*model.Harvest, error) {
	if err := validateHarvestParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	h := &model.Harvest{
		HiveID:      hiveID,
		HarvestedBy: userID,
		HarvestedAt: params.HarvestedAt,
		Frames:      params.Frames,
		HalfFrames:  params.HalfFrames,
		Kilograms:   params.Kilograms,
		Notes:       params.Notes,
	}
	if err := s.harvests.Create(ctx, h); err != nil {
		return nil, fmt.Errorf("create harvest: %w", err)
	}
	return h, nil
}

// Get returns a single harvest, verifying apiary membership and hive ownership.
func (s *HarvestService) Get(ctx context.Context, userID, apiaryID, hiveID, harvestID int64) (*model.Harvest, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	h, err := s.harvests.GetByID(ctx, harvestID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHarvestNotFound
		}
		return nil, fmt.Errorf("get harvest: %w", err)
	}
	return h, nil
}

// List returns a paginated slice of harvests and the total count for a hive.
func (s *HarvestService) List(ctx context.Context, userID, apiaryID, hiveID int64, limit, offset int) ([]*model.Harvest, int64, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, 0, err
	}
	total, err := s.harvests.CountByHiveID(ctx, hiveID)
	if err != nil {
		return nil, 0, fmt.Errorf("count harvests: %w", err)
	}
	harvests, err := s.harvests.ListByHiveID(ctx, hiveID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list harvests: %w", err)
	}
	return harvests, total, nil
}

// Update validates params and overwrites all mutable fields of an existing harvest.
func (s *HarvestService) Update(ctx context.Context, userID, apiaryID, hiveID, harvestID int64, params HarvestParams) (*model.Harvest, error) {
	if err := validateHarvestParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	h, err := s.harvests.GetByID(ctx, harvestID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHarvestNotFound
		}
		return nil, fmt.Errorf("get harvest: %w", err)
	}
	h.HarvestedAt = params.HarvestedAt
	h.Frames = params.Frames
	h.HalfFrames = params.HalfFrames
	h.Kilograms = params.Kilograms
	h.Notes = params.Notes
	if err := s.harvests.Update(ctx, h); err != nil {
		return nil, fmt.Errorf("update harvest: %w", err)
	}
	return h, nil
}

// Delete removes a harvest after verifying membership.
func (s *HarvestService) Delete(ctx context.Context, userID, apiaryID, hiveID, harvestID int64) error {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return err
	}
	if _, err := s.harvests.GetByID(ctx, harvestID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHarvestNotFound
		}
		return fmt.Errorf("get harvest: %w", err)
	}
	if err := s.harvests.Delete(ctx, harvestID); err != nil {
		return fmt.Errorf("delete harvest: %w", err)
	}
	return nil
}
