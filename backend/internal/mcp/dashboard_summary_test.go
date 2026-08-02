package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

func TestGetDashboardSummaryAggregatesAcrossApiaries(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{
		{Apiary: &model.Apiary{ID: 1}},
		{Apiary: &model.Apiary{ID: 2}},
	}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true},
			{ID: 11, ApiaryID: 1, Name: "Hive B", Active: true, NeedsFood: true},
			{ID: 12, ApiaryID: 1, Name: "Hive D", Active: true, QueenNeedsReplacement: true},
			{ID: 13, ApiaryID: 1, Name: "Hive E", Active: true, BoxNeedsAdding: true},
		},
		2: {
			{ID: 20, ApiaryID: 2, Name: "Hive C", Active: true, ReadyForHarvest: true},
		},
	}, diseasesByHiveID: map[int64][]string{11: {"varroa"}}}
	inspections := &mockInspectionLister{byHiveID: map[int64][]*model.Inspection{
		10: {{ID: 100, HiveID: 10}},
	}}
	treatments := &mockTreatmentLister{byHiveID: map[int64][]*model.Treatment{
		11: {{ID: 200, HiveID: 11}},
	}}
	feedings := &mockFeedingLister{byHiveID: map[int64][]*model.Feeding{
		11: {{ID: 300, HiveID: 11}},
	}}
	harvests := &mockHarvestLister{byHiveID: map[int64][]*model.Harvest{
		20: {{ID: 400, HiveID: 20, Kilograms: 12.5}, {ID: 401, HiveID: 20, Kilograms: 7.5}},
	}}
	tools := NewHiveTools(apiaries, hives, inspections, treatments, harvests, feedings)

	result, err := tools.GetDashboardSummary(context.Background(), 99, nil, nil)
	if err != nil {
		t.Fatalf("GetDashboardSummary returned error: %v", err)
	}
	if result.ApiaryCount != 2 {
		t.Errorf("expected apiary_count 2, got %d", result.ApiaryCount)
	}
	if result.HiveCount != 5 {
		t.Errorf("expected hive_count 5, got %d", result.HiveCount)
	}
	if result.QueenNeedsReplacement != 1 || result.NeedsFood != 1 || result.BoxNeedsAdding != 1 || result.Sick != 1 || result.ReadyForHarvest != 1 {
		t.Errorf("unexpected status counts: %+v", result)
	}
	if result.Inspections != 1 || result.Treatments != 1 || result.Feedings != 1 || result.Harvests != 2 {
		t.Errorf("unexpected activity counts: %+v", result)
	}
	if result.HarvestKilograms != 20 {
		t.Errorf("expected 20 kg harvested, got %v", result.HarvestKilograms)
	}
}

func TestGetDashboardSummaryScopedToOneApiary(t *testing.T) {
	apiaries := &mockApiaryLister{}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}},
	}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	apiaryID := int64(1)
	result, err := tools.GetDashboardSummary(context.Background(), 99, &apiaryID, nil)
	if err != nil {
		t.Fatalf("GetDashboardSummary returned error: %v", err)
	}
	if result.ApiaryCount != 1 {
		t.Errorf("expected apiary_count 1 when scoped, got %d", result.ApiaryCount)
	}
	if result.HiveCount != 1 {
		t.Errorf("expected hive_count 1, got %d", result.HiveCount)
	}
}

func TestGetDashboardSummaryExcludesInactiveHives(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Active", Active: true},
			{ID: 11, ApiaryID: 1, Name: "Inactive", Active: false},
		},
	}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	result, err := tools.GetDashboardSummary(context.Background(), 99, nil, nil)
	if err != nil {
		t.Fatalf("GetDashboardSummary returned error: %v", err)
	}
	if result.HiveCount != 1 {
		t.Errorf("expected only the active hive counted, got %d", result.HiveCount)
	}
}

func TestGetDashboardSummaryUnknownApiaryReturnsError(t *testing.T) {
	apiaries := &mockApiaryLister{}
	hives := &mockHiveLister{}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})
	apiaries.getMembershipFn = func(_, _ int64) (*model.Apiary, string, error) {
		return nil, "", gorm.ErrRecordNotFound
	}

	apiaryID := int64(1)
	if _, err := tools.GetDashboardSummary(context.Background(), 99, &apiaryID, nil); err != ErrApiaryNotFound {
		t.Fatalf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestGetDashboardSummaryNoHivesReturnsZeroCounts(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{}
	inspections := &mockInspectionLister{}
	tools := NewHiveTools(apiaries, hives, inspections, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	days := 7
	result, err := tools.GetDashboardSummary(context.Background(), 99, nil, &days)
	if err != nil {
		t.Fatalf("GetDashboardSummary returned error: %v", err)
	}
	if result.ApiaryCount != 1 || result.HiveCount != 0 {
		t.Errorf("expected apiary_count 1 and hive_count 0, got %+v", result)
	}
	if result.QueenNeedsReplacement != 0 || result.NeedsFood != 0 || result.BoxNeedsAdding != 0 || result.Sick != 0 || result.ReadyForHarvest != 0 {
		t.Errorf("expected zero status counts, got %+v", result)
	}
	if result.Inspections != 0 || result.Treatments != 0 || result.Feedings != 0 || result.Harvests != 0 || result.HarvestKilograms != 0 {
		t.Errorf("expected zero activity counts, got %+v", result)
	}

	wantFrom, wantTo := daysToRange(&days)
	const tolerance = time.Second
	if diff := wantFrom.Sub(inspections.gotFrom); diff > tolerance || diff < -tolerance {
		t.Errorf("expected from around %v, got %v", wantFrom, inspections.gotFrom)
	}
	if diff := wantTo.Sub(inspections.gotTo); diff > tolerance || diff < -tolerance {
		t.Errorf("expected to around %v, got %v", wantTo, inspections.gotTo)
	}
}

func TestGetDashboardSummaryToolDispatchesThroughRegistry(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}},
	}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	r := NewRegistry()
	r.Register(tools.GetDashboardSummaryTool())

	out, err := r.Call(context.Background(), 99, "get_dashboard_summary", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	result, ok := out.(DashboardSummary)
	if !ok {
		t.Fatalf("expected DashboardSummary, got %T", out)
	}
	if result.HiveCount != 1 {
		t.Errorf("expected hive_count 1, got %d", result.HiveCount)
	}
}
