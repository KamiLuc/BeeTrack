package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	// ErrHiveIDsRequired is returned when no requested hive ID belongs to the apiary.
	ErrHiveIDsRequired = errors.New("at least one hive id is required")
	// ErrReportCategoriesRequired is returned when no record category was selected.
	ErrReportCategoriesRequired = errors.New("at least one category is required")
	// ErrInvalidDateRange is returned when from is after to.
	ErrInvalidDateRange = errors.New("from must be on or before to")
	// ErrInvalidReportCategory is returned for an unrecognized category value.
	ErrInvalidReportCategory = errors.New("invalid report category")
)

// ReportCategory selects which kind of record a report includes.
type ReportCategory string

const (
	ReportCategoryInspections ReportCategory = "inspections"
	ReportCategoryFeedings    ReportCategory = "feedings"
	ReportCategoryTreatments  ReportCategory = "treatments"
	ReportCategoryHarvests    ReportCategory = "harvests"
)

// ReportCategoryOrder is the fixed display order for report sections.
var ReportCategoryOrder = []ReportCategory{
	ReportCategoryInspections,
	ReportCategoryFeedings,
	ReportCategoryTreatments,
	ReportCategoryHarvests,
}

// ParseReportCategory validates and converts a raw string into a ReportCategory.
func ParseReportCategory(raw string) (ReportCategory, error) {
	switch ReportCategory(raw) {
	case ReportCategoryInspections, ReportCategoryFeedings, ReportCategoryTreatments, ReportCategoryHarvests:
		return ReportCategory(raw), nil
	default:
		return "", ErrInvalidReportCategory
	}
}

// ReportFilter describes what a report should cover.
type ReportFilter struct {
	HiveIDs    []int64
	Categories []ReportCategory
	From       time.Time
	To         time.Time
}

// ReportEntry is a single record in a report, typed by Category. Only the
// field matching Category is populated.
type ReportEntry struct {
	Category   ReportCategory
	Date       time.Time
	Inspection *model.Inspection
	Treatment  *model.Treatment
	Feeding    *model.Feeding
	Harvest    *model.Harvest
}

// ReportHive is one hive's matching entries, newest first.
type ReportHive struct {
	Hive    *model.Hive
	Entries []ReportEntry
}

// Report is the full generated result for an apiary.
type Report struct {
	Apiary *model.Apiary
	From   time.Time
	To     time.Time
	Hives  []ReportHive
}

// ReportInspectionReader lists inspections for a set of hives within a date range.
type ReportInspectionReader interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Inspection, error)
}

// ReportTreatmentReader lists treatments for a set of hives within a date range.
type ReportTreatmentReader interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Treatment, error)
}

// ReportFeedingReader lists feedings for a set of hives within a date range.
type ReportFeedingReader interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Feeding, error)
}

// ReportHarvestReader lists harvests for a set of hives within a date range.
type ReportHarvestReader interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Harvest, error)
}

// ReportService generates apiary activity reports for PDF export.
type ReportService struct {
	apiaries    ApiaryMembershipReader
	hives       BulkHiveReader
	inspections ReportInspectionReader
	treatments  ReportTreatmentReader
	feedings    ReportFeedingReader
	harvests    ReportHarvestReader
}

// NewReportService creates a ReportService with the given dependencies.
func NewReportService(
	apiaries ApiaryMembershipReader,
	hives BulkHiveReader,
	inspections ReportInspectionReader,
	treatments ReportTreatmentReader,
	feedings ReportFeedingReader,
	harvests ReportHarvestReader,
) *ReportService {
	return &ReportService{
		apiaries:    apiaries,
		hives:       hives,
		inspections: inspections,
		treatments:  treatments,
		feedings:    feedings,
		harvests:    harvests,
	}
}

