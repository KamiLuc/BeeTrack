package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

type HiveHandler struct {
	hives *service.HiveService
}

func NewHiveHandler(hives *service.HiveService) *HiveHandler {
	return &HiveHandler{hives: hives}
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
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveNotFound):
			respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"id":         hive.ID,
		"apiary_id":  hive.ApiaryID,
		"name":       hive.Name,
		"type":       hive.Type,
		"active":     hive.Active,
		"grid_row":   hive.GridRow,
		"grid_col":   hive.GridCol,
		"created_at": hive.CreatedAt,
		"updated_at": hive.UpdatedAt,
	})
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

	respond.JSON(w, http.StatusOK, map[string]any{
		"id":         hive.ID,
		"apiary_id":  hive.ApiaryID,
		"name":       hive.Name,
		"type":       hive.Type,
		"active":     hive.Active,
		"grid_row":   hive.GridRow,
		"grid_col":   hive.GridCol,
		"created_at": hive.CreatedAt,
		"updated_at": hive.UpdatedAt,
	})
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
		Active bool   `json:"active"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	hive, err := h.hives.Update(r.Context(), userID, apiaryID, hiveID, req.Name, req.Type, req.Active)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired):
			respond.Error(w, http.StatusBadRequest, "NAME_REQUIRED", err.Error())
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveNotFound):
			respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"id":         hive.ID,
		"apiary_id":  hive.ApiaryID,
		"name":       hive.Name,
		"type":       hive.Type,
		"active":     hive.Active,
		"grid_row":   hive.GridRow,
		"grid_col":   hive.GridCol,
		"created_at": hive.CreatedAt,
		"updated_at": hive.UpdatedAt,
	})
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
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveNotFound):
			respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
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

	type item struct {
		ID        int64  `json:"id"`
		ApiaryID  int64  `json:"apiary_id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Active    bool   `json:"active"`
		GridRow   int    `json:"grid_row"`
		GridCol   int    `json:"grid_col"`
		CreatedAt any    `json:"created_at"`
		UpdatedAt any    `json:"updated_at"`
	}
	items := make([]item, len(hives))
	for i, hive := range hives {
		items[i] = item{
			ID:        hive.ID,
			ApiaryID:  hive.ApiaryID,
			Name:      hive.Name,
			Type:      hive.Type,
			Active:    hive.Active,
			GridRow:   hive.GridRow,
			GridCol:   hive.GridCol,
			CreatedAt: hive.CreatedAt,
			UpdatedAt: hive.UpdatedAt,
		}
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
		Active  *bool  `json:"active"`
		GridCol int    `json:"grid_col"`
		GridRow int    `json:"grid_row"`
		Name    string `json:"name"`
		Type    string `json:"type"`
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

	hive, err := h.hives.Add(r.Context(), userID, apiaryID, req.Name, hiveType, active, req.GridRow, req.GridCol)
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
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{
		"id":         hive.ID,
		"apiary_id":  hive.ApiaryID,
		"name":       hive.Name,
		"type":       hive.Type,
		"active":     hive.Active,
		"grid_row":   hive.GridRow,
		"grid_col":   hive.GridCol,
		"created_at": hive.CreatedAt,
		"updated_at": hive.UpdatedAt,
	})
}
