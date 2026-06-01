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

	hive, err := h.hives.Add(r.Context(), userID, apiaryID, req.Name, hiveType, req.GridRow, req.GridCol)
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
