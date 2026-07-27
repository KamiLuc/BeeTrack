package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

var errListingIDRequired = fmt.Errorf("listing_id is required")

// Hidden or not-yet-approved listings are visible only to their owner, same as GET /api/v1/listings/{id}.
func (t *ListingTools) GetListing(ctx context.Context, userID, listingID int64) (*ListingSummary, error) {
	l, err := t.listings.Get(ctx, userID, listingID)
	if err != nil {
		return nil, err
	}
	summary := listingSummary(l)
	return &summary, nil
}

type getListingInput struct {
	ListingID int64 `json:"listing_id"`
}

func (t *ListingTools) GetListingTool() Tool {
	return Tool{
		Name:        "get_listing",
		Description: "Get one marketplace listing's full detail, for a follow-up question about a specific search_listings result.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"listing_id": map[string]any{
					"type":        "integer",
					"description": "The listing to look up.",
				},
			},
			Required: []string{"listing_id"},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in getListingInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode get_listing input: %w", err)
				}
			}
			if in.ListingID <= 0 {
				return nil, errListingIDRequired
			}
			return t.GetListing(ctx, userID, in.ListingID)
		},
	}
}
