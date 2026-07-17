package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// FeedingHandler handles HTTP requests for feeding resources.
type FeedingHandler struct {
	feedings *service.FeedingService
}

// NewFeedingHandler creates a FeedingHandler backed by svc.
func NewFeedingHandler(feedings *service.FeedingService) *FeedingHandler {
	return &FeedingHandler{feedings: feedings}
}

type feedingRequest struct {
	FedAt    *time.Time `json:"fed_at"`
	FeedType string     `json:"feed_type"`
	Amount   string     `json:"amount"`
	Notes    string     `json:"notes"`
}

type bulkFeedingRequest struct {
	feedingRequest
	HiveIDs []int64 `json:"hive_ids"`
}

func (req feedingRequest) toParams() service.FeedingParams {
	var at time.Time
	if req.FedAt != nil {
		at = *req.FedAt
	}
	return service.FeedingParams{
		FedAt:    at,
		FeedType: req.FeedType,
		Amount:   req.Amount,
		Notes:    req.Notes,
	}
}

func feedingJSON(f *model.Feeding) map[string]any {
	return map[string]any{
		"id":          f.ID,
		"hive_id":     f.HiveID,
		"fed_by":      f.FedBy,
		"fed_by_name": f.FedByName,
		"fed_at":      f.FedAt,
		"feed_type":   f.FeedType,
		"amount":      f.Amount,
		"notes":       f.Notes,
		"created_at":  f.CreatedAt,
		"updated_at":  f.UpdatedAt,
	}
}

func feedingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrHiveNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
	case errors.Is(err, service.ErrFeedingNotFound):
		respond.Error(w, http.StatusNotFound, "FEEDING_NOT_FOUND", "feeding not found")
	case errors.Is(err, service.ErrFedAtRequired):
		respond.Error(w, http.StatusBadRequest, "FED_AT_REQUIRED", err.Error())
	case errors.Is(err, service.ErrFeedTypeRequired):
		respond.Error(w, http.StatusBadRequest, "FEED_TYPE_REQUIRED", err.Error())
	case errors.Is(err, service.ErrFeedTypeTooLong):
		respond.Error(w, http.StatusBadRequest, "FEED_TYPE_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrAmountTooLong):
		respond.Error(w, http.StatusBadRequest, "AMOUNT_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrFeedingNotesTooLong):
		respond.Error(w, http.StatusBadRequest, "NOTES_TOO_LONG", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseFeedingPathIDs(r *http.Request) (apiaryID, hiveID int64, err error) {
	apiaryID, err = strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return
	}
	hiveID, err = strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	return
}

// FeedTypes handles GET /api/v1/feed-types — returns feed types this user has previously used.
func (h *FeedingHandler) FeedTypes(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	types, err := h.feedings.FeedTypeSuggestions(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, types)
}

// Amounts handles GET /api/v1/feed-amounts — returns amounts this user has previously used.
func (h *FeedingHandler) Amounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	amounts, err := h.feedings.AmountSuggestions(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, amounts)
}

// BulkCreate handles POST /api/v1/apiaries/{id}/feedings/bulk — creates one feeding per hive in the apiary.
func (h *FeedingHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary id")
		return
	}

	var req bulkFeedingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	count, err := h.feedings.BulkFeed(r.Context(), userID, apiaryID, req.HiveIDs, req.feedingRequest.toParams())
	if err != nil {
		feedingError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{"count": count})
}

// Create handles POST /api/v1/apiaries/{id}/hives/{hiveId}/feedings — creates a new feeding.
func (h *FeedingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseFeedingPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	var req feedingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	f, err := h.feedings.Create(r.Context(), userID, apiaryID, hiveID, req.toParams())
	if err != nil {
		feedingError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, feedingJSON(f))
}

// Get handles GET /api/v1/apiaries/{id}/hives/{hiveId}/feedings/{feedingId} — returns a single feeding.
func (h *FeedingHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseFeedingPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	feedingID, err := strconv.ParseInt(r.PathValue("feedingId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid feeding id")
		return
	}

	f, err := h.feedings.Get(r.Context(), userID, apiaryID, hiveID, feedingID)
	if err != nil {
		feedingError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, feedingJSON(f))
}

// List handles GET /api/v1/apiaries/{id}/hives/{hiveId}/feedings — returns paginated feedings.
func (h *FeedingHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseFeedingPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	limit := 20
	offset := 0
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

	feedings, total, err := h.feedings.List(r.Context(), userID, apiaryID, hiveID, limit, offset)
	if err != nil {
		feedingError(w, err)
		return
	}

	items := make([]map[string]any, len(feedings))
	for i, f := range feedings {
		items[i] = feedingJSON(f)
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// Update handles PATCH /api/v1/apiaries/{id}/hives/{hiveId}/feedings/{feedingId} — updates a feeding.
func (h *FeedingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseFeedingPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	feedingID, err := strconv.ParseInt(r.PathValue("feedingId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid feeding id")
		return
	}

	var req feedingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	f, err := h.feedings.Update(r.Context(), userID, apiaryID, hiveID, feedingID, req.toParams())
	if err != nil {
		feedingError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, feedingJSON(f))
}

// Delete handles DELETE /api/v1/apiaries/{id}/hives/{hiveId}/feedings/{feedingId} — deletes a feeding.
func (h *FeedingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseFeedingPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	feedingID, err := strconv.ParseInt(r.PathValue("feedingId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid feeding id")
		return
	}

	if err := h.feedings.Delete(r.Context(), userID, apiaryID, hiveID, feedingID); err != nil {
		feedingError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
