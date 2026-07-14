package handler

import (
	"errors"
	"net/http"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// ListingFavoriteHandler handles HTTP requests for favoriting marketplace listings.
type ListingFavoriteHandler struct {
	favorites *service.ListingFavoriteService
}

// NewListingFavoriteHandler creates a ListingFavoriteHandler backed by svc.
func NewListingFavoriteHandler(favorites *service.ListingFavoriteService) *ListingFavoriteHandler {
	return &ListingFavoriteHandler{favorites: favorites}
}

func listingFavoriteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrListingNotFound):
		respond.Error(w, http.StatusNotFound, "LISTING_NOT_FOUND", "listing not found")
	case errors.Is(err, service.ErrCannotFavoriteOwnListing):
		respond.Error(w, http.StatusForbidden, "CANNOT_FAVORITE_OWN_LISTING", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

// Add handles POST /api/v1/listings/{id}/favorite — saves a listing to the caller's favorites.
func (h *ListingFavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
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

	if err := h.favorites.Add(r.Context(), userID, id); err != nil {
		listingFavoriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Remove handles DELETE /api/v1/listings/{id}/favorite — removes a listing from the caller's favorites.
func (h *ListingFavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
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

	if err := h.favorites.Remove(r.Context(), userID, id); err != nil {
		listingFavoriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Check handles GET /api/v1/listings/{id}/favorite — reports whether the caller has favorited the listing.
func (h *ListingFavoriteHandler) Check(w http.ResponseWriter, r *http.Request) {
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

	isFavorite, err := h.favorites.IsFavorite(r.Context(), userID, id)
	if err != nil {
		listingFavoriteError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{"is_favorite": isFavorite})
}

// List handles GET /api/v1/favorites — returns the caller's favorited listings.
func (h *ListingFavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	listings, err := h.favorites.List(r.Context(), userID)
	if err != nil {
		listingFavoriteError(w, err)
		return
	}

	items := make([]map[string]any, len(listings))
	for i, l := range listings {
		items[i] = listingJSON(l)
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": items, "total": len(listings)})
}
