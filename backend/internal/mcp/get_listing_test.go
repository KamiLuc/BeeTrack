package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

func TestGetListingReturnsSummary(t *testing.T) {
	batchID := int64(9)
	searcher := &mockListingSearcher{byID: map[int64]*model.Listing{
		1: {ID: 1, Title: "Honey batch", HoneyBatchID: &batchID, HoneyBatchHoneyType: "Acacia"},
	}}
	tools := NewListingTools(searcher)

	result, err := tools.GetListing(context.Background(), 7, 1)
	if err != nil {
		t.Fatalf("GetListing returned error: %v", err)
	}
	if result.Title != "Honey batch" {
		t.Errorf("expected title 'Honey batch', got %q", result.Title)
	}
	if result.HoneyBatch == nil || result.HoneyBatch.HoneyType != "Acacia" {
		t.Errorf("expected honey batch summary, got %+v", result.HoneyBatch)
	}
}

func TestGetListingPropagatesNotFound(t *testing.T) {
	searcher := &mockListingSearcher{byID: map[int64]*model.Listing{}}
	tools := NewListingTools(searcher)

	if _, err := tools.GetListing(context.Background(), 7, 404); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetListingToolRequiresListingID(t *testing.T) {
	searcher := &mockListingSearcher{byID: map[int64]*model.Listing{}}
	tools := NewListingTools(searcher)

	r := NewRegistry()
	r.Register(tools.GetListingTool())

	if _, err := r.Call(context.Background(), 7, "get_listing", json.RawMessage(`{}`)); !errors.Is(err, errListingIDRequired) {
		t.Fatalf("expected errListingIDRequired, got %v", err)
	}
}

func TestGetListingToolDecodesInput(t *testing.T) {
	searcher := &mockListingSearcher{byID: map[int64]*model.Listing{
		5: {ID: 5, Title: "Queen bee"},
	}}
	tools := NewListingTools(searcher)

	r := NewRegistry()
	r.Register(tools.GetListingTool())

	input, _ := json.Marshal(map[string]any{"listing_id": 5})
	out, err := r.Call(context.Background(), 7, "get_listing", input)
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	result, ok := out.(*ListingSummary)
	if !ok {
		t.Fatalf("expected *ListingSummary, got %T", out)
	}
	if result.Title != "Queen bee" {
		t.Errorf("expected title 'Queen bee', got %q", result.Title)
	}
}
