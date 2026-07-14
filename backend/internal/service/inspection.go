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

// maxInspectionFrameCount is the largest frame count (frames_brood/feed/pollen, queen_cells_count) an inspection may record.
const maxInspectionFrameCount = 99

// maxFramesAddedDelta is the largest magnitude allowed for a frames_added_* signed delta.
const maxFramesAddedDelta = 99

var (
	ErrDiseaseNotFound              = errors.New("disease not found")
	ErrInspectionFrameCountInvalid  = fmt.Errorf("frame counts must be between 0 and %d", maxInspectionFrameCount)
	ErrInspectionFramesAddedInvalid = fmt.Errorf("frames added/removed must be between -%d and %d", maxFramesAddedDelta, maxFramesAddedDelta)
	ErrInspectionNotFound           = errors.New("inspection not found")
	ErrInspectedAtRequired          = errors.New("inspected_at is required")
	ErrInspectionNotesTooLong       = fmt.Errorf("notes must be at most %d characters", validation.ExtraLarge.MaxLength())
	ErrInvalidAggressiveness        = errors.New("invalid aggressiveness value")
	ErrInvalidBroodPattern          = errors.New("invalid brood_pattern value")
	ErrInvalidDisease               = errors.New("invalid disease value")
	ErrInvalidQueenStatus           = errors.New("invalid queen_status value")
)

var validDiseases = map[string]bool{
	"american_foulbrood": true, "chalkbrood": true, "dwv": true,
	"european_foulbrood": true, "laying_workers": true, "nosema": true, "varroa": true,
}

var validAggressiveness = map[string]bool{
	"aggressive": true, "calm": true, "mild": true, "very_aggressive": true,
}

var validBroodPattern = map[string]bool{
	"excellent": true, "good": true, "none": true, "poor": true,
	"few": true, "medium": true, "many": true,
}

var validQueenStatus = map[string]bool{
	"not_seen": true, "seen": true,
}

// InspectionHiveReader is the subset of HiveRepository needed by InspectionService.
type InspectionHiveReader interface {
	GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error)
}

// InspectionRepository is the persistence interface for inspections and their diseases.
type InspectionRepository interface {
	Create(ctx context.Context, insp *model.Inspection) error
	CreateDisease(ctx context.Context, d *model.InspectionDisease) error
	Delete(ctx context.Context, inspectionID int64) error
	DeleteDisease(ctx context.Context, diseaseID, inspectionID int64) error
	GetByID(ctx context.Context, inspectionID, hiveID int64) (*model.Inspection, error)
	GetDiseaseByID(ctx context.Context, diseaseID, inspectionID int64) (*model.InspectionDisease, error)
	CountByHiveID(ctx context.Context, hiveID int64) (int64, error)
	ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Inspection, error)
	ListDiseasesByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionDisease, error)
	LastInspectionDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error)
	ListDiseasesByInspectionIDs(ctx context.Context, ids []int64) ([]*model.InspectionDisease, error)
	Update(ctx context.Context, insp *model.Inspection) error
}

// InspectionService handles business logic for inspection records.
type InspectionService struct {
	apiaries    ApiaryMembershipReader
	hives       InspectionHiveReader
	inspections InspectionRepository
}

// NewInspectionService creates an InspectionService with the given dependencies.
func NewInspectionService(apiaries ApiaryMembershipReader, hives InspectionHiveReader, inspections InspectionRepository) *InspectionService {
	return &InspectionService{apiaries: apiaries, hives: hives, inspections: inspections}
}

// InspectionParams holds the mutable fields for create and update operations.
type InspectionParams struct {
	InspectedAt           time.Time
	QueenStatus           string
	BroodPattern          string
	FramesBrood           *int
	FramesFeed            *int
	FramesPollen          *int
	QueenCellsCount       *int
	Aggressiveness        string
	FramesAddedFoundation *int
	FramesAddedDrawn      *int
	FramesAddedBrood      *int
	FramesAddedFeed       *int
	QueenAdded            bool
	Notes                 string
}

