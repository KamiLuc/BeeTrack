package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
)

type mockListingSearcher struct {
	listings   []*model.Listing
	total      int64
	searchErr  error
	byID       map[int64]*model.Listing
	getErr     error
	lastFilter repository.ListingFilter
}

func (m *mockListingSearcher) Search(_ context.Context, f repository.ListingFilter) ([]*model.Listing, int64, error) {
	m.lastFilter = f
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.listings, m.total, nil
}

func (m *mockListingSearcher) Get(_ context.Context, _, listingID int64) (*model.Listing, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	l, ok := m.byID[listingID]
	if !ok {
		return nil, errors.New("listing not found")
	}
	return l, nil
}

func TestSearchListingsReturnsItemsAndTotal(t *testing.T) {
	price := 42.5
	searcher := &mockListingSearcher{
		listings: []*model.Listing{
			{ID: 1, Title: "Dadant frames", Category: "EQUIPMENT", Price: &price},
		},
		total: 5,
	}
	tools := NewListingTools(searcher)

	result, err := tools.SearchListings(context.Background(), searchListingsInput{Category: "EQUIPMENT"})
	if err != nil {
		t.Fatalf("SearchListings returned error: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
	if len(result.Items) != 1 || result.Items[0].Title != "Dadant frames" || result.Items[0].Price != 42.5 {
		t.Errorf("unexpected items: %+v", result.Items)
	}
	if searcher.lastFilter.Category != "EQUIPMENT" {
		t.Errorf("expected filter category EQUIPMENT, got %q", searcher.lastFilter.Category)
	}
}

func TestSearchListingsEmptyResults(t *testing.T) {
	searcher := &mockListingSearcher{listings: nil, total: 0}
	tools := NewListingTools(searcher)

	result, err := tools.SearchListings(context.Background(), searchListingsInput{Category: "OTHER"})
	if err != nil {
		t.Fatalf("SearchListings returned error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if result.Items == nil || len(result.Items) != 0 {
		t.Errorf("expected empty non-nil items slice, got %+v", result.Items)
	}
}

func TestSearchListingsDefaultsAndCapsLimit(t *testing.T) {
	searcher := &mockListingSearcher{}
	tools := NewListingTools(searcher)

	if _, err := tools.SearchListings(context.Background(), searchListingsInput{}); err != nil {
		t.Fatalf("SearchListings returned error: %v", err)
	}
	if searcher.lastFilter.Limit != defaultSearchListingsLimit {
		t.Errorf("expected default limit %d, got %d", defaultSearchListingsLimit, searcher.lastFilter.Limit)
	}

	if _, err := tools.SearchListings(context.Background(), searchListingsInput{Limit: 500}); err != nil {
		t.Fatalf("SearchListings returned error: %v", err)
	}
	if searcher.lastFilter.Limit != maxSearchListingsLimit {
		t.Errorf("expected capped limit %d, got %d", maxSearchListingsLimit, searcher.lastFilter.Limit)
	}
}

func TestSearchListingsIgnoresPartialNearFilter(t *testing.T) {
	searcher := &mockListingSearcher{}
	tools := NewListingTools(searcher)

	lat := 50.06
	if _, err := tools.SearchListings(context.Background(), searchListingsInput{NearLat: &lat}); err != nil {
		t.Fatalf("SearchListings returned error: %v", err)
	}
	if searcher.lastFilter.NearLat != nil {
		t.Errorf("expected near filter to be dropped when incomplete, got %+v", searcher.lastFilter)
	}
}

func TestSearchListingsPropagatesSearchError(t *testing.T) {
	searcher := &mockListingSearcher{searchErr: errors.New("boom")}
	tools := NewListingTools(searcher)

	if _, err := tools.SearchListings(context.Background(), searchListingsInput{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSearchListingsToolDecodesInput(t *testing.T) {
	searcher := &mockListingSearcher{listings: []*model.Listing{{ID: 1}}, total: 1}
	tools := NewListingTools(searcher)

	r := NewRegistry()
	r.Register(tools.SearchListingsTool())

	input, _ := json.Marshal(map[string]any{"keyword": "honey"})
	out, err := r.Call(context.Background(), 7, "search_listings", input)
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	result, ok := out.(SearchListingsResult)
	if !ok {
		t.Fatalf("expected SearchListingsResult, got %T", out)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if searcher.lastFilter.Keyword != "honey" {
		t.Errorf("expected keyword 'honey', got %q", searcher.lastFilter.Keyword)
	}
}
