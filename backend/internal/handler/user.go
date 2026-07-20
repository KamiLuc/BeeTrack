package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// Me handles GET /api/v1/users/me — returns the caller's own profile,
// including role (client-side UX only, e.g. gating the admin panel's login
// screen — every actual admin route is still enforced server-side by
// RequireAdmin).
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	user, err := h.users.Me(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if user == nil {
		respond.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"id":       user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
		"verified": user.Verified,
	})
}

func (h *UserHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.users.UpdateName(r.Context(), userID, req.Name); err != nil {
		switch {
		case errors.Is(err, service.ErrNameRequired):
			respond.Error(w, http.StatusBadRequest, "NAME_REQUIRED", err.Error())
		case errors.Is(err, service.ErrNameTooLong):
			respond.Error(w, http.StatusBadRequest, "NAME_TOO_LONG", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
