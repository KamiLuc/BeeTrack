package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

type ApiaryHandler struct {
	apiary *service.ApiaryService
}

func NewApiaryHandler(apiary *service.ApiaryService) *ApiaryHandler {
	return &ApiaryHandler{apiary: apiary}
}

func (h *ApiaryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	var req struct {
		GridCols int      `json:"grid_cols"`
		GridRows int      `json:"grid_rows"`
		Lat      *float64 `json:"lat"`
		Lng      *float64 `json:"lng"`
		Name     string   `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	apiary, err := h.apiary.Create(r.Context(), userID, req.Name, req.Lat, req.Lng, req.GridRows, req.GridCols)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired):
			respond.Error(w, http.StatusBadRequest, "NAME_REQUIRED", err.Error())
		case errors.Is(err, service.ErrInvalidGridSize):
			respond.Error(w, http.StatusBadRequest, "INVALID_GRID_SIZE", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{
		"id":         apiary.ID,
		"name":       apiary.Name,
		"lat":        apiary.Lat,
		"lng":        apiary.Lng,
		"grid_rows":  apiary.GridRows,
		"grid_cols":  apiary.GridCols,
		"created_at": apiary.CreatedAt,
		"updated_at": apiary.UpdatedAt,
	})
}
