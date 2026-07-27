package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
)

// listingCategories mirrors service.validListingCategories (unexported
// there) — the tool schema only needs it for guidance, not validation, since
// an unknown category simply yields zero results, same as the REST endpoint.
var listingCategories = []string{
	"HONEY", "POLLEN", "BEE_COLONIES", "QUEEN_BEES",
	"BEEHIVES", "POPULATED_BEEHIVES", "EQUIPMENT",
	"EXTRACTION_EQUIPMENT", "FEED", "SUPPLIES",
	"WAX_FOUNDATION", "BEESWAX", "PROPOLIS",
	"SERVICES", "OTHER",
}

const (
	defaultSearchListingsLimit = 20
	maxSearchListingsLimit     = 50
)

// ListingSearcher is the read surface of ListingService the listing tools
// need; satisfied by *service.ListingService.
type ListingSearcher interface {
	Search(ctx context.Context, f repository.ListingFilter) ([]*model.Listing, int64, error)
	Get(ctx context.Context, viewerUserID, listingID int64) (*model.Listing, error)
}

type ListingTools struct {
	listings ListingSearcher
}

func NewListingTools(listings ListingSearcher) *ListingTools {
	return &ListingTools{listings: listings}
}

type HoneyBatchSummary struct {
	ID                  int64      `json:"id"`
	HoneyType           string     `json:"honey_type"`
	GatheringDate       *time.Time `json:"gathering_date"`
	AmountGrams         *int64     `json:"amount_grams"`
	ProcessingMethod    string     `json:"processing_method"`
	CertificationStatus string     `json:"certification_status"`
	HasPDF              bool       `json:"has_pdf"`
}

type ListingSummary struct {
	ID              int64              `json:"id"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Category        string             `json:"category"`
	Price           float64            `json:"price"`
	Quantity        string             `json:"quantity"`
	Address         string             `json:"address"`
	Lat             float64            `json:"lat"`
	Lng             float64            `json:"lng"`
	DistanceKm      *float64           `json:"distance_km,omitempty"`
	ApiaryName      string             `json:"apiary_name,omitempty"`
	ContactPhone    string             `json:"contact_phone"`
	ContactEmail    string             `json:"contact_email"`
	ImageCount      int                `json:"image_count"`
	HoneyBatch      *HoneyBatchSummary `json:"honey_batch,omitempty"`
	Status          string             `json:"status"`
	IsHidden        bool               `json:"is_hidden"`
	RejectionReason *string            `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
}

func listingSummary(l *model.Listing) ListingSummary {
	price := 0.0
	if l.Price != nil {
		price = *l.Price
	}
	s := ListingSummary{
		ID:              l.ID,
		Title:           l.Title,
		Description:     l.Description,
		Category:        l.Category,
		Price:           price,
		Quantity:        l.Quantity,
		Address:         l.Address,
		Lat:             l.Lat,
		Lng:             l.Lng,
		DistanceKm:      l.DistanceKm,
		ApiaryName:      l.ApiaryName,
		ContactPhone:    l.ContactPhone,
		ContactEmail:    l.ContactEmail,
		ImageCount:      len(l.Images),
		Status:          l.Status,
		IsHidden:        l.IsHidden,
		RejectionReason: l.RejectionReason,
		CreatedAt:       l.CreatedAt,
	}
	if l.HoneyBatchID != nil {
		s.HoneyBatch = &HoneyBatchSummary{
			ID:                  *l.HoneyBatchID,
			HoneyType:           l.HoneyBatchHoneyType,
			GatheringDate:       l.HoneyBatchGatheringDate,
			AmountGrams:         l.HoneyBatchAmountGrams,
			ProcessingMethod:    l.HoneyBatchProcessingMethod,
			CertificationStatus: l.HoneyBatchCertificationStatus,
			HasPDF:              l.HoneyBatchHasPDF,
		}
	}
	return s
}

// SearchListingsResult matches the REST search endpoint's {"items", "total"} body.
type SearchListingsResult struct {
	Items []ListingSummary `json:"items"`
	Total int64            `json:"total"`
}

type searchListingsInput struct {
	Category  string   `json:"category,omitempty"`
	Keyword   string   `json:"keyword,omitempty"`
	PriceMin  *float64 `json:"price_min,omitempty"`
	PriceMax  *float64 `json:"price_max,omitempty"`
	NearLat   *float64 `json:"near_lat,omitempty"`
	NearLng   *float64 `json:"near_lng,omitempty"`
	RadiusKm  *float64 `json:"radius_km,omitempty"`
	HasApiary bool     `json:"has_apiary,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}

func (in searchListingsInput) toFilter() repository.ListingFilter {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultSearchListingsLimit
	}
	if limit > maxSearchListingsLimit {
		limit = maxSearchListingsLimit
	}
	f := repository.ListingFilter{
		Category:  in.Category,
		Keyword:   in.Keyword,
		PriceMin:  in.PriceMin,
		PriceMax:  in.PriceMax,
		HasApiary: in.HasApiary,
		Limit:     limit,
		Offset:    in.Offset,
	}
	if in.NearLat != nil && in.NearLng != nil && in.RadiusKm != nil {
		f.NearLat = in.NearLat
		f.NearLng = in.NearLng
		f.RadiusKm = in.RadiusKm
	}
	return f
}

// No caller scoping — this is the same public, approved-and-visible listing
// set an anonymous browser would see.
func (t *ListingTools) SearchListings(ctx context.Context, in searchListingsInput) (SearchListingsResult, error) {
	listings, total, err := t.listings.Search(ctx, in.toFilter())
	if err != nil {
		return SearchListingsResult{}, err
	}
	items := make([]ListingSummary, len(listings))
	for i, l := range listings {
		items[i] = listingSummary(l)
	}
	return SearchListingsResult{Items: items, Total: total}, nil
}

func (t *ListingTools) SearchListingsTool() Tool {
	return Tool{
		Name:        "search_listings",
		Description: "Search public marketplace listings by category, keyword, price range, and/or location radius. Mirrors the Marketplace screen's own filters.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"category": map[string]any{
					"type":        "string",
					"enum":        listingCategories,
					"description": "Restrict to this listing category.",
				},
				"keyword": map[string]any{
					"type":        "string",
					"description": "Matched against listing title and description.",
				},
				"price_min": map[string]any{
					"type":        "number",
					"description": "Minimum price.",
				},
				"price_max": map[string]any{
					"type":        "number",
					"description": "Maximum price.",
				},
				"near_lat": map[string]any{
					"type":        "number",
					"description": "Latitude of the search center; must be given together with near_lng and radius_km.",
				},
				"near_lng": map[string]any{
					"type":        "number",
					"description": "Longitude of the search center; must be given together with near_lat and radius_km.",
				},
				"radius_km": map[string]any{
					"type":        "number",
					"description": "Search radius in kilometers around (near_lat, near_lng).",
				},
				"has_apiary": map[string]any{
					"type":        "boolean",
					"description": "Only listings linked to an apiary.",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Max results to return (default 20, capped at 50).",
				},
				"offset": map[string]any{
					"type":        "integer",
					"description": "Pagination offset.",
				},
			},
		},
		Handler: func(ctx context.Context, _ int64, input json.RawMessage) (any, error) {
			var in searchListingsInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode search_listings input: %w", err)
				}
			}
			return t.SearchListings(ctx, in)
		},
	}
}
