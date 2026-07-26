package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

func TestListUntreatedHivesNilDaysMeansNeverTreated(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Never Treated"},
			{ID: 11, ApiaryID: 1, Name: "Treated Long Ago"},
		},
	}}
	longAgo := time.Now().AddDate(-1, 0, 0)
	treatments := &mockTreatmentLister{lastTreatedByID: map[int64]*time.Time{11: &longAgo}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, treatments, &mockHarvestLister{}, &mockFeedingLister{})

	result, err := tools.ListUntreatedHives(context.Background(), 99, nil, nil)
	if err != nil {
		t.Fatalf("ListUntreatedHives returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Never Treated" {
		t.Errorf("expected only the never-treated hive, got %+v", result)
	}
}

func TestListUntreatedHivesWithDaysIncludesStaleTreatments(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Never Treated"},
			{ID: 11, ApiaryID: 1, Name: "Treated Long Ago"},
			{ID: 12, ApiaryID: 1, Name: "Treated Recently"},
		},
	}}
	longAgo := time.Now().AddDate(0, 0, -60)
	recent := time.Now().AddDate(0, 0, -1)
	treatments := &mockTreatmentLister{lastTreatedByID: map[int64]*time.Time{11: &longAgo, 12: &recent}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, treatments, &mockHarvestLister{}, &mockFeedingLister{})

	days := 30
	result, err := tools.ListUntreatedHives(context.Background(), 99, nil, &days)
	if err != nil {
		t.Fatalf("ListUntreatedHives returned error: %v", err)
	}
	names := map[string]bool{}
	for _, r := range result {
		names[r.Name] = true
	}
	if len(result) != 2 || !names["Never Treated"] || !names["Treated Long Ago"] {
		t.Errorf("expected Never Treated + Treated Long Ago, got %+v", result)
	}
}

func TestListUninspectedHivesNilDaysMeansNeverInspected(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Never Inspected"},
			{ID: 11, ApiaryID: 1, Name: "Inspected"},
		},
	}}
	recent := time.Now().AddDate(0, 0, -1)
	inspections := &mockInspectionLister{lastInspectedByID: map[int64]*time.Time{11: &recent}}
	tools := NewHiveTools(apiaries, hives, inspections, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	result, err := tools.ListUninspectedHives(context.Background(), 99, nil, nil)
	if err != nil {
		t.Fatalf("ListUninspectedHives returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Never Inspected" {
		t.Errorf("expected only the never-inspected hive, got %+v", result)
	}
}

func TestListUnfedHivesNilDaysMeansNeverFed(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Never Fed"},
			{ID: 11, ApiaryID: 1, Name: "Fed"},
		},
	}}
	recent := time.Now().AddDate(0, 0, -1)
	feedings := &mockFeedingLister{lastFedByID: map[int64]*time.Time{11: &recent}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, feedings)

	result, err := tools.ListUnfedHives(context.Background(), 99, nil, nil)
	if err != nil {
		t.Fatalf("ListUnfedHives returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Never Fed" {
		t.Errorf("expected only the never-fed hive, got %+v", result)
	}
}

func TestUnrecentHivesToolsDispatchThroughRegistry(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A"}},
	}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	r := NewRegistry()
	r.Register(tools.ListUntreatedHivesTool())
	r.Register(tools.ListUninspectedHivesTool())
	r.Register(tools.ListUnfedHivesTool())

	for _, toolName := range []string{"list_untreated_hives", "list_uninspected_hives", "list_unfed_hives"} {
		result, err := r.Call(context.Background(), 99, toolName, json.RawMessage(`{}`))
		if err != nil {
			t.Errorf("%s: Call returned error: %v", toolName, err)
		}
		if result == nil {
			t.Errorf("%s: expected a non-nil result", toolName)
		}
	}
}

func TestLacksRecentRecord(t *testing.T) {
	now := time.Now()
	old := now.AddDate(0, 0, -10)
	recent := now.AddDate(0, 0, -1)
	cutoff := now.AddDate(0, 0, -5)

	if !lacksRecentRecord(nil, nil) {
		t.Error("no record and no cutoff should lack a recent record")
	}
	if lacksRecentRecord(&recent, nil) {
		t.Error("having any record with a nil cutoff should not lack one")
	}
	if !lacksRecentRecord(&old, &cutoff) {
		t.Error("a record older than cutoff should lack a recent one")
	}
	if lacksRecentRecord(&recent, &cutoff) {
		t.Error("a record newer than cutoff should not lack a recent one")
	}
	if !lacksRecentRecord(nil, &cutoff) {
		t.Error("no record at all should lack a recent one even with a cutoff")
	}
}