func validateInspectionParams(p InspectionParams) error {
	if p.InspectedAt.IsZero() {
		return ErrInspectedAtRequired
	}
	if p.Aggressiveness != "" && !validAggressiveness[p.Aggressiveness] {
		return ErrInvalidAggressiveness
	}
	if p.BroodPattern != "" && !validBroodPattern[p.BroodPattern] {
		return ErrInvalidBroodPattern
	}
	if p.QueenStatus != "" && !validQueenStatus[p.QueenStatus] {
		return ErrInvalidQueenStatus
	}
	if !validFrameCount(p.FramesBrood) || !validFrameCount(p.FramesFeed) ||
		!validFrameCount(p.FramesPollen) || !validFrameCount(p.QueenCellsCount) {
		return ErrInspectionFrameCountInvalid
	}
	if !validFramesAddedDelta(p.FramesAddedFoundation) || !validFramesAddedDelta(p.FramesAddedDrawn) ||
		!validFramesAddedDelta(p.FramesAddedBrood) || !validFramesAddedDelta(p.FramesAddedFeed) {
		return ErrInspectionFramesAddedInvalid
	}
	if validation.TooLong(p.Notes, validation.ExtraLarge) {
		return ErrInspectionNotesTooLong
	}
	return nil
}

// validFrameCount reports whether v is unset or within [0, maxInspectionFrameCount].
func validFrameCount(v *int) bool {
	return v == nil || (*v >= 0 && *v <= maxInspectionFrameCount)
}

// validFramesAddedDelta reports whether v is unset or within [-maxFramesAddedDelta, maxFramesAddedDelta].
func validFramesAddedDelta(v *int) bool {
	return v == nil || (*v >= -maxFramesAddedDelta && *v <= maxFramesAddedDelta)
}

func (s *InspectionService) checkAccess(ctx context.Context, apiaryID, userID, hiveID int64) error {
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

// Create validates params, checks membership, and inserts a new inspection.
func (s *InspectionService) Create(ctx context.Context, userID, apiaryID, hiveID int64, params InspectionParams) (*model.Inspection, error) {
	if err := validateInspectionParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	insp := &model.Inspection{
		HiveID:                hiveID,
		InspectedBy:           userID,
		InspectedAt:           params.InspectedAt,
		QueenStatus:           params.QueenStatus,
		BroodPattern:          params.BroodPattern,
		FramesBrood:           params.FramesBrood,
		FramesFeed:            params.FramesFeed,
		FramesPollen:          params.FramesPollen,
		QueenCellsCount:       params.QueenCellsCount,
		Aggressiveness:        params.Aggressiveness,
		FramesAddedFoundation: params.FramesAddedFoundation,
		FramesAddedDrawn:      params.FramesAddedDrawn,
		FramesAddedBrood:      params.FramesAddedBrood,
		FramesAddedFeed:       params.FramesAddedFeed,
		QueenAdded:            params.QueenAdded,
		Notes:                 params.Notes,
	}
	if err := s.inspections.Create(ctx, insp); err != nil {
		return nil, fmt.Errorf("create inspection: %w", err)
	}
	return insp, nil
}

// Get returns a single inspection, verifying apiary membership and hive ownership.
func (s *InspectionService) Get(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64) (*model.Inspection, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	insp, err := s.inspections.GetByID(ctx, inspectionID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("get inspection: %w", err)
	}
	return insp, nil
}

// List returns paginated inspections for a hive ordered by inspected_at descending.
// List returns a paginated slice of inspections and the total count for a hive.
func (s *InspectionService) List(ctx context.Context, userID, apiaryID, hiveID int64, limit, offset int) ([]*model.Inspection, int64, error) {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, 0, err
	}
	total, err := s.inspections.CountByHiveID(ctx, hiveID)
	if err != nil {
		return nil, 0, fmt.Errorf("count inspections: %w", err)
	}
	inspections, err := s.inspections.ListByHiveID(ctx, hiveID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list inspections: %w", err)
	}
	return inspections, total, nil
}

