package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	var certRequestStatus any
	if l.CertificationRequestID != nil {
		certRequestStatus = l.CertificationRequestStatus
	}
	return map[string]any{
		"id":                           l.ID,
		"user_id":                      l.UserID,
		"owner_email":                  l.OwnerEmail,
		"title":                        l.Title,
		"description":                  l.Description,
		"category":                     l.Category,
		"price":                        l.Price,
		"quantity":                     l.Quantity,
		"address":                      l.Address,
		"lat":                          l.Lat,
		"lng":                          l.Lng,
		"contact_phone":                l.ContactPhone,
		"contact_email":                l.ContactEmail,
		"status":                       l.Status,
		"rejection_reason":             l.RejectionReason,
		"is_edit":                      l.IsEdit(),
		"created_at":                   l.CreatedAt,
		"updated_at":                   l.UpdatedAt,
		"images":                       adminListingImagesJSON(l),
		"certification_request_id":     l.CertificationRequestID,
		"certification_request_status": certRequestStatus,
	}
}

// allowedListingStatuses are the values the admin panel's status filter accepts;
// empty means no filter (all statuses).
var allowedListingStatuses = map[string]bool{
	"":                          true,
	model.ListingStatusPending:  true,
	model.ListingStatusApproved: true,
	model.ListingStatusRejected: true,
	model.ListingStatusRemoved:  true,
}

// parseListingReviewQuery reads the "status", "q", and "sort" query params for the
// admin listing queue, defaulting to no status/keyword filter and ascending
// (oldest-first) order.
func parseListingReviewQuery(r *http.Request) (status, keyword, sortDir string, ok bool) {
	status = r.URL.Query().Get("status")
	if !allowedListingStatuses[status] {
		return "", "", "", false
	}
	keyword = strings.TrimSpace(r.URL.Query().Get("q"))
	sortDir = r.URL.Query().Get("sort")
	if sortDir != "desc" {
		sortDir = "asc"
	}
	return status, keyword, sortDir, true
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
	case errors.Is(err, service.ErrListingNotApproved):
		respond.Error(w, http.StatusConflict, "LISTING_NOT_APPROVED", err.Error())
	case errors.Is(err, service.ErrListingNotRemoved):
		respond.Error(w, http.StatusConflict, "LISTING_NOT_REMOVED", err.Error())
	case errors.Is(err, service.ErrListingPhotoRequired):
		respond.Error(w, http.StatusBadRequest, "PHOTO_REQUIRED", err.Error())
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

// List handles GET /api/v1/admin/listings — returns listings for the review queue,
// optionally filtered by ?status= (pending/approved/rejected/removed, default all),
// by ?q= (keyword matched against title and owner email), and ordered by ?sort=
// (asc/desc, default asc).
func (h *AdminListingHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	status, keyword, sortDir, ok := parseListingReviewQuery(r)
	if !ok {
		respond.Error(w, http.StatusBadRequest, "INVALID_STATUS", "invalid status filter")
		return
	}

	items, total, err := h.moderation.List(r.Context(), status, keyword, sortDir, limit, offset)
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

type removeRequest struct {
	Reason string `json:"reason"`
}

// Remove handles POST /api/v1/admin/listings/{id}/remove — takes a live listing off the marketplace.
func (h *AdminListingHandler) Remove(w http.ResponseWriter, r *http.Request) {
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

	var req removeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.moderation.Remove(r.Context(), adminID, id, req.Reason); err != nil {
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

// Restore handles POST /api/v1/admin/listings/{id}/restore — brings a removed listing back as approved.
func (h *AdminListingHandler) Restore(w http.ResponseWriter, r *http.Request) {
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

	if err := h.moderation.Restore(r.Context(), adminID, id); err != nil {
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
