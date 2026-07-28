package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

func TestListApiariesMapsMembershipsToSummaries(t *testing.T) {
	lastInspected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	apiaries := &mockApiaryLister{
		memberships: []model.ApiaryMembership{
			{Apiary: &model.Apiary{ID: 1, Name: "Apiary A"}, UserRole: "owner", HiveCount: 3, LastInspectedAt: &lastInspected},
			{Apiary: &model.Apiary{ID: 2, Name: "Apiary B"}, UserRole: "member", HiveCount: 0, LastInspectedAt: nil},
		},
	}
	tools := NewHiveTools(apiaries, &mockHiveLister{}, nil, nil, nil, nil)

	result, err := tools.ListApiaries(context.Background(), 99)
	if err != nil {
		t.Fatalf("ListApiaries returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 apiaries, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Name != "Apiary A" || result[0].Role != "owner" || result[0].HiveCount != 3 {
		t.Errorf("unexpected first summary: %+v", result[0])
	}
	if result[0].LastInspectedAt == nil || !result[0].LastInspectedAt.Equal(lastInspected) {
		t.Errorf("expected LastInspectedAt %v, got %v", lastInspected, result[0].LastInspectedAt)
	}
}

func TestListApiariesIncludesApiariesWithZeroHives(t *testing.T) {
	apiaries := &mockApiaryLister{
		memberships: []model.ApiaryMembership{
			{Apiary: &model.Apiary{ID: 1, Name: "Empty Apiary"}, UserRole: "owner", HiveCount: 0},
		},
	}
	tools := NewHiveTools(apiaries, &mockHiveLister{}, nil, nil, nil, nil)

	result, err := tools.ListApiaries(context.Background(), 99)
	if err != nil {
		t.Fatalf("ListApiaries returned error: %v", err)
	}
	if len(result) != 1 || result[0].HiveCount != 0 {
		t.Fatalf("expected apiary with zero hives to be included, got %+v", result)
	}
	if result[0].Name != "Empty Apiary" {
		t.Errorf("unexpected name: %s", result[0].Name)
	}
}

func TestListApiariesPropagatesError(t *testing.T) {
	wantErr := errors.New("db down")
	apiaries := &mockApiaryLister{membershipErr: wantErr}
	tools := NewHiveTools(apiaries, &mockHiveLister{}, nil, nil, nil, nil)

	_, err := tools.ListApiaries(context.Background(), 99)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected error wrapping %v, got %v", wantErr, err)
	}
}

func TestListApiariesToolDispatchesThroughRegistry(t *testing.T) {
	apiaries := &mockApiaryLister{
		memberships: []model.ApiaryMembership{
			{Apiary: &model.Apiary{ID: 1, Name: "Apiary A"}, UserRole: "owner", HiveCount: 1},
		},
	}
	tools := NewHiveTools(apiaries, &mockHiveLister{}, nil, nil, nil, nil)

	r := NewRegistry()
	r.Register(tools.ListApiariesTool())

	result, err := r.Call(context.Background(), 99, "list_apiaries", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	summaries, ok := result.([]ApiarySummary)
	if !ok || len(summaries) != 1 || summaries[0].Name != "Apiary A" {
		t.Errorf("unexpected result: %+v", result)
	}
}
