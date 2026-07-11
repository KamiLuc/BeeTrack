package handler

import (
	"context"
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

type HiveHandler struct {
	hives       *service.HiveService
	inspections *service.InspectionService
}

func NewHiveHandler(hives *service.HiveService, inspections *service.InspectionService) *HiveHandler {
	return &HiveHandler{hives: hives, inspections: inspections}
}

func hiveDiseaseJSON(d *model.HiveDisease) map[string]any {
	return map[string]any{
		"id":         d.ID,
		"disease":    d.Disease,
		"created_at": d.CreatedAt,
	}
}

func hiveJSON(hive *model.Hive, diseases []*model.HiveDisease, lastInspectedAt *time.Time) map[string]any {
	dd := make([]map[string]any, len(diseases))
	for i, d := range diseases {
		dd[i] = hiveDiseaseJSON(d)
	}
	return map[string]any{
		"id":                hive.ID,
		"apiary_id":         hive.ApiaryID,
		"name":              hive.Name,
		"type":              hive.Type,
		"active":            hive.Active,
		"frames":            hive.Frames,
		"queenless":         hive.Queenless,
		"ready_for_harvest": hive.ReadyForHarvest,
		"grid_row":          hive.GridRow,
		"grid_col":          hive.GridCol,
		"diseases":          dd,
		"last_inspected_at": lastInspectedAt,
		"created_at":        hive.CreatedAt,
		"updated_at":        hive.UpdatedAt,
	}
}

func (h *HiveHandler) withDiseases(ctx context.Context, hive *model.Hive) (map[string]any, error) {
	diseases, err := h.hives.DiseasesByHive(ctx, hive.ID)
	if err != nil {
		return nil, err
	}
	dates, err := h.inspections.LastInspectionDates(ctx, []int64{hive.ID})
	if err != nil {
		return nil, err
	}
	return hiveJSON(hive, diseases, dates[hive.ID]), nil
}

func hiveError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrHiveNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
	case errors.Is(err, service.ErrHiveDiseaseNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_DISEASE_NOT_FOUND", "hive disease not found")
	case errors.Is(err, service.ErrInvalidDisease):
		respond.Error(w, http.StatusBadRequest, "INVALID_DISEASE", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func (h *HiveHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	hive, err := h.hives.Get(r.Context(), userID, apiaryID, hiveID)
	if err != nil {
		hiveError(w, err)
		return
	}

	body, err := h.withDiseases(r.Context(), hive)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, body)
}

func (h *HiveHandler) Move(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	var req struct {
		GridCol int `json:"grid_col"`
		GridRow int `json:"grid_row"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	hive, err := h.hives.Move(r.Context(), userID, apiaryID, hiveID, req.GridRow, req.GridCol)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveNotFound):
			respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
		case errors.Is(err, service.ErrInvalidGridPosition):
			respond.Error(w, http.StatusBadRequest, "INVALID_GRID_POSITION", err.Error())
		case errors.Is(err, service.ErrPositionOccupied):
			respond.Error(w, http.StatusConflict, "POSITION_OCCUPIED", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	body, err := h.withDiseases(r.Context(), hive)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, body)
}

func (h *HiveHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	var req struct {
		Active          bool   `json:"active"`
		Frames          int    `json:"frames"`
		Name            string `json:"name"`
		Queenless       bool   `json:"queenless"`
		ReadyForHarvest bool   `json:"ready_for_harvest"`
		Type            string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	hive, err := h.hives.Update(r.Context(), userID, apiaryID, hiveID, req.Name, req.Type, req.Active, req.ReadyForHarvest, req.Queenless, req.Frames)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired):
			respond.Error(w, http.StatusBadRequest, "NAME_REQUIRED", err.Error())
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveNotFound):
			respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
		case errors.Is(err, service.ErrDuplicateHiveName):
			respond.Error(w, http.StatusConflict, "DUPLICATE_HIVE_NAME", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	body, err := h.withDiseases(r.Context(), hive)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, body)
}

func (h *HiveHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	if err := h.hives.Delete(r.Context(), userID, apiaryID, hiveID); err != nil {
		hiveError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HiveHandler) List(w http.ResponseWriter, r *http.Request) {
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

	hives, err := h.hives.List(r.Context(), userID, apiaryID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	ids := make([]int64, len(hives))
	for i, h := range hives {
		ids[i] = h.ID
	}
	diseaseMap, err := h.hives.DiseasesForHives(r.Context(), ids)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	lastDates, err := h.inspections.LastInspectionDates(r.Context(), ids)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	items := make([]map[string]any, len(hives))
	for i, hive := range hives {
		items[i] = hiveJSON(hive, diseaseMap[hive.ID], lastDates[hive.ID])
	}
	respond.JSON(w, http.StatusOK, items)
}

func (h *HiveHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Active          *bool  `json:"active"`
		Frames          int    `json:"frames"`
		GridCol         int    `json:"grid_col"`
		GridRow         int    `json:"grid_row"`
		Name            string `json:"name"`
		Queenless       bool   `json:"queenless"`
		ReadyForHarvest bool   `json:"ready_for_harvest"`
		Type            string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	hiveType := req.Type
	if hiveType == "" {
		hiveType = "langstroth"
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	hive, err := h.hives.Add(r.Context(), userID, apiaryID, req.Name, hiveType, active, req.Queenless, req.ReadyForHarvest, req.GridRow, req.GridCol, req.Frames)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired):
			respond.Error(w, http.StatusBadRequest, "NAME_REQUIRED", err.Error())
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrInvalidGridPosition):
			respond.Error(w, http.StatusBadRequest, "INVALID_GRID_POSITION", err.Error())
		case errors.Is(err, service.ErrPositionOccupied):
			respond.Error(w, http.StatusConflict, "POSITION_OCCUPIED", err.Error())
		case errors.Is(err, service.ErrDuplicateHiveName):
			respond.Error(w, http.StatusConflict, "DUPLICATE_HIVE_NAME", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusCreated, hiveJSON(hive, []*model.HiveDisease{}, nil))
}

// AddFrames handles PATCH /api/v1/apiaries/{id}/hives/{hiveId}/frames — atomically adds frames to a hive.
func (h *HiveHandler) AddFrames(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	var req struct {
		Delta int `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.hives.AddFrames(r.Context(), userID, apiaryID, hiveID, req.Delta); err != nil {
		hiveError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangeApiary handles POST /api/v1/apiaries/{id}/hives/{hiveId}/transfer — moves a hive to another apiary.
func (h *HiveHandler) ChangeApiary(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	var req struct {
		TargetApiaryID int64 `json:"target_apiary_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	hive, err := h.hives.ChangeApiary(r.Context(), userID, apiaryID, hiveID, req.TargetApiaryID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveNotFound):
			respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
		case errors.Is(err, service.ErrSameApiary):
			respond.Error(w, http.StatusBadRequest, "SAME_APIARY", err.Error())
		case errors.Is(err, service.ErrTargetApiaryFull):
			respond.Error(w, http.StatusConflict, "TARGET_APIARY_FULL", err.Error())
		case errors.Is(err, service.ErrDuplicateHiveName):
			respond.Error(w, http.StatusConflict, "DUPLICATE_HIVE_NAME", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	body, err := h.withDiseases(r.Context(), hive)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, body)
}

// AddDisease handles POST /api/v1/apiaries/{id}/hives/{hiveId}/diseases — adds a disease to a hive.
func (h *HiveHandler) AddDisease(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	var req struct {
		Disease string `json:"disease"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	d, err := h.hives.AddDisease(r.Context(), userID, apiaryID, hiveID, req.Disease)
	if err != nil {
		hiveError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, hiveDiseaseJSON(d))
}

// RemoveDisease handles DELETE /api/v1/apiaries/{id}/hives/{hiveId}/diseases/{diseaseId} — removes a disease from a hive.
func (h *HiveHandler) RemoveDisease(w http.ResponseWriter, r *http.Request) {
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

	hiveID, err := strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid hive id")
		return
	}

	diseaseID, err := strconv.ParseInt(r.PathValue("diseaseId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid disease id")
		return
	}

	if err := h.hives.RemoveDisease(r.Context(), userID, apiaryID, hiveID, diseaseID); err != nil {
		hiveError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
