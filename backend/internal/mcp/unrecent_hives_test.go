package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

func TestListHivesMissingRecordsNilDaysMeansNeverHadOne(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Never Treated", Active: true},
			{ID: 11, ApiaryID: 1, Name: "Treated Long Ago", Active: true},
		},
	}}
	longAgo := time.Now().AddDate(-1, 0, 0)
	treatments := &mockTreatmentLister{lastTreatedByID: map[int64]*time.Time{11: &longAgo}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, treatments, &mockHarvestLister{}, &mockFeedingLister{})

	result, err := tools.ListHivesMissingRecords(context.Background(), 99, nil, []string{"treatment"}, nil)
	if err != nil {
		t.Fatalf("ListHivesMissingRecords returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Never Treated" {
		t.Errorf("expected only the never-treated hive, got %+v", result)
	}
}

func TestListHivesMissingRecordsWithDaysIncludesStaleRecords(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {
			{ID: 10, ApiaryID: 1, Name: "Never Treated", Active: true},
			{ID: 11, ApiaryID: 1, Name: "Treated Long Ago", Active: true},
			{ID: 12, ApiaryID: 1, Name: "Treated Recently", Active: true},
		},
	}}
	longAgo := time.Now().AddDate(0, 0, -60)
	recent := time.Now().AddDate(0, 0, -1)
	treatments := &mockTreatmentLister{lastTreatedByID: map[int64]*time.Time{11: &longAgo, 12: &recent}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, treatments, &mockHarvestLister{}, &mockFeedingLister{})

	days := 30
	result, err := tools.ListHivesMissingRecords(context.Background(), 99, nil, []string{"treatment"}, &days)
	if err != nil {
		t.Fatalf("ListHivesMissingRecords returned error: %v", err)
	}
	names := map[string]bool{}
	for _, r := range result {
		names[r.Name] = true
	}
	if len(result) != 2 || !names["Never Treated"] || !names["Treated Long Ago"] {
		t.Errorf("expected Never Treated + Treated Long Ago, got %+v", result)
	}
}

func TestListHivesMissingRecordsWithNoRecordTypesChecksAllThree(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Fully Attended", Active: true}},
	}}
	now := time.Now()
	treatments := &mockTreatmentLister{lastTreatedByID: map[int64]*time.Time{10: &now}}
	inspections := &mockInspectionLister{lastInspectedByID: map[int64]*time.Time{10: &now}}
	feedings := &mockFeedingLister{lastFedByID: map[int64]*time.Time{10: &now}}
	tools := NewHiveTools(apiaries, hives, inspections, treatments, &mockHarvestLister{}, feedings)

	result, err := tools.ListHivesMissingRecords(context.Background(), 99, nil, nil, nil)
	if err != nil {
		t.Fatalf("ListHivesMissingRecords returned error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected no hives missing anything, got %+v", result)
	}
}

func TestListHivesMissingRecordsReportsWhichTypesAreMissing(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}},
	}}
	now := time.Now()
	treatments := &mockTreatmentLister{lastTreatedByID: map[int64]*time.Time{10: &now}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, treatments, &mockHarvestLister{}, &mockFeedingLister{})

	result, err := tools.ListHivesMissingRecords(context.Background(), 99, nil, []string{"treatment", "inspection", "feeding"}, nil)
	if err != nil {
		t.Fatalf("ListHivesMissingRecords returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 hive, got %+v", result)
	}
	if _, treated := result[0].LastRecordedAt["treatment"]; treated {
		t.Error("did not expect treatment to be reported as missing")
	}
	if _, inspected := result[0].LastRecordedAt["inspection"]; !inspected {
		t.Error("expected inspection to be reported as missing")
	}
	if _, fed := result[0].LastRecordedAt["feeding"]; !fed {
		t.Error("expected feeding to be reported as missing")
	}
}

func TestListHivesMissingRecordsInvalidRecordTypeReturnsError(t *testing.T) {
	apiaries := &mockApiaryLister{}
	tools := NewHiveTools(apiaries, &mockHiveLister{}, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	_, err := tools.ListHivesMissingRecords(context.Background(), 99, nil, []string{"harvest"}, nil)
	if err == nil {
		t.Fatal("expected an error for an invalid record type (harvest isn't supported here)")
	}
}

func TestListHivesMissingRecordsToolDispatchesThroughRegistry(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}},
	}}
	tools := NewHiveTools(apiaries, hives, &mockInspectionLister{}, &mockTreatmentLister{}, &mockHarvestLister{}, &mockFeedingLister{})

	r := NewRegistry()
	r.Register(tools.ListHivesMissingRecordsTool())

	result, err := r.Call(context.Background(), 99, "list_hives_missing_records", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected a non-nil result")
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
