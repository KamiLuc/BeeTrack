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

type ApiaryHandler struct {
	apiary *service.ApiaryService
}

func NewApiaryHandler(apiary *service.ApiaryService) *ApiaryHandler {
	return &ApiaryHandler{apiary: apiary}
}

// Create handles POST /api/v1/apiaries — creates a new apiary owned by the authenticated user.
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

// List handles GET /api/v1/apiaries — returns all apiaries the authenticated user is a member of.
func (h *ApiaryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	memberships, err := h.apiary.List(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	type item struct {
		CreatedAt any      `json:"created_at"`
		GridCols  int      `json:"grid_cols"`
		GridRows  int      `json:"grid_rows"`
		ID        int64    `json:"id"`
		Lat       *float64 `json:"lat"`
		Lng       *float64 `json:"lng"`
		Name      string   `json:"name"`
		UpdatedAt any      `json:"updated_at"`
		UserRole  string   `json:"user_role"`
	}
	items := make([]item, len(memberships))
	for i, m := range memberships {
		items[i] = item{
			CreatedAt: m.Apiary.CreatedAt,
			GridCols:  m.Apiary.GridCols,
			GridRows:  m.Apiary.GridRows,
			ID:        m.Apiary.ID,
			Lat:       m.Apiary.Lat,
			Lng:       m.Apiary.Lng,
			Name:      m.Apiary.Name,
			UpdatedAt: m.Apiary.UpdatedAt,
			UserRole:  m.UserRole,
		}
	}

	respond.JSON(w, http.StatusOK, items)
}

// Update handles PATCH /api/v1/apiaries/{id} — updates an apiary; only the owner may do this.
func (h *ApiaryHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	apiary, err := h.apiary.Update(r.Context(), userID, apiaryID, req.Name, req.Lat, req.Lng, req.GridRows, req.GridCols)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired):
			respond.Error(w, http.StatusBadRequest, "NAME_REQUIRED", err.Error())
		case errors.Is(err, service.ErrInvalidGridSize):
			respond.Error(w, http.StatusBadRequest, "INVALID_GRID_SIZE", err.Error())
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrForbidden):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "only the owner can edit this apiary")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
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

// Delete handles DELETE /api/v1/apiaries/{id} — deletes an apiary; only the owner may do this.
func (h *ApiaryHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.apiary.Delete(r.Context(), userID, apiaryID); err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrForbidden):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "only the owner can delete this apiary")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
