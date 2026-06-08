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

// HarvestHandler handles HTTP requests for harvest resources.
type HarvestHandler struct {
	harvests *service.HarvestService
}

// NewHarvestHandler creates a HarvestHandler backed by svc.
func NewHarvestHandler(harvests *service.HarvestService) *HarvestHandler {
	return &HarvestHandler{harvests: harvests}
}

type harvestRequest struct {
	HarvestedAt *time.Time `json:"harvested_at"`
	Frames      int        `json:"frames"`
	HalfFrames  int        `json:"half_frames"`
	Kilograms   float64    `json:"kilograms"`
	Notes       string     `json:"notes"`
}

func (req harvestRequest) toParams() service.HarvestParams {
	var at time.Time
	if req.HarvestedAt != nil {
		at = *req.HarvestedAt
	}
	return service.HarvestParams{
		HarvestedAt: at,
		Frames:      req.Frames,
		HalfFrames:  req.HalfFrames,
		Kilograms:   req.Kilograms,
		Notes:       req.Notes,
	}
}

func harvestJSON(h *model.Harvest) map[string]any {
	var harvestedByName any
	if h.HarvestedByName != "" {
		harvestedByName = h.HarvestedByName
	}
	return map[string]any{
		"id":                h.ID,
		"hive_id":           h.HiveID,
		"harvested_by":      h.HarvestedBy,
		"harvested_by_name": harvestedByName,
		"harvested_at":      h.HarvestedAt,
		"frames":            h.Frames,
		"half_frames":       h.HalfFrames,
		"kilograms":         h.Kilograms,
		"notes":             h.Notes,
		"created_at":        h.CreatedAt,
		"updated_at":        h.UpdatedAt,
	}
}

func harvestError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrHiveNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
	case errors.Is(err, service.ErrHarvestNotFound):
		respond.Error(w, http.StatusNotFound, "HARVEST_NOT_FOUND", "harvest not found")
	case errors.Is(err, service.ErrHarvestedAtRequired):
		respond.Error(w, http.StatusBadRequest, "HARVESTED_AT_REQUIRED", err.Error())
	case errors.Is(err, service.ErrHarvestFramesRequired):
		respond.Error(w, http.StatusBadRequest, "HARVEST_FRAMES_REQUIRED", err.Error())
	case errors.Is(err, service.ErrHarvestKilogramsRequired):
		respond.Error(w, http.StatusBadRequest, "HARVEST_KILOGRAMS_REQUIRED", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseHarvestPathIDs(r *http.Request) (apiaryID, hiveID int64, err error) {
	apiaryID, err = strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return
	}
	hiveID, err = strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	return
}

// Create handles POST /api/v1/apiaries/{id}/hives/{hiveId}/harvests — creates a new harvest.
func (h *HarvestHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseHarvestPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	var req harvestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	harvest, err := h.harvests.Create(r.Context(), userID, apiaryID, hiveID, req.toParams())
	if err != nil {
		harvestError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, harvestJSON(harvest))
}

// Get handles GET /api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId} — returns a single harvest.
func (h *HarvestHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseHarvestPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	harvestID, err := strconv.ParseInt(r.PathValue("harvestId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid harvest id")
		return
	}

	harvest, err := h.harvests.Get(r.Context(), userID, apiaryID, hiveID, harvestID)
	if err != nil {
		harvestError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, harvestJSON(harvest))
}

// List handles GET /api/v1/apiaries/{id}/hives/{hiveId}/harvests — returns paginated harvests.
func (h *HarvestHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseHarvestPathIDs(r)
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

	harvests, total, err := h.harvests.List(r.Context(), userID, apiaryID, hiveID, limit, offset)
	if err != nil {
		harvestError(w, err)
		return
	}

	items := make([]map[string]any, len(harvests))
	for i, hv := range harvests {
		items[i] = harvestJSON(hv)
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// Update handles PATCH /api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId} — updates a harvest.
func (h *HarvestHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseHarvestPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	harvestID, err := strconv.ParseInt(r.PathValue("harvestId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid harvest id")
		return
	}

	var req harvestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	harvest, err := h.harvests.Update(r.Context(), userID, apiaryID, hiveID, harvestID, req.toParams())
	if err != nil {
		harvestError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, harvestJSON(harvest))
}

// Delete handles DELETE /api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId} — deletes a harvest.
func (h *HarvestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseHarvestPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	harvestID, err := strconv.ParseInt(r.PathValue("harvestId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid harvest id")
		return
	}

	if err := h.harvests.Delete(r.Context(), userID, apiaryID, hiveID, harvestID); err != nil {
		harvestError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
