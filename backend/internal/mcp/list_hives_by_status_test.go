package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

func newStatusTestTools(hives ...*model.Hive) *HiveTools {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	byApiary := map[int64][]*model.Hive{1: hives}
	return NewHiveTools(apiaries, &mockHiveLister{hivesByApiary: byApiary}, nil, nil, nil, nil)
}

func TestListHivesByStatusFiltersToOneStatus(t *testing.T) {
	tools := newStatusTestTools(
		&model.Hive{ID: 10, ApiaryID: 1, Name: "Queenless", Active: true, Queenless: true},
		&model.Hive{ID: 11, ApiaryID: 1, Name: "Fine", Active: true},
	)

	result, err := tools.ListHivesByStatus(context.Background(), 99, nil, []string{"queenless"})
	if err != nil {
		t.Fatalf("ListHivesByStatus returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Queenless" {
		t.Errorf("expected only the queenless hive, got %+v", result)
	}
}

func TestListHivesByStatusMatchesAnyOfMultipleStatuses(t *testing.T) {
	tools := newStatusTestTools(
		&model.Hive{ID: 10, ApiaryID: 1, Name: "Queenless", Active: true, Queenless: true},
		&model.Hive{ID: 11, ApiaryID: 1, Name: "Needs Food", Active: true, NeedsFood: true},
		&model.Hive{ID: 12, ApiaryID: 1, Name: "Fine", Active: true},
	)

	result, err := tools.ListHivesByStatus(context.Background(), 99, nil, []string{"queenless", "needs_food"})
	if err != nil {
		t.Fatalf("ListHivesByStatus returned error: %v", err)
	}
	names := map[string]bool{}
	for _, r := range result {
		names[r.Name] = true
	}
	if len(result) != 2 || !names["Queenless"] || !names["Needs Food"] {
		t.Errorf("expected Queenless + Needs Food, got %+v", result)
	}
}

func TestListHivesByStatusSickMeansHasADisease(t *testing.T) {
	apiaries := &mockApiaryLister{memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}}}
	hives := &mockHiveLister{
		hivesByApiary: map[int64][]*model.Hive{
			1: {
				{ID: 10, ApiaryID: 1, Name: "Sick", Active: true},
				{ID: 11, ApiaryID: 1, Name: "Healthy", Active: true},
			},
		},
		diseasesByHiveID: map[int64][]string{10: {"varroa"}},
	}
	tools := NewHiveTools(apiaries, hives, nil, nil, nil, nil)

	result, err := tools.ListHivesByStatus(context.Background(), 99, nil, []string{"sick"})
	if err != nil {
		t.Fatalf("ListHivesByStatus returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Sick" {
		t.Errorf("expected only the sick hive, got %+v", result)
	}
}

func TestListHivesByStatusWithNoStatusesMatchesAnyFlag(t *testing.T) {
	tools := newStatusTestTools(
		&model.Hive{ID: 10, ApiaryID: 1, Name: "Ready", Active: true, ReadyForHarvest: true},
		&model.Hive{ID: 11, ApiaryID: 1, Name: "Fine", Active: true},
	)

	result, err := tools.ListHivesByStatus(context.Background(), 99, nil, nil)
	if err != nil {
		t.Fatalf("ListHivesByStatus returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Ready" {
		t.Errorf("expected only the ready-for-harvest hive, got %+v", result)
	}
}

func TestListHivesByStatusExcludesInactiveHives(t *testing.T) {
	tools := newStatusTestTools(
		&model.Hive{ID: 10, ApiaryID: 1, Name: "Inactive Queenless", Active: false, Queenless: true},
	)

	result, err := tools.ListHivesByStatus(context.Background(), 99, nil, []string{"queenless"})
	if err != nil {
		t.Fatalf("ListHivesByStatus returned error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected inactive hive to be excluded, got %+v", result)
	}
}

func TestListHivesByStatusInvalidStatusReturnsError(t *testing.T) {
	tools := newStatusTestTools()

	_, err := tools.ListHivesByStatus(context.Background(), 99, nil, []string{"nonsense"})
	if err == nil {
		t.Fatal("expected an error for an invalid status")
	}
}

func TestListHivesByStatusToolDispatchesThroughRegistry(t *testing.T) {
	tools := newStatusTestTools(&model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true, Queenless: true})

	r := NewRegistry()
	r.Register(tools.ListHivesByStatusTool())

	result, err := r.Call(context.Background(), 99, "list_hives_by_status", json.RawMessage(`{"statuses":["queenless"]}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	summaries, ok := result.([]HiveSummary)
	if !ok || len(summaries) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
}
