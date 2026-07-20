package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// AdminListingHandler handles HTTP requests for the admin listing review queue.
type AdminListingHandler struct {
	moderation *service.ListingModerationService
}

func NewAdminListingHandler(moderation *service.ListingModerationService) *AdminListingHandler {
	return &AdminListingHandler{moderation: moderation}
}

func adminListingJSON(l *model.Listing) map[string]any {
	return map[string]any{
		"id":                l.ID,
		"user_id":           l.UserID,
		"title":             l.Title,
		"description":       l.Description,
		"category":          l.Category,
		"price":             l.Price,
		"quantity":          l.Quantity,
		"address":           l.Address,
		"lat":               l.Lat,
		"lng":               l.Lng,
		"contact_phone":     l.ContactPhone,
		"contact_email":     l.ContactEmail,
		"status":            l.Status,
		"rejection_reason":  l.RejectionReason,
		"is_edit":           l.IsEdit(),
		"created_at":        l.CreatedAt,
		"updated_at":        l.UpdatedAt,
		"images":            adminListingImagesJSON(l),
	}
}

func adminListingImagesJSON(l *model.Listing) []map[string]any {
	images := make([]map[string]any, len(l.Images))
	for i, img := range l.Images {
		images[i] = listingImageJSON(img)
	}
	return images
}

func adminListingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrListingNotFound):
		respond.Error(w, http.StatusNotFound, "LISTING_NOT_FOUND", "listing not found")
	case errors.Is(err, service.ErrListingNotPending):
		respond.Error(w, http.StatusConflict, "LISTING_NOT_PENDING", err.Error())
	case errors.Is(err, service.ErrRejectionReasonRequired):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_REQUIRED", err.Error())
	case errors.Is(err, service.ErrRejectionReasonTooShort):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_TOO_SHORT", err.Error())
	case errors.Is(err, service.ErrRejectionReasonTooLong):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_TOO_LONG", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit, offset = 20, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

// ListPending handles GET /api/v1/admin/listings — returns pending listings (new and edited).
func (h *AdminListingHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	items, total, err := h.moderation.ListPending(r.Context(), limit, offset)
	if err != nil {
		adminListingError(w, err)
		return
	}

	out := make([]map[string]any, len(items))
	for i, l := range items {
		out[i] = adminListingJSON(l)
	}
	respond.JSON(w, http.StatusOK, map[string]any{"items": out, "total": total})
}

// Get handles GET /api/v1/admin/listings/{id} — returns a single listing regardless of status.
func (h *AdminListingHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	l, err := h.moderation.Get(r.Context(), id)
	if err != nil {
		adminListingError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, adminListingJSON(l))
}

// Approve handles POST /api/v1/admin/listings/{id}/approve.
func (h *AdminListingHandler) Approve(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	if err := h.moderation.Approve(r.Context(), adminID, id); err != nil {
		adminListingError(w, err)
		return
	}

	l, err := h.moderation.Get(r.Context(), id)
	if err != nil {
		adminListingError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, adminListingJSON(l))
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

// Reject handles POST /api/v1/admin/listings/{id}/reject.
func (h *AdminListingHandler) Reject(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	var req rejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.moderation.Reject(r.Context(), adminID, id, req.Reason); err != nil {
		adminListingError(w, err)
		return
	}

	l, err := h.moderation.Get(r.Context(), id)
	if err != nil {
		adminListingError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, adminListingJSON(l))
}
