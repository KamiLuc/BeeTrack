package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

type mockReportInspectionReader struct {
	inspections []*model.Inspection
}

func (m *mockReportInspectionReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Inspection, error) {
	return m.inspections, nil
}

type mockReportTreatmentReader struct {
	treatments []*model.Treatment
}

func (m *mockReportTreatmentReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Treatment, error) {
	return m.treatments, nil
}

type mockReportFeedingReader struct {
	feedings []*model.Feeding
}

func (m *mockReportFeedingReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Feeding, error) {
	return m.feedings, nil
}

type mockReportHarvestReader struct {
	harvests []*model.Harvest
}

func (m *mockReportHarvestReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Harvest, error) {
	return m.harvests, nil
}

func newReportSvc(hives []*model.Hive, insp []*model.Inspection, tr []*model.Treatment, feed []*model.Feeding, harv []*model.Harvest) *ReportService {
	return NewReportService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1, Name: "Pasieka Testowa"}, role: "member"},
		&mockBulkHiveReader{hives: hives},
		&mockReportInspectionReader{inspections: insp},
		&mockReportTreatmentReader{treatments: tr},
		&mockReportFeedingReader{feedings: feed},
		&mockReportHarvestReader{harvests: harv},
	)
}

func TestReportGenerateRequiresCategories(t *testing.T) {
	svc := newReportSvc(nil, nil, nil, nil, nil)
	_, err := svc.Generate(context.Background(), 1, 1, ReportFilter{
		HiveIDs:    []int64{10},
		Categories: nil,
		From:       time.Now(),
		To:         time.Now(),
	})
	if !errors.Is(err, ErrReportCategoriesRequired) {
		t.Fatalf("expected ErrReportCategoriesRequired, got %v", err)
	}
}

func TestReportGenerateInvalidDateRange(t *testing.T) {
	svc := newReportSvc(nil, nil, nil, nil, nil)
	now := time.Now()
	_, err := svc.Generate(context.Background(), 1, 1, ReportFilter{
		HiveIDs:    []int64{10},
		Categories: []ReportCategory{ReportCategoryInspections},
		From:       now,
		To:         now.AddDate(0, 0, -1),
	})
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("expected ErrInvalidDateRange, got %v", err)
	}
}

func TestReportGenerateApiaryNotFound(t *testing.T) {
	svc := newReportSvc(nil, nil, nil, nil, nil)
	_, err := svc.Generate(context.Background(), 1, 999, ReportFilter{
		HiveIDs:    []int64{10},
		Categories: []ReportCategory{ReportCategoryInspections},
		From:       time.Now(),
		To:         time.Now(),
	})
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Fatalf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestReportGenerateDropsHiveIDsNotInApiary(t *testing.T) {
	hives := []*model.Hive{{ID: 10, ApiaryID: 1, Name: "Ul 1"}}
	svc := newReportSvc(hives, nil, nil, nil, nil)
	_, err := svc.Generate(context.Background(), 1, 1, ReportFilter{
		HiveIDs:    []int64{999},
		Categories: []ReportCategory{ReportCategoryInspections},
		From:       time.Now(),
		To:         time.Now(),
	})
	if !errors.Is(err, ErrHiveIDsRequired) {
		t.Fatalf("expected ErrHiveIDsRequired when no requested hive belongs to the apiary, got %v", err)
	}
}

func TestReportGenerateGroupsAndSortsEntriesPerHive(t *testing.T) {
	hives := []*model.Hive{
		{ID: 10, ApiaryID: 1, Name: "Ul 1"},
		{ID: 11, ApiaryID: 1, Name: "Ul 2"},
	}
	now := time.Now()
	insp := []*model.Inspection{
		{ID: 1, HiveID: 10, InspectedAt: now.AddDate(0, 0, -5)},
		{ID: 2, HiveID: 10, InspectedAt: now},
		{ID: 3, HiveID: 11, InspectedAt: now.AddDate(0, 0, -1)},
	}
	tr := []*model.Treatment{
		{ID: 1, HiveID: 10, TreatedAt: now.AddDate(0, 0, -2)},
	}

	svc := newReportSvc(hives, insp, tr, nil, nil)
	report, err := svc.Generate(context.Background(), 1, 1, ReportFilter{
		HiveIDs:    []int64{10, 11},
		Categories: []ReportCategory{ReportCategoryInspections, ReportCategoryTreatments},
		From:       now.AddDate(0, 0, -10),
		To:         now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Hives) != 2 {
		t.Fatalf("expected 2 hives, got %d", len(report.Hives))
	}

	ul1 := report.Hives[0]
	if ul1.Hive.ID != 10 {
		t.Fatalf("expected first hive to be id 10, got %d", ul1.Hive.ID)
	}
	if len(ul1.Entries) != 3 {
		t.Fatalf("expected 3 entries for hive 10 (2 inspections + 1 treatment), got %d", len(ul1.Entries))
	}
	// Newest first.
	if ul1.Entries[0].Inspection == nil || ul1.Entries[0].Inspection.ID != 2 {
		t.Fatalf("expected newest inspection (id 2) first, got %+v", ul1.Entries[0])
	}

	ul2 := report.Hives[1]
	if len(ul2.Entries) != 1 || ul2.Entries[0].Inspection == nil || ul2.Entries[0].Inspection.ID != 3 {
		t.Fatalf("expected hive 11 to have only inspection id 3, got %+v", ul2.Entries)
	}
}

func TestParseReportCategory(t *testing.T) {
	valid := []string{"inspections", "feedings", "treatments", "harvests"}
	for _, v := range valid {
		if _, err := ParseReportCategory(v); err != nil {
			t.Errorf("expected %q to be valid, got error %v", v, err)
		}
	}
	if _, err := ParseReportCategory("bogus"); !errors.Is(err, ErrInvalidReportCategory) {
		t.Errorf("expected ErrInvalidReportCategory for bogus input, got %v", err)
	}
}
