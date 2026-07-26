package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockApiaryLister struct {
	memberships     []model.ApiaryMembership
	membershipErr   error
	getMembershipFn func(apiaryID, userID int64) (*model.Apiary, string, error)
}

func (m *mockApiaryLister) GetMembership(_ context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	if m.getMembershipFn != nil {
		return m.getMembershipFn(apiaryID, userID)
	}
	return &model.Apiary{ID: apiaryID}, "owner", nil
}

func (m *mockApiaryLister) ListByUserID(_ context.Context, userID int64) ([]model.ApiaryMembership, error) {
	return m.memberships, m.membershipErr
}

type mockHiveLister struct {
	hivesByApiary    map[int64][]*model.Hive
	diseasesByHiveID map[int64][]string
}

func (m *mockHiveLister) ListByApiaryID(_ context.Context, apiaryID int64) ([]*model.Hive, error) {
	return m.hivesByApiary[apiaryID], nil
}

func (m *mockHiveLister) GetByID(_ context.Context, hiveID int64) (*model.Hive, error) {
	for _, hives := range m.hivesByApiary {
		for _, h := range hives {
			if h.ID == hiveID {
				return h, nil
			}
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockHiveLister) ListDiseasesByHiveIDs(_ context.Context, ids []int64) ([]*model.HiveDisease, error) {
	var diseases []*model.HiveDisease
	for _, id := range ids {
		for _, disease := range m.diseasesByHiveID[id] {
			diseases = append(diseases, &model.HiveDisease{HiveID: id, Disease: disease})
		}
	}
	return diseases, nil
}

func TestListHivesWithoutFilterReturnsHivesAcrossAllApiaries(t *testing.T) {
	apiaries := &mockApiaryLister{
		memberships: []model.ApiaryMembership{
			{Apiary: &model.Apiary{ID: 1}},
			{Apiary: &model.Apiary{ID: 2}},
		},
	}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A"}},
		2: {{ID: 20, ApiaryID: 2, Name: "Hive B", Queenless: true}},
	}}
	tools := NewHiveTools(apiaries, hives, nil, nil, nil, nil)

	result, err := tools.ListHives(context.Background(), 99, nil)
	if err != nil {
		t.Fatalf("ListHives returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 hives, got %d", len(result))
	}
	if result[0].Name != "Hive A" || result[1].Name != "Hive B" || !result[1].Queenless {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestListHivesAttachesDiseasesPerHive(t *testing.T) {
	apiaries := &mockApiaryLister{
		memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}},
	}
	hives := &mockHiveLister{
		hivesByApiary: map[int64][]*model.Hive{
			1: {{ID: 10, ApiaryID: 1, Name: "Hive A"}},
		},
		diseasesByHiveID: map[int64][]string{
			10: {"varroa", "nosema"},
		},
	}
	tools := NewHiveTools(apiaries, hives, nil, nil, nil, nil)

	result, err := tools.ListHives(context.Background(), 99, nil)
	if err != nil {
		t.Fatalf("ListHives returned error: %v", err)
	}
	if len(result) != 1 || len(result[0].Diseases) != 2 {
		t.Fatalf("expected Hive A to have 2 diseases, got %+v", result)
	}
	if result[0].Diseases[0] != "varroa" || result[0].Diseases[1] != "nosema" {
		t.Errorf("unexpected diseases: %v", result[0].Diseases)
	}
}

func TestListHivesWithApiaryIDFiltersToThatApiary(t *testing.T) {
	apiaries := &mockApiaryLister{}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A"}},
		2: {{ID: 20, ApiaryID: 2, Name: "Hive B"}},
	}}
	tools := NewHiveTools(apiaries, hives, nil, nil, nil, nil)

	apiaryID := int64(1)
	result, err := tools.ListHives(context.Background(), 99, &apiaryID)
	if err != nil {
		t.Fatalf("ListHives returned error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Hive A" {
		t.Errorf("expected only Hive A, got %+v", result)
	}
}

func TestListHivesWithApiaryIDNotAMemberReturnsErrApiaryNotFound(t *testing.T) {
	apiaries := &mockApiaryLister{
		getMembershipFn: func(apiaryID, userID int64) (*model.Apiary, string, error) {
			return nil, "", gorm.ErrRecordNotFound
		},
	}
	tools := NewHiveTools(apiaries, &mockHiveLister{}, nil, nil, nil, nil)

	apiaryID := int64(1)
	_, err := tools.ListHives(context.Background(), 99, &apiaryID)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestHiveToolsToolDispatchesThroughRegistry(t *testing.T) {
	apiaries := &mockApiaryLister{}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A"}},
	}}
	tools := NewHiveTools(apiaries, hives, nil, nil, nil, nil)

	r := NewRegistry()
	r.Register(tools.ListHivesTool())

	result, err := r.Call(context.Background(), 99, "list_hives", json.RawMessage(`{"apiary_id":1}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	summaries, ok := result.([]HiveSummary)
	if !ok || len(summaries) != 1 || summaries[0].Name != "Hive A" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestHiveToolsToolWithEmptyInputListsAllHives(t *testing.T) {
	apiaries := &mockApiaryLister{
		memberships: []model.ApiaryMembership{{Apiary: &model.Apiary{ID: 1}}},
	}
	hives := &mockHiveLister{hivesByApiary: map[int64][]*model.Hive{
		1: {{ID: 10, ApiaryID: 1, Name: "Hive A"}},
	}}
	tools := NewHiveTools(apiaries, hives, nil, nil, nil, nil)

	r := NewRegistry()
	r.Register(tools.ListHivesTool())

	result, err := r.Call(context.Background(), 99, "list_hives", nil)
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	summaries, ok := result.([]HiveSummary)
	if !ok || len(summaries) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
}
