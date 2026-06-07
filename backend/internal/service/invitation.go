package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
)

var (
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationPending   = errors.New("invitation already pending")
	ErrAlreadyMember       = errors.New("user is already a member of this apiary")
	ErrCannotInviteSelf    = errors.New("cannot invite yourself")
	ErrCannotRemoveOwner   = errors.New("cannot remove the owner")
	ErrInvitationMismatch  = errors.New("invitation does not belong to you")
	ErrUserNotFound        = errors.New("user with that email not found")
)

type InvitationRepo interface {
	Create(ctx context.Context, inv *model.ApiaryInvitation) error
	GetPending(ctx context.Context, apiaryID int64, email string) (*model.ApiaryInvitation, error)
	GetByID(ctx context.Context, id int64) (*model.ApiaryInvitation, error)
	ListByApiary(ctx context.Context, apiaryID int64) ([]model.ApiaryMemberInfo, []model.ApiaryInvitation, error)
	ListPendingByEmail(ctx context.Context, email string) ([]model.MyInvitationView, error)
	CountPendingByEmail(ctx context.Context, email string) (int, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	Delete(ctx context.Context, id int64) error
	AddMember(ctx context.Context, apiaryID, userID int64) error
	IsMember(ctx context.Context, apiaryID int64, email string) (bool, error)
	RemoveMember(ctx context.Context, apiaryID, userID int64) error
}

type UserLookup interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type InvitationService struct {
	apiaries    ApiaryRepository
	invitations InvitationRepo
	users       UserLookup
}

func NewInvitationService(apiaries ApiaryRepository, invitations InvitationRepo, users UserLookup) *InvitationService {
	return &InvitationService{apiaries: apiaries, invitations: invitations, users: users}
}

// Invite sends an invitation for apiaryID to email; only the owner may invite.
func (s *InvitationService) Invite(ctx context.Context, ownerUserID, apiaryID int64, email string) (*model.ApiaryInvitation, error) {
	_, role, err := s.apiaries.GetMembership(ctx, apiaryID, ownerUserID)
	if err != nil {
		return nil, ErrApiaryNotFound
	}
	if role != "owner" {
		return nil, ErrForbidden
	}

	owner, err := s.users.GetByID(ctx, ownerUserID)
	if err != nil || owner == nil {
		return nil, fmt.Errorf("get owner: %w", err)
	}
	if owner.Email == email {
		return nil, ErrCannotInviteSelf
	}

	invitee, err := s.users.GetByEmail(ctx, email)
	if err != nil || invitee == nil {
		return nil, ErrUserNotFound
	}

	isMember, err := s.invitations.IsMember(ctx, apiaryID, email)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if isMember {
		return nil, ErrAlreadyMember
	}

	existing, err := s.invitations.GetPending(ctx, apiaryID, email)
	if err != nil {
		return nil, fmt.Errorf("check pending: %w", err)
	}
	if existing != nil {
		return nil, ErrInvitationPending
	}

	inv := &model.ApiaryInvitation{
		ApiaryID:        apiaryID,
		InvitedByUserID: ownerUserID,
		InvitedEmail:    email,
		Status:          "pending",
	}
	if err := s.invitations.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}
	return inv, nil
}

// ListForApiary returns members and pending invitations for an apiary; owner only.
func (s *InvitationService) ListForApiary(ctx context.Context, userID, apiaryID int64) ([]model.ApiaryMemberInfo, []model.ApiaryInvitation, error) {
	_, role, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		return nil, nil, ErrApiaryNotFound
	}
	if role != "owner" {
		return nil, nil, ErrForbidden
	}

	members, invitations, err := s.invitations.ListByApiary(ctx, apiaryID)
	if err != nil {
		return nil, nil, fmt.Errorf("list apiary members: %w", err)
	}
	return members, invitations, nil
}

// CancelInvitation deletes a pending invitation; owner only.
func (s *InvitationService) CancelInvitation(ctx context.Context, userID, apiaryID, invitationID int64) error {
	_, role, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		return ErrApiaryNotFound
	}
	if role != "owner" {
		return ErrForbidden
	}

	inv, err := s.invitations.GetByID(ctx, invitationID)
	if err != nil || inv == nil || inv.ApiaryID != apiaryID {
		return ErrInvitationNotFound
	}

	return s.invitations.Delete(ctx, invitationID)
}

// RemoveMember removes a non-owner member from the apiary; owner only.
func (s *InvitationService) RemoveMember(ctx context.Context, ownerUserID, apiaryID, memberUserID int64) error {
	_, role, err := s.apiaries.GetMembership(ctx, apiaryID, ownerUserID)
	if err != nil {
		return ErrApiaryNotFound
	}
	if role != "owner" {
		return ErrForbidden
	}
	if ownerUserID == memberUserID {
		return ErrCannotRemoveOwner
	}

	return s.invitations.RemoveMember(ctx, apiaryID, memberUserID)
}

// ListMine returns pending invitations addressed to the authenticated user's email.
func (s *InvitationService) ListMine(ctx context.Context, userID int64) ([]model.MyInvitationView, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	invitations, err := s.invitations.ListPendingByEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	return invitations, nil
}

// CountMine returns the number of pending invitations for the authenticated user.
func (s *InvitationService) CountMine(ctx context.Context, userID int64) (int, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return 0, fmt.Errorf("get user: %w", err)
	}
	return s.invitations.CountPendingByEmail(ctx, user.Email)
}

// Accept accepts a pending invitation and adds the user to the apiary.
func (s *InvitationService) Accept(ctx context.Context, userID, invitationID int64) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("get user: %w", err)
	}

	inv, err := s.invitations.GetByID(ctx, invitationID)
	if err != nil || inv == nil {
		return ErrInvitationNotFound
	}
	if inv.InvitedEmail != user.Email || inv.Status != "pending" {
		return ErrInvitationMismatch
	}

	if err := s.invitations.AddMember(ctx, inv.ApiaryID, userID); err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return s.invitations.UpdateStatus(ctx, invitationID, "accepted")
}

// Leave removes the authenticated member from an apiary; the owner cannot leave.
func (s *InvitationService) Leave(ctx context.Context, userID, apiaryID int64) error {
	_, role, err := s.apiaries.GetMembership(ctx, apiaryID, userID)
	if err != nil {
		return ErrApiaryNotFound
	}
	if role == "owner" {
		return ErrCannotRemoveOwner
	}
	return s.invitations.RemoveMember(ctx, apiaryID, userID)
}

// Decline declines a pending invitation.
func (s *InvitationService) Decline(ctx context.Context, userID, invitationID int64) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("get user: %w", err)
	}

	inv, err := s.invitations.GetByID(ctx, invitationID)
	if err != nil || inv == nil {
		return ErrInvitationNotFound
	}
	if inv.InvitedEmail != user.Email || inv.Status != "pending" {
		return ErrInvitationMismatch
	}

	return s.invitations.UpdateStatus(ctx, invitationID, "declined")
}