// Update validates params and overwrites all mutable fields of an existing inspection.
func (s *InspectionService) Update(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64, params InspectionParams) (*model.Inspection, error) {
	if err := validateInspectionParams(params); err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	insp, err := s.inspections.GetByID(ctx, inspectionID, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("get inspection: %w", err)
	}
	insp.InspectedAt = params.InspectedAt
	insp.QueenStatus = params.QueenStatus
	insp.BroodPattern = params.BroodPattern
	insp.FramesBrood = params.FramesBrood
	insp.FramesFeed = params.FramesFeed
	insp.FramesPollen = params.FramesPollen
	insp.QueenCellsCount = params.QueenCellsCount
	insp.Aggressiveness = params.Aggressiveness
	insp.FramesAddedFoundation = params.FramesAddedFoundation
	insp.FramesAddedDrawn = params.FramesAddedDrawn
	insp.FramesAddedBrood = params.FramesAddedBrood
	insp.FramesAddedFeed = params.FramesAddedFeed
	insp.QueenAdded = params.QueenAdded
	insp.Notes = params.Notes
	if err := s.inspections.Update(ctx, insp); err != nil {
		return nil, fmt.Errorf("update inspection: %w", err)
	}
	return insp, nil
}

// DiseasesByInspection returns diseases for an inspection. Caller must have already verified access.
func (s *InspectionService) DiseasesByInspection(ctx context.Context, inspectionID int64) ([]*model.InspectionDisease, error) {
	diseases, err := s.inspections.ListDiseasesByInspectionID(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("list diseases: %w", err)
	}
	return diseases, nil
}

// DiseasesForInspections returns diseases grouped by inspection ID. Caller must have already verified access.
func (s *InspectionService) DiseasesForInspections(ctx context.Context, ids []int64) (map[int64][]*model.InspectionDisease, error) {
	rows, err := s.inspections.ListDiseasesByInspectionIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list diseases: %w", err)
	}
	grouped := make(map[int64][]*model.InspectionDisease, len(ids))
	for _, id := range ids {
		grouped[id] = []*model.InspectionDisease{}
	}
	for _, d := range rows {
		grouped[d.InspectionID] = append(grouped[d.InspectionID], d)
	}
	return grouped, nil
}

// AddDisease validates the disease name, verifies access, and attaches a disease to an inspection.
func (s *InspectionService) AddDisease(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64, disease, notes string) (*model.InspectionDisease, error) {
	if !validDiseases[disease] {
		return nil, ErrInvalidDisease
	}
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return nil, err
	}
	if _, err := s.inspections.GetByID(ctx, inspectionID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("get inspection: %w", err)
	}
	d := &model.InspectionDisease{
		InspectionID: inspectionID,
		Disease:      disease,
		Notes:        notes,
	}
	if err := s.inspections.CreateDisease(ctx, d); err != nil {
		return nil, fmt.Errorf("create disease: %w", err)
	}
	return d, nil
}

// RemoveDisease verifies access and deletes a disease record from an inspection.
func (s *InspectionService) RemoveDisease(ctx context.Context, userID, apiaryID, hiveID, inspectionID, diseaseID int64) error {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return err
	}
	if _, err := s.inspections.GetByID(ctx, inspectionID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInspectionNotFound
		}
		return fmt.Errorf("get inspection: %w", err)
	}
	if _, err := s.inspections.GetDiseaseByID(ctx, diseaseID, inspectionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDiseaseNotFound
		}
		return fmt.Errorf("get disease: %w", err)
	}
	if err := s.inspections.DeleteDisease(ctx, diseaseID, inspectionID); err != nil {
		return fmt.Errorf("delete disease: %w", err)
	}
	return nil
}

// LastInspectionDates returns the most recent inspected_at per hive for the given IDs.
func (s *InspectionService) LastInspectionDates(ctx context.Context, hiveIDs []int64) (map[int64]*time.Time, error) {
	if len(hiveIDs) == 0 {
		return map[int64]*time.Time{}, nil
	}
	dates, err := s.inspections.LastInspectionDatesByHiveIDs(ctx, hiveIDs)
	if err != nil {
		return nil, fmt.Errorf("last inspection dates: %w", err)
	}
	return dates, nil
}

// Delete removes an inspection after verifying membership.
func (s *InspectionService) Delete(ctx context.Context, userID, apiaryID, hiveID, inspectionID int64) error {
	if err := s.checkAccess(ctx, apiaryID, userID, hiveID); err != nil {
		return err
	}
	if _, err := s.inspections.GetByID(ctx, inspectionID, hiveID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInspectionNotFound
		}
		return fmt.Errorf("get inspection: %w", err)
	}
	if err := s.inspections.Delete(ctx, inspectionID); err != nil {
		return fmt.Errorf("delete inspection: %w", err)
	}
	return nil
}
