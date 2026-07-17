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
	ErrTreatmentNotFound     = errors.New("treatment not found")
	ErrTreatedAtRequired     = errors.New("treated_at is required")
	ErrMedicineNameRequired  = errors.New("medicine_name is required")
	ErrMedicineNameTooLong   = fmt.Errorf("medicine_name must be at most %d characters", validation.Small.MaxLength())
	ErrDoseTooLong           = fmt.Errorf("dose must be at most %d characters", validation.SuperSmall.MaxLength())
	ErrTreatmentNotesTooLong = fmt.Errorf("notes must be at most %d characters", validation.ExtraLarge.MaxLength())
)

// TreatmentRepository is the persistence interface for treatments.
type TreatmentRepository interface {
	BulkCreate(ctx context.Context, treatments []*model.Treatment) error
	Create(ctx context.Context, t *model.Treatment) error
	Delete(ctx context.Context, treatmentID int64) error
	GetByID(ctx context.Context, treatmentID, hiveID int64) (*model.Treatment, error)
	CountByHiveID(ctx context.Context, hiveID int64) (int64, error)
	ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Treatment, error)
	Update(ctx context.Context, t *model.Treatment) error
	DistinctMedicineNames(ctx context.Context, userID int64) ([]string, error)
	DistinctDoses(ctx context.Context, userID int64) ([]string, error)
}

// BulkHiveReader lists all hives for an apiary.
type BulkHiveReader interface {
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
}

// TreatmentService handles business logic for treatment records.
type TreatmentService struct {
	apiaries   ApiaryMembershipReader
	hives      InspectionHiveReader
	allHives   BulkHiveReader
	treatments TreatmentRepository
}

// NewTreatmentService creates a TreatmentService with the given dependencies.
func NewTreatmentService(apiaries ApiaryMembershipReader, hives InspectionHiveReader, allHives BulkHiveReader, treatments TreatmentRepository) *TreatmentService {
	return &TreatmentService{apiaries: apiaries, hives: hives, allHives: allHives, treatments: treatments}
}

// TreatmentParams holds the mutable fields for create and update operations.
type TreatmentParams struct {
	TreatedAt    time.Time
	MedicineName string
	Dose         string
	Notes        string
}

func validateTreatmentParams(p TreatmentParams) error {
	if p.TreatedAt.IsZero() {
		return ErrTreatedAtRequired
	}
	if p.MedicineName == "" {
		return ErrMedicineNameRequired
	}
	if validation.TooLong(p.MedicineName, validation.Small) {
		return ErrMedicineNameTooLong
	}
	if validation.TooLong(p.Dose, validation.SuperSmall) {
		return ErrDoseTooLong
	}
	if validation.TooLong(p.Notes, validation.ExtraLarge) {
		return ErrTreatmentNotesTooLong
	}
	return nil
}

func (s *TreatmentService) checkAccess(ctx context.Context, apiaryID, userID, hiveID int64) error {
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

// Create validates params, checks membership, and inserts a new treatment.
func (s *TreatmentService) Create(ctx context.Context, userID, apiaryID, hiveID int64, params TreatmentParams) (*model.Treatment, error) {
	if err := validateTreatmentParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	dose := params.Dose
	if dose == "" {
		dose = "1"
	}
	t := &model.Treatment{
		HiveID:       hiveID,
		TreatedBy:    userID,
		TreatedAt:    params.TreatedAt,
		MedicineName: params.MedicineName,
		Dose:         dose,
		Notes:        params.Notes,
	}
	if err := s.treatments.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create treatment: %w", err)
	}
	return t, nil
}

// Get returns a single treatment, verifying apiary membership and hive ownership.
func (s *TreatmentService) Get(ctx context.Context, userID, apiaryID, hiveID, treatmentID int64) (*model.Treatment, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	t, err := s.treatments.GetByID(ctx, treatmentID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTreatmentNotFound
		}
		return nil, fmt.Errorf("get treatment: %w", err)
	}
	return t, nil
}

// List returns a paginated slice of treatments and the total count for a hive.
func (s *TreatmentService) List(ctx context.Context, userID, apiaryID, hiveID int64, limit, offset int) ([]*model.Treatment, int64, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, 0, err
	}
	total, err := s.treatments.CountByHiveID(ctx, hiveID)
	if err != nil {
		return nil, 0, fmt.Errorf("count treatments: %w", err)
	}
	treatments, err := s.treatments.ListByHiveID(ctx, hiveID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list treatments: %w", err)
	}
	return treatments, total, nil
}

// Update validates params and overwrites all mutable fields of an existing treatment.
func (s *TreatmentService) Update(ctx context.Context, userID, apiaryID, hiveID, treatmentID int64, params TreatmentParams) (*model.Treatment, error) {
	if err := validateTreatmentParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	t, err := s.treatments.GetByID(ctx, treatmentID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTreatmentNotFound
		}
		return nil, fmt.Errorf("get treatment: %w", err)
	}
	dose := params.Dose
	if dose == "" {
		dose = "1"
	}
	t.TreatedAt = params.TreatedAt
	t.MedicineName = params.MedicineName
	t.Dose = dose
	t.Notes = params.Notes
	if err := s.treatments.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update treatment: %w", err)
	}
	return t, nil
}

// BulkTreat creates one treatment record for each hive in hiveIDs (or every hive in the
// apiary when hiveIDs is empty) within a single transaction. Any id in hiveIDs that does
// not belong to the apiary is silently ignored. It returns the number of treatments inserted.
func (s *TreatmentService) BulkTreat(ctx context.Context, userID, apiaryID int64, hiveIDs []int64, params TreatmentParams) (int, error) {
	if err := validateTreatmentParams(params); err != nil {
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
	dose := params.Dose
	if dose == "" {
		dose = "1"
	}
	treatments := make([]*model.Treatment, len(hives))
	for i, h := range hives {
		treatments[i] = &model.Treatment{
			HiveID:       h.ID,
			TreatedBy:    userID,
			TreatedAt:    params.TreatedAt,
			MedicineName: params.MedicineName,
			Dose:         dose,
			Notes:        params.Notes,
		}
	}
	if err := s.treatments.BulkCreate(ctx, treatments); err != nil {
		return 0, fmt.Errorf("bulk create treatments: %w", err)
	}
	return len(treatments), nil
}

// MedicineSuggestions returns the medicine names userID has previously used,
// most recently used first.
func (s *TreatmentService) MedicineSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return s.treatments.DistinctMedicineNames(ctx, userID)
}

// DoseSuggestions returns the doses userID has previously used, most recently
// used first.
func (s *TreatmentService) DoseSuggestions(ctx context.Context, userID int64) ([]string, error) {
	return s.treatments.DistinctDoses(ctx, userID)
}

// Delete removes a treatment after verifying membership.
func (s *TreatmentService) Delete(ctx context.Context, userID, apiaryID, hiveID, treatmentID int64) error {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return err
	}
	if _, err := s.treatments.GetByID(ctx, treatmentID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTreatmentNotFound
		}
		return fmt.Errorf("get treatment: %w", err)
	}
	if err := s.treatments.Delete(ctx, treatmentID); err != nil {
		return fmt.Errorf("delete treatment: %w", err)
	}
	return nil
}
