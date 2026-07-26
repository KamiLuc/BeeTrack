package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

func newCompareTestTools(hives ...*model.Hive) (*HiveTools, *mockInspectionLister, *mockTreatmentLister, *mockFeedingLister, *mockHarvestLister) {
	apiaries := &mockApiaryLister{}
	byApiary := map[int64][]*model.Hive{}
	for _, h := range hives {
		byApiary[h.ApiaryID] = append(byApiary[h.ApiaryID], h)
	}
	inspections := &mockInspectionLister{byHiveID: map[int64][]*model.Inspection{}}
	treatments := &mockTreatmentLister{byHiveID: map[int64][]*model.Treatment{}}
	feedings := &mockFeedingLister{byHiveID: map[int64][]*model.Feeding{}}
	harvests := &mockHarvestLister{byHiveID: map[int64][]*model.Harvest{}}
	tools := NewHiveTools(apiaries, &mockHiveLister{hivesByApiary: byApiary}, inspections, treatments, harvests, feedings)
	return tools, inspections, treatments, feedings, harvests
}

func TestCompareHivesReturnsLatestInspectionPerHive(t *testing.T) {
	hiveA := &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}
	hiveB := &model.Hive{ID: 20, ApiaryID: 1, Name: "Hive B", Active: true, Queenless: true}
	tools, inspections, _, _, _ := newCompareTestTools(hiveA, hiveB)
	inspections.byHiveID[10] = []*model.Inspection{{ID: 100, HiveID: 10, BroodPattern: "solid"}}
	inspections.byHiveID[20] = []*model.Inspection{{ID: 200, HiveID: 20, BroodPattern: "spotty"}}

	result, err := tools.CompareHives(context.Background(), 99, []int64{10, 20})
	if err != nil {
		t.Fatalf("CompareHives returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 hives, got %d", len(result))
	}
	if result[0].Name != "Hive A" || result[0].BroodPattern != "solid" {
		t.Errorf("unexpected first hive: %+v", result[0])
	}
	if result[1].Name != "Hive B" || result[1].BroodPattern != "spotty" || !result[1].Queenless {
		t.Errorf("unexpected second hive: %+v", result[1])
	}
}

func TestCompareHivesReturnsLatestTreatmentFeedingHarvest(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}
	tools, _, treatments, feedings, harvests := newCompareTestTools(hive)
	treatments.byHiveID[10] = []*model.Treatment{{ID: 200, HiveID: 10, MedicineName: "Apiwarol"}}
	feedings.byHiveID[10] = []*model.Feeding{{ID: 300, HiveID: 10, FeedType: "sugar syrup"}}
	harvests.byHiveID[10] = []*model.Harvest{{ID: 400, HiveID: 10, Kilograms: 12.5}}

	result, err := tools.CompareHives(context.Background(), 99, []int64{10})
	if err != nil {
		t.Fatalf("CompareHives returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 hive, got %d", len(result))
	}
	h := result[0]
	if h.LastTreatment == nil || h.LastTreatment.MedicineName != "Apiwarol" {
		t.Errorf("unexpected last treatment: %+v", h.LastTreatment)
	}
	if h.LastFeeding == nil || h.LastFeeding.FeedType != "sugar syrup" {
		t.Errorf("unexpected last feeding: %+v", h.LastFeeding)
	}
	if h.LastHarvest == nil || h.LastHarvest.Kilograms != 12.5 {
		t.Errorf("unexpected last harvest: %+v", h.LastHarvest)
	}
}

func TestCompareHivesWithNoInspectionsLeavesInspectionFieldsZero(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}
	tools, _, _, _, _ := newCompareTestTools(hive)

	result, err := tools.CompareHives(context.Background(), 99, []int64{10})
	if err != nil {
		t.Fatalf("CompareHives returned error: %v", err)
	}
	if len(result) != 1 || result[0].LastInspectedAt != nil || result[0].BroodPattern != "" {
		t.Errorf("expected no inspection data, got %+v", result[0])
	}
	if result[0].LastTreatment != nil || result[0].LastFeeding != nil || result[0].LastHarvest != nil {
		t.Errorf("expected no treatment/feeding/harvest data, got %+v", result[0])
	}
}

func TestCompareHivesEmptyHiveIDsReturnsError(t *testing.T) {
	tools, _, _, _, _ := newCompareTestTools()

	_, err := tools.CompareHives(context.Background(), 99, nil)
	if err == nil {
		t.Fatal("expected an error when no hive_ids are given")
	}
}

func TestCompareHivesUnknownHiveReturnsErrHiveNotFound(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}
	tools, _, _, _, _ := newCompareTestTools(hive)

	_, err := tools.CompareHives(context.Background(), 99, []int64{10, 999})
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestCompareHivesToolDispatchesThroughRegistry(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Active: true}
	tools, _, _, _, _ := newCompareTestTools(hive)

	r := NewRegistry()
	r.Register(tools.CompareHivesTool())

	result, err := r.Call(context.Background(), 99, "compare_hives", json.RawMessage(`{"hive_ids":[10]}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	comparisons, ok := result.([]HiveComparison)
	if !ok || len(comparisons) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
}
