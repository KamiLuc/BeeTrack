package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"gorm.io/gorm"
)

var (
	ErrFeedingNotFound     = errors.New("feeding not found")
	ErrFedAtRequired       = errors.New("fed_at is required")
	ErrFeedTypeRequired    = errors.New("feed_type is required")
	ErrFeedTypeTooLong     = fmt.Errorf("feed_type must be at most %d characters", validation.Small.MaxLength())
	ErrAmountTooLong       = fmt.Errorf("amount must be at most %d characters", validation.SuperSmall.MaxLength())
	ErrFeedingNotesTooLong = fmt.Errorf("notes must be at most %d characters", validation.ExtraLarge.MaxLength())
)

// FeedingRepository is the persistence interface for feedings.
type FeedingRepository interface {
	BulkCreate(ctx context.Context, feedings []*model.Feeding) error
	Create(ctx context.Context, f *model.Feeding) error
	Delete(ctx context.Context, feedingID int64) error
	GetByID(ctx context.Context, feedingID, hiveID int64) (*model.Feeding, error)
	CountByHiveID(ctx context.Context, hiveID int64) (int64, error)
	ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Feeding, error)
	Update(ctx context.Context, f *model.Feeding) error
	DistinctFeedTypes(ctx context.Context, userID int64) ([]string, error)
	DistinctAmounts(ctx context.Context, userID int64) ([]string, error)
}

// FeedingService handles business logic for feeding records.
type FeedingService struct {
	apiaries ApiaryMembershipReader
	hives    InspectionHiveReader
	allHives BulkHiveReader
	feedings FeedingRepository
}

// NewFeedingService creates a FeedingService with the given dependencies.
func NewFeedingService(apiaries ApiaryMembershipReader, hives InspectionHiveReader, allHives BulkHiveReader, feedings FeedingRepository) *FeedingService {
	return &FeedingService{apiaries: apiaries, hives: hives, allHives: allHives, feedings: feedings}
}

// FeedingParams holds the mutable fields for create and update operations.
type FeedingParams struct {
	FedAt    time.Time
	FeedType string
	Amount   string
	Notes    string
}

func validateFeedingParams(p FeedingParams) error {
	if p.FedAt.IsZero() {
		return ErrFedAtRequired
	}
	if p.FeedType == "" {
		return ErrFeedTypeRequired
	}
	if validation.TooLong(p.FeedType, validation.Small) {
		return ErrFeedTypeTooLong
	}
	if validation.TooLong(p.Amount, validation.SuperSmall) {
		return ErrAmountTooLong
	}
	if validation.TooLong(p.Notes, validation.ExtraLarge) {
		return ErrFeedingNotesTooLong
	}
	return nil
}

func (s *FeedingService) checkAccess(ctx context.Context, apiaryID, userID, hiveID int64) error {
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

// Create validates params, checks membership, and inserts a new feeding.
func (s *FeedingService) Create(ctx context.Context, userID, apiaryID, hiveID int64, params FeedingParams) (*model.Feeding, error) {
	if err := validateFeedingParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	f := &model.Feeding{
		HiveID:   hiveID,
		FedBy:    userID,
		FedAt:    params.FedAt,
		FeedType: params.FeedType,
		Amount:   params.Amount,
		Notes:    params.Notes,
	}
	if err := s.feedings.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("create feeding: %w", err)
	}
	return f, nil
}

// Get returns a single feeding, verifying apiary membership and hive ownership.
func (s *FeedingService) Get(ctx context.Context, userID, apiaryID, hiveID, feedingID int64) (*model.Feeding, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	f, err := s.feedings.GetByID(ctx, feedingID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeedingNotFound
		}
		return nil, fmt.Errorf("get feeding: %w", err)
	}
	return f, nil
}

// List returns a paginated slice of feedings and the total count for a hive.
func (s *FeedingService) List(ctx context.Context, userID, apiaryID, hiveID int64, limit, offset int) ([]*model.Feeding, int64, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, 0, err
	}
	total, err := s.feedings.CountByHiveID(ctx, hiveID)
	if err != nil {
		return nil, 0, fmt.Errorf("count feedings: %w", err)
	}
	feedings, err := s.feedings.ListByHiveID(ctx, hiveID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list feedings: %w", err)
	}
	return feedings, total, nil
}

// Update validates params and overwrites all mutable fields of an existing feeding.
func (s *FeedingService) Update(ctx context.Context, userID, apiaryID, hiveID, feedingID int64, params FeedingParams) (*model.Feeding, error) {
	if err := validateFeedingParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	f, err := s.feedings.GetByID(ctx, feedingID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeedingNotFound
		}
		return nil, fmt.Errorf("get feeding: %w", err)
	}
	f.FedAt = params.FedAt
	f.FeedType = params.FeedType
	f.Amount = params.Amount
	f.Notes = params.Notes
	if err := s.feedings.Update(ctx, f); err != nil {
		return nil, fmt.Errorf("update feeding: %w", err)
	}
	return f, nil
}

// BulkFeed creates one feeding record for each hive in hiveIDs (or every hive in the
// apiary when hiveIDs is empty) within a single transaction. Any id in hiveIDs that does
// not belong to the apiary is silently ignored. It returns the number of feedings inserted.
func (s *FeedingService) BulkFeed(ctx context.Context, userID, apiaryID int64, hiveIDs []int64, params FeedingParams) (int, error) {
	if err := validateFeedingParams(params); err != nil {
		return 0, err
	}
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrApiaryNotFound
		}
		return 0, fmt.Errorf("get apiary: %w", err)
	}
	allHives, err := s.allHives.ListByApiaryID(ctx, apiaryID)
	if err != nil {
		return 0, fmt.Errorf("list hives: %w", err)
	}
	hives := allHives
	if len(hiveIDs) > 0 {
		wanted := make(map[int64]bool, len(hiveIDs))
		for _, id := range hiveIDs {
			wanted[id] = true
		}
		hives = make([]*model.Hive, 0, len(allHives))
		for _, h := range allHives {
			if wanted[h.ID] {
				hives = append(hives, h)
			}
		}
	}
	feedings := make([]*model.Feeding, len(hives))
	for i, h := range hives {
		feedings[i] = &model.Feeding{
			HiveID:   h.ID,
			FedBy:    userID,
			FedAt:    params.FedAt,
			FeedType: params.FeedType,
			Amount:   params.Amount,
			Notes:    params.Notes,
		}
	}
	if err := s.feedings.BulkCreate(ctx, feedings); err != nil {
		return 0, fmt.Errorf("bulk create feedings: %w", err)
	}
	return len(feedings), nil
}

// FeedTypeSuggestions returns the feed types userID has previously used, most
// recently used first.
func (s *FeedingService) FeedTypeSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return s.feedings.DistinctFeedTypes(ctx, userID)
}

// AmountSuggestions returns the amounts userID has previously used, most
// recently used first.
func (s *FeedingService) AmountSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return s.feedings.DistinctAmounts(ctx, userID)
}

// Delete removes a feeding after verifying membership.
func (s *FeedingService) Delete(ctx context.Context, userID, apiaryID, hiveID, feedingID int64) error {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return err
	}
	if _, err := s.feedings.GetByID(ctx, feedingID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeedingNotFound
		}
		return fmt.Errorf("get feeding: %w", err)
	}
	if err := s.feedings.Delete(ctx, feedingID); err != nil {
		return fmt.Errorf("delete feeding: %w", err)
	}
	return nil
}
