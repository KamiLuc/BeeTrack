package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			respond.Error(w, http.StatusConflict, "EMAIL_TAKEN", err.Error())
		case errors.Is(err, service.ErrInvalidEmail):
			respond.Error(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		case errors.Is(err, service.ErrWeakPassword):
			respond.Error(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	accessToken, refreshToken, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPassword):
			respond.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	accessToken, refreshToken, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			respond.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error())
		case errors.Is(err, service.ErrTokenExpired):
			respond.Error(w, http.StatusUnauthorized, "TOKEN_EXPIRED", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
