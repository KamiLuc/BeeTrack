package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// ListingHandler handles HTTP requests for marketplace listing resources.
type ListingHandler struct {
	listings *service.ListingService
}

// NewListingHandler creates a ListingHandler backed by svc.
func NewListingHandler(listings *service.ListingService) *ListingHandler {
	return &ListingHandler{listings: listings}
}

type listingRequest struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Price        *float64 `json:"price"`
	Quantity     string   `json:"quantity"`
	Address      string   `json:"address"`
	Lat          *float64 `json:"lat"`
	Lng          *float64 `json:"lng"`
	ApiaryID     *int64   `json:"apiary_id"`
	ContactPhone string   `json:"contact_phone"`
	ContactEmail string   `json:"contact_email"`
	ImageURLs    []string `json:"image_urls"`
}

func (req listingRequest) toParams() service.ListingParams {
	return service.ListingParams{
		Title:        req.Title,
		Description:  req.Description,
		Category:     req.Category,
		Price:        req.Price,
		Quantity:     req.Quantity,
		Address:      req.Address,
		Lat:          req.Lat,
		Lng:          req.Lng,
		ApiaryID:     req.ApiaryID,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		ImageURLs:    req.ImageURLs,
	}
}

type hideRequest struct {
	Hidden bool `json:"hidden"`
}

func listingImageJSON(img model.ListingImage) map[string]any {
	return map[string]any{
		"id":            img.ID,
		"listing_id":    img.ListingID,
		"url":           fmt.Sprintf("/api/v1/listings/%d/images/%d/file", img.ListingID, img.ID),
		"display_order": img.DisplayOrder,
		"created_at":    img.CreatedAt,
	}
}

func listingJSON(l *model.Listing) map[string]any {
	var apiaryName any
	if l.ApiaryName != "" {
		apiaryName = l.ApiaryName
	}
	images := make([]map[string]any, len(l.Images))
	for i, img := range l.Images {
		images[i] = listingImageJSON(img)
	}
	return map[string]any{
		"id":            l.ID,
		"user_id":       l.UserID,
		"title":         l.Title,
		"description":   l.Description,
		"category":      l.Category,
		"price":         l.Price,
		"quantity":      l.Quantity,
		"address":       l.Address,
		"lat":           l.Lat,
		"lng":           l.Lng,
		"distance_km":   l.DistanceKm,
		"apiary_id":     l.ApiaryID,
		"apiary_name":   apiaryName,
		"contact_phone": l.ContactPhone,
		"contact_email": l.ContactEmail,
		"is_hidden":     l.IsHidden,
		"created_at":    l.CreatedAt,
		"updated_at":    l.UpdatedAt,
		"images":        images,
	}
}

func listingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrListingNotFound):
		respond.Error(w, http.StatusNotFound, "LISTING_NOT_FOUND", "listing not found")
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrNotListingOwner):
		respond.Error(w, http.StatusForbidden, "NOT_OWNER", "not the listing owner")
	case errors.Is(err, service.ErrListingTitleRequired):
		respond.Error(w, http.StatusBadRequest, "TITLE_REQUIRED", err.Error())
	case errors.Is(err, service.ErrListingTitleTooLong):
		respond.Error(w, http.StatusBadRequest, "TITLE_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrListingCategoryInvalid):
		respond.Error(w, http.StatusBadRequest, "CATEGORY_INVALID", err.Error())
	case errors.Is(err, service.ErrListingTooManyImages):
		respond.Error(w, http.StatusBadRequest, "TOO_MANY_IMAGES", err.Error())
	case errors.Is(err, service.ErrListingDescriptionTooLong):
		respond.Error(w, http.StatusBadRequest, "DESCRIPTION_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrListingQuantityTooLong):
		respond.Error(w, http.StatusBadRequest, "QUANTITY_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrListingAddressTooLong):
		respond.Error(w, http.StatusBadRequest, "ADDRESS_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrListingLocationRequired):
		respond.Error(w, http.StatusBadRequest, "LOCATION_REQUIRED", err.Error())
	case errors.Is(err, service.ErrInvalidGPS):
		respond.Error(w, http.StatusBadRequest, "INVALID_GPS", err.Error())
	case errors.Is(err, service.ErrListingContactPhoneTooLong):
		respond.Error(w, http.StatusBadRequest, "CONTACT_PHONE_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrListingContactEmailTooLong):
		respond.Error(w, http.StatusBadRequest, "CONTACT_EMAIL_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrListingPriceTooLarge):
		respond.Error(w, http.StatusBadRequest, "PRICE_TOO_LARGE", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseListingID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// parseListingFilter builds a ListingFilter from the request query string.
func parseListingFilter(r *http.Request) repository.ListingFilter {
	q := r.URL.Query()
	f := repository.ListingFilter{
		Category: q.Get("category"),
		Keyword:  q.Get("keyword"),
		Limit:    20,
		Offset:   0,
	}
	if v := q.Get("price_min"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			f.PriceMin = &n
		}
	}
	if v := q.Get("price_max"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			f.PriceMax = &n
		}
	}
	if v := q.Get("posted_after"); v != "" {
		f.PostedAfter = &v
	}
	if lat, lng, radius, ok := parseNearFilter(q); ok {
		f.NearLat = &lat
		f.NearLng = &lng
		f.RadiusKm = &radius
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			f.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			f.Offset = n
		}
	}
	return f
}

// maxRadiusKm caps the distance filter's radius so a bogus huge value can't be
// used to force what looks like a filtered query into a full table scan.
const maxRadiusKm = 20_000

// parseNearFilter reads near_lat, near_lng, and radius_km from the query string.
// It returns ok=false unless all three are present, parseable, and within bounds —
// a partial or invalid distance filter is treated as no filter at all.
func parseNearFilter(q url.Values) (lat, lng, radius float64, ok bool) {
	latStr, lngStr, radiusStr := q.Get("near_lat"), q.Get("near_lng"), q.Get("radius_km")
	if latStr == "" || lngStr == "" || radiusStr == "" {
		return 0, 0, 0, false
	}
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		return 0, 0, 0, false
	}
	lng, err = strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		return 0, 0, 0, false
	}
	radius, err = strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 || radius > maxRadiusKm {
		return 0, 0, 0, false
	}
	return lat, lng, radius, true
}

// Create handles POST /api/v1/listings — creates a new listing.
func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	var req listingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	listing, err := h.listings.Create(r.Context(), userID, req.toParams())
	if err != nil {
		listingError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, listingJSON(listing))
}

// Search handles GET /api/v1/listings — returns paginated listings matching the query filters.
// With mine=true (auth required) it returns the caller's own listings, including hidden ones.
func (h *ListingHandler) Search(w http.ResponseWriter, r *http.Request) {
	viewerID, _ := middleware.UserIDFromContext(r.Context())

	filter := parseListingFilter(r)
	if r.URL.Query().Get("mine") == "true" {
		if viewerID == 0 {
			respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
			return
		}
		filter.OwnerUserID = &viewerID
	}

	listings, total, err := h.listings.Search(r.Context(), filter)
	if err != nil {
		listingError(w, err)
		return
	}

	items := make([]map[string]any, len(listings))
	for i, l := range listings {
		items[i] = listingJSON(l)
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// Get handles GET /api/v1/listings/{id} — returns a single listing.
func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request) {
	viewerID, _ := middleware.UserIDFromContext(r.Context())

	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	listing, err := h.listings.Get(r.Context(), viewerID, id)
	if err != nil {
		listingError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, listingJSON(listing))
}

// Update handles PATCH /api/v1/listings/{id} — updates a listing owned by the caller.
func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	var req listingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	listing, err := h.listings.Update(r.Context(), userID, id, req.toParams())
	if err != nil {
		listingError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, listingJSON(listing))
}

// Hide handles PATCH /api/v1/listings/{id}/hide — toggles a listing's visibility.
func (h *ListingHandler) Hide(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	var req hideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.listings.SetHidden(r.Context(), userID, id, req.Hidden); err != nil {
		listingError(w, err)
		return
	}

	listing, err := h.listings.Get(r.Context(), userID, id)
	if err != nil {
		listingError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, listingJSON(listing))
}

// Delete handles DELETE /api/v1/listings/{id} — deletes a listing owned by the caller.
func (h *ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	if err := h.listings.Delete(r.Context(), userID, id); err != nil {
		listingError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
