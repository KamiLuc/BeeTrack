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

type invitationService interface {
	Invite(ctx context.Context, ownerUserID, apiaryID int64, email string) (*model.ApiaryInvitation, error)
	ListForApiary(ctx context.Context, userID, apiaryID int64) ([]model.ApiaryMemberInfo, []model.ApiaryInvitation, error)
	CancelInvitation(ctx context.Context, userID, apiaryID, invitationID int64) error
	RemoveMember(ctx context.Context, ownerUserID, apiaryID, memberUserID int64) error
	Leave(ctx context.Context, userID, apiaryID int64) error
	ListMine(ctx context.Context, userID int64) ([]model.MyInvitationView, error)
	CountMine(ctx context.Context, userID int64) (int, error)
	Accept(ctx context.Context, userID, invitationID int64) error
	Decline(ctx context.Context, userID, invitationID int64) error
}

type InvitationHandler struct {
	svc invitationService
}

func NewInvitationHandler(svc invitationService) *InvitationHandler {
	return &InvitationHandler{svc: svc}
}

// Invite handles POST /api/v1/apiaries/{id}/invitations — sends an invitation; owner only.
func (h *InvitationHandler) Invite(w http.ResponseWriter, r *http.Request) {
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
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "email is required")
		return
	}

	inv, err := h.svc.Invite(r.Context(), userID, apiaryID, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			respond.Error(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		case errors.Is(err, service.ErrEmailTooLong):
			respond.Error(w, http.StatusBadRequest, "EMAIL_TOO_LONG", err.Error())
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrForbidden):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "only the owner can invite members")
		case errors.Is(err, service.ErrCannotInviteSelf):
			respond.Error(w, http.StatusBadRequest, "CANNOT_INVITE_SELF", "cannot invite yourself")
		case errors.Is(err, service.ErrUserNotFound):
			respond.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "no account found for that email address")
		case errors.Is(err, service.ErrAlreadyMember):
			respond.Error(w, http.StatusConflict, "ALREADY_MEMBER", "user is already a member")
		case errors.Is(err, service.ErrInvitationPending):
			respond.Error(w, http.StatusConflict, "INVITATION_PENDING", "invitation already pending for this email")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{
		"id":            inv.ID,
		"apiary_id":     inv.ApiaryID,
		"invited_email": inv.InvitedEmail,
		"status":        inv.Status,
		"created_at":    inv.CreatedAt,
	})
}

// ListForApiary handles GET /api/v1/apiaries/{id}/invitations — lists members and pending invitations; owner only.
func (h *InvitationHandler) ListForApiary(w http.ResponseWriter, r *http.Request) {
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

	members, invitations, err := h.svc.ListForApiary(r.Context(), userID, apiaryID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrForbidden):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "only the owner can view members")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	type memberItem struct {
		Email    string    `json:"email"`
		JoinedAt time.Time `json:"joined_at"`
		Name     string    `json:"name"`
		Role     string    `json:"role"`
		UserID   int64     `json:"user_id"`
	}
	type invitationItem struct {
		ApiaryID     int64     `json:"apiary_id"`
		CreatedAt    time.Time `json:"created_at"`
		ID           int64     `json:"id"`
		InvitedEmail string    `json:"invited_email"`
		Status       string    `json:"status"`
	}

	memberItems := make([]memberItem, len(members))
	for i, m := range members {
		memberItems[i] = memberItem{
			Email:    m.Email,
			JoinedAt: m.JoinedAt,
			Name:     m.Name,
			Role:     m.Role,
			UserID:   m.UserID,
		}
	}
	invItems := make([]invitationItem, len(invitations))
	for i, inv := range invitations {
		invItems[i] = invitationItem{
			ApiaryID:     inv.ApiaryID,
			CreatedAt:    inv.CreatedAt,
			ID:           inv.ID,
			InvitedEmail: inv.InvitedEmail,
			Status:       inv.Status,
		}
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"members":     memberItems,
		"invitations": invItems,
	})
}

// CancelInvitation handles DELETE /api/v1/apiaries/{id}/invitations/{invitationId} — owner cancels a pending invite.
func (h *InvitationHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
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
	invitationID, err := strconv.ParseInt(r.PathValue("invitationId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid invitation id")
		return
	}

	if err := h.svc.CancelInvitation(r.Context(), userID, apiaryID, invitationID); err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrForbidden):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "only the owner can cancel invitations")
		case errors.Is(err, service.ErrInvitationNotFound):
			respond.Error(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "invitation not found")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember handles DELETE /api/v1/apiaries/{id}/members/{userId} — owner removes a member.
func (h *InvitationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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
	memberUserID, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid user id")
		return
	}

	if err := h.svc.RemoveMember(r.Context(), userID, apiaryID, memberUserID); err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrForbidden):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "only the owner can remove members")
		case errors.Is(err, service.ErrCannotRemoveOwner):
			respond.Error(w, http.StatusBadRequest, "CANNOT_REMOVE_OWNER", "cannot remove the owner")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Leave handles DELETE /api/v1/apiaries/{id}/leave — removes the authenticated user from the apiary.
func (h *InvitationHandler) Leave(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.Leave(r.Context(), userID, apiaryID); err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrCannotRemoveOwner):
			respond.Error(w, http.StatusBadRequest, "CANNOT_LEAVE_AS_OWNER", "owner cannot leave the apiary")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListMine handles GET /api/v1/invitations — returns the user's pending invitations.
func (h *InvitationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	invitations, err := h.svc.ListMine(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	type item struct {
		ApiaryID      int64     `json:"apiary_id"`
		ApiaryName    string    `json:"apiary_name"`
		CreatedAt     time.Time `json:"created_at"`
		ID            int64     `json:"id"`
		InvitedByName string    `json:"invited_by_name"`
	}
	items := make([]item, len(invitations))
	for i, v := range invitations {
		items[i] = item{
			ApiaryID:      v.ApiaryID,
			ApiaryName:    v.ApiaryName,
			CreatedAt:     v.CreatedAt,
			ID:            v.ID,
			InvitedByName: v.InvitedByName,
		}
	}
	respond.JSON(w, http.StatusOK, items)
}

// CountMine handles GET /api/v1/invitations/count — returns the count of the user's pending invitations.
func (h *InvitationHandler) CountMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	count, err := h.svc.CountMine(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{"count": count})
}

// Accept handles POST /api/v1/invitations/{id}/accept — accepts a pending invitation.
func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	invitationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid invitation id")
		return
	}

	if err := h.svc.Accept(r.Context(), userID, invitationID); err != nil {
		switch {
		case errors.Is(err, service.ErrInvitationNotFound):
			respond.Error(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "invitation not found")
		case errors.Is(err, service.ErrInvitationMismatch):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "this invitation does not belong to you")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Decline handles POST /api/v1/invitations/{id}/decline — declines a pending invitation.
func (h *InvitationHandler) Decline(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	invitationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid invitation id")
		return
	}

	if err := h.svc.Decline(r.Context(), userID, invitationID); err != nil {
		switch {
		case errors.Is(err, service.ErrInvitationNotFound):
			respond.Error(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "invitation not found")
		case errors.Is(err, service.ErrInvitationMismatch):
			respond.Error(w, http.StatusForbidden, "FORBIDDEN", "this invitation does not belong to you")
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
