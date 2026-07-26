package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

type mockInspectionLister struct {
	byHiveID          map[int64][]*model.Inspection
	lastInspectedByID map[int64]*time.Time
	gotFrom           time.Time
	gotTo             time.Time
}

func (m *mockInspectionLister) ListByHiveIDsAndRange(_ context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Inspection, error) {
	m.gotFrom, m.gotTo = from, to
	return m.byHiveID[hiveIDs[0]], nil
}

func (m *mockInspectionLister) LastInspectionDatesByHiveIDs(_ context.Context, ids []int64) (map[int64]*time.Time, error) {
	out := make(map[int64]*time.Time, len(ids))
	for _, id := range ids {
		if t, ok := m.lastInspectedByID[id]; ok {
			out[id] = t
		}
	}
	return out, nil
}

type mockTreatmentLister struct {
	byHiveID        map[int64][]*model.Treatment
	lastTreatedByID map[int64]*time.Time
}

func (m *mockTreatmentLister) ListByHiveIDsAndRange(_ context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Treatment, error) {
	return m.byHiveID[hiveIDs[0]], nil
}

func (m *mockTreatmentLister) LastTreatmentDatesByHiveIDs(_ context.Context, ids []int64) (map[int64]*time.Time, error) {
	out := make(map[int64]*time.Time, len(ids))
	for _, id := range ids {
		if t, ok := m.lastTreatedByID[id]; ok {
			out[id] = t
		}
	}
	return out, nil
}

type mockHarvestLister struct {
	byHiveID map[int64][]*model.Harvest
}

func (m *mockHarvestLister) ListByHiveIDsAndRange(_ context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Harvest, error) {
	return m.byHiveID[hiveIDs[0]], nil
}

type mockFeedingLister struct {
	byHiveID    map[int64][]*model.Feeding
	lastFedByID map[int64]*time.Time
}

func (m *mockFeedingLister) ListByHiveIDsAndRange(_ context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Feeding, error) {
	return m.byHiveID[hiveIDs[0]], nil
}

func (m *mockFeedingLister) LastFeedingDatesByHiveIDs(_ context.Context, ids []int64) (map[int64]*time.Time, error) {
	out := make(map[int64]*time.Time, len(ids))
	for _, id := range ids {
		if t, ok := m.lastFedByID[id]; ok {
			out[id] = t
		}
	}
	return out, nil
}

func newTestHiveTools(hive *model.Hive) (*HiveTools, *mockInspectionLister, *mockTreatmentLister, *mockHarvestLister, *mockFeedingLister) {
	apiaries := &mockApiaryLister{}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{hive.ApiaryID: {hive}}}
	inspections := &mockInspectionLister{byHiveID: map[int64][]*model.Inspection{}}
	treatments := &mockTreatmentLister{byHiveID: map[int64][]*model.Treatment{}}
	harvests := &mockHarvestLister{byHiveID: map[int64][]*model.Harvest{}}
	feedings := &mockFeedingLister{byHiveID: map[int64][]*model.Feeding{}}
	tools := NewHiveTools(apiaries, hives, inspections, treatments, harvests, feedings)
	return tools, inspections, treatments, harvests, feedings
}