// Generate validates filter, verifies apiary membership, and builds a Report
// covering only the requested hives (silently dropping any hive ID that
// doesn't belong to the apiary) and categories.
func (s *ReportService) Generate(ctx context.Context, userID, apiaryID int64, filter ReportFilter) (*Report, error) {
	if len(filter.Categories) == 0 {
		return nil, ErrReportCategoriesRequired
	}
	if filter.From.After(filter.To) {
		return nil, ErrInvalidDateRange
	}

	apiary, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}

	apiaryHives, err := s.hives.ListByApiaryID(ctx, apiaryID)
	if err != nil {
		return nil, fmt.Errorf("list hives: %w", err)
	}
	byID := make(map[int64]*model.Hive, len(apiaryHives))
	for _, h := range apiaryHives {
		byID[h.ID] = h
	}

	var validHives []*model.Hive
	var validIDs []int64
	for _, id := range filter.HiveIDs {
		if h, ok := byID[id]; ok {
			validHives = append(validHives, h)
			validIDs = append(validIDs, id)
		}
	}
	if len(validHives) == 0 {
		return nil, ErrHiveIDsRequired
	}

	// to is treated as an inclusive calendar day, so query with an exclusive
	// upper bound of the day after.
	upperBound := time.Date(filter.To.Year(), filter.To.Month(), filter.To.Day(), 0, 0, 0, 0, filter.To.Location()).AddDate(0, 0, 1)

	selected := make(map[ReportCategory]bool, len(filter.Categories))
	for _, c := range filter.Categories {
		selected[c] = true
	}

	byHive := make(map[int64][]ReportEntry, len(validIDs))

	if selected[ReportCategoryInspections] {
		rows, err := s.inspections.ListByHiveIDsAndRange(ctx, validIDs, filter.From, upperBound)
		if err != nil {
			return nil, fmt.Errorf("list inspections: %w", err)
		}
		for _, insp := range rows {
			byHive[insp.HiveID] = append(byHive[insp.HiveID], ReportEntry{
				Category:   ReportCategoryInspections,
				Date:       insp.InspectedAt,
				Inspection: insp,
			})
		}
	}
	if selected[ReportCategoryTreatments] {
		rows, err := s.treatments.ListByHiveIDsAndRange(ctx, validIDs, filter.From, upperBound)
		if err != nil {
			return nil, fmt.Errorf("list treatments: %w", err)
		}
		for _, t := range rows {
			byHive[t.HiveID] = append(byHive[t.HiveID], ReportEntry{
				Category:  ReportCategoryTreatments,
				Date:      t.TreatedAt,
				Treatment: t,
			})
		}
	}
	if selected[ReportCategoryFeedings] {
		rows, err := s.feedings.ListByHiveIDsAndRange(ctx, validIDs, filter.From, upperBound)
		if err != nil {
			return nil, fmt.Errorf("list feedings: %w", err)
		}
		for _, f := range rows {
			byHive[f.HiveID] = append(byHive[f.HiveID], ReportEntry{
				Category: ReportCategoryFeedings,
				Date:     f.FedAt,
				Feeding:  f,
			})
		}
	}
	if selected[ReportCategoryHarvests] {
		rows, err := s.harvests.ListByHiveIDsAndRange(ctx, validIDs, filter.From, upperBound)
		if err != nil {
			return nil, fmt.Errorf("list harvests: %w", err)
		}
		for _, h := range rows {
			byHive[h.HiveID] = append(byHive[h.HiveID], ReportEntry{
				Category: ReportCategoryHarvests,
				Date:     h.HarvestedAt,
				Harvest:  h,
			})
		}
	}

	report := &Report{Apiary: apiary, From: filter.From, To: filter.To}
	for _, h := range validHives {
		entries := byHive[h.ID]
		sort.Slice(entries, func(i, j int) bool { return entries[i].Date.After(entries[j].Date) })
		report.Hives = append(report.Hives, ReportHive{Hive: h, Entries: entries})
	}

	return report, nil
}