func TestListInspectionsReturnsMappedSummaries(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	tools, inspections, _, _, _ := newTestHiveTools(hive)
	inspections.byHiveID[10] = []*model.Inspection{
		{ID: 100, HiveID: 10, QueenStatus: "seen", Notes: "looked good"},
	}

	result, err := tools.ListInspections(context.Background(), 99, 10, nil)
	if err != nil {
		t.Fatalf("ListInspections returned error: %v", err)
	}
	if len(result) != 1 || result[0].QueenStatus != "seen" || result[0].Notes != "looked good" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestListInspectionsUnknownHiveReturnsErrHiveNotFound(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	tools, _, _, _, _ := newTestHiveTools(hive)

	_, err := tools.ListInspections(context.Background(), 99, 999, nil)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestDaysToRangeNilMeansSinceEver(t *testing.T) {
	from, to := daysToRange(nil)
	if !from.IsZero() {
		t.Errorf("expected zero from, got %v", from)
	}
	if to.Before(time.Now()) {
		t.Errorf("expected to to be at or after now, got %v", to)
	}
}

func TestDaysToRangeWithDaysLooksBackThatFar(t *testing.T) {
	days := 7
	from, to := daysToRange(&days)
	wantFrom := to.AddDate(0, 0, -7)
	if !from.Equal(wantFrom) {
		t.Errorf("expected from %v, got %v", wantFrom, from)
	}
}

func TestGetHiveSummaryAggregatesAllRecordTypes(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", Queenless: true}
	tools, inspections, treatments, harvests, feedings := newTestHiveTools(hive)
	inspections.byHiveID[10] = []*model.Inspection{{ID: 100, HiveID: 10}}
	treatments.byHiveID[10] = []*model.Treatment{{ID: 200, HiveID: 10, MedicineName: "Apiwarol"}}
	harvests.byHiveID[10] = []*model.Harvest{{ID: 300, HiveID: 10, Kilograms: 12.5}}
	feedings.byHiveID[10] = []*model.Feeding{{ID: 400, HiveID: 10, FeedType: "sugar syrup"}}

	result, err := tools.GetHiveSummary(context.Background(), 99, 10, nil)
	if err != nil {
		t.Fatalf("GetHiveSummary returned error: %v", err)
	}
	if result.Hive.Name != "Hive A" || !result.Hive.Queenless {
		t.Errorf("unexpected hive summary: %+v", result.Hive)
	}
	if len(result.Inspections) != 1 || len(result.Treatments) != 1 || len(result.Harvests) != 1 || len(result.Feedings) != 1 {
		t.Errorf("expected one of each record type, got %+v", result)
	}
	if result.Treatments[0].MedicineName != "Apiwarol" {
		t.Errorf("unexpected treatment: %+v", result.Treatments[0])
	}
}

func TestGetHiveSummaryUnknownHiveReturnsErrHiveNotFound(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	tools, _, _, _, _ := newTestHiveTools(hive)

	_, err := tools.GetHiveSummary(context.Background(), 99, 999, nil)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestDecodeHiveDaysInputMissingHiveIDReturnsError(t *testing.T) {
	if _, err := decodeHiveDaysInput(json.RawMessage(`{}`)); err == nil {
		t.Fatal("expected an error when hive_id is missing")
	}
	if _, err := decodeHiveDaysInput(nil); err == nil {
		t.Fatal("expected an error when input is empty")
	}
}

func TestHiveHistoryToolsDispatchThroughRegistry(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	tools, inspections, treatments, harvests, feedings := newTestHiveTools(hive)
	inspections.byHiveID[10] = []*model.Inspection{{ID: 100, HiveID: 10}}
	treatments.byHiveID[10] = []*model.Treatment{{ID: 200, HiveID: 10}}
	harvests.byHiveID[10] = []*model.Harvest{{ID: 300, HiveID: 10}}
	feedings.byHiveID[10] = []*model.Feeding{{ID: 400, HiveID: 10}}

	r := NewRegistry()
	r.Register(tools.ListInspectionsTool())
	r.Register(tools.ListTreatmentsTool())
	r.Register(tools.ListHarvestsTool())
	r.Register(tools.ListFeedingsTool())
	r.Register(tools.GetHiveSummaryTool())

	cases := []struct {
		tool string
	}{
		{"list_inspections"}, {"list_treatments"}, {"list_harvests"}, {"list_feedings"}, {"get_hive_summary"},
	}
	for _, c := range cases {
		result, err := r.Call(context.Background(), 99, c.tool, json.RawMessage(`{"hive_id":10}`))
		if err != nil {
			t.Errorf("%s: Call returned error: %v", c.tool, err)
		}
		if result == nil {
			t.Errorf("%s: expected a non-nil result", c.tool)
		}
	}
}

func TestHiveHistoryToolsMissingHiveIDReturnsError(t *testing.T) {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	tools, _, _, _, _ := newTestHiveTools(hive)

	r := NewRegistry()
	r.Register(tools.ListInspectionsTool())

	if _, err := r.Call(context.Background(), 99, "list_inspections", json.RawMessage(`{}`)); err == nil {
		t.Fatal("expected an error when hive_id is missing")
	}
}
