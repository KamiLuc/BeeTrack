package repository

import (
	"context"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

// Create inserts a new pending invitation.
func (r *InvitationRepository) Create(ctx context.Context, inv *model.ApiaryInvitation) error {
	return r.db.WithContext(ctx).
		Raw(`INSERT INTO apiary_invitations (apiary_id, invited_by_user_id, invited_email)
		     VALUES (?, ?, ?)
		     RETURNING id, status, created_at`,
			inv.ApiaryID, inv.InvitedByUserID, inv.InvitedEmail).
		Scan(inv).Error
}

// GetPending returns a pending invitation for the given apiary+email, or nil if none.
func (r *InvitationRepository) GetPending(ctx context.Context, apiaryID int64, email string) (*model.ApiaryInvitation, error) {
	var inv model.ApiaryInvitation
	err := r.db.WithContext(ctx).
		Raw(`SELECT id, apiary_id, invited_by_user_id, invited_email, status, created_at
		     FROM apiary_invitations
		     WHERE apiary_id = ? AND invited_email = ? AND status = 'pending'`,
			apiaryID, email).
		Scan(&inv).Error
	if err != nil {
		return nil, err
	}
	if inv.ID == 0 {
		return nil, nil
	}
	return &inv, nil
}

// GetByID returns a single invitation by ID.
func (r *InvitationRepository) GetByID(ctx context.Context, id int64) (*model.ApiaryInvitation, error) {
	var inv model.ApiaryInvitation
	err := r.db.WithContext(ctx).
		Raw(`SELECT id, apiary_id, invited_by_user_id, invited_email, status, created_at
		     FROM apiary_invitations WHERE id = ?`, id).
		Scan(&inv).Error
	if err != nil {
		return nil, err
	}
	if inv.ID == 0 {
		return nil, nil
	}
	return &inv, nil
}

// ListByApiary returns all pending invitations for an apiary plus all non-owner members with user info.
func (r *InvitationRepository) ListByApiary(ctx context.Context, apiaryID int64) ([]model.ApiaryMemberInfo, []model.ApiaryInvitation, error) {
	var members []model.ApiaryMemberInfo
	err := r.db.WithContext(ctx).
		Raw(`SELECT u.id AS user_id, u.name, u.email, am.role, am.joined_at
		     FROM apiary_members am
		     JOIN users u ON u.id = am.user_id
		     WHERE am.apiary_id = ? AND am.role != 'owner'
		     ORDER BY am.joined_at ASC`, apiaryID).
		Scan(&members).Error
	if err != nil {
		return nil, nil, err
	}

	var invitations []model.ApiaryInvitation
	err = r.db.WithContext(ctx).
		Raw(`SELECT id, apiary_id, invited_by_user_id, invited_email, status, created_at
		     FROM apiary_invitations
		     WHERE apiary_id = ? AND status = 'pending'
		     ORDER BY created_at ASC`, apiaryID).
		Scan(&invitations).Error
	if err != nil {
		return nil, nil, err
	}

	return members, invitations, nil
}

// ListPendingByEmail returns pending invitations for a user's email, enriched with apiary and inviter names.
func (r *InvitationRepository) ListPendingByEmail(ctx context.Context, email string) ([]model.MyInvitationView, error) {
	var views []model.MyInvitationView
	err := r.db.WithContext(ctx).
		Raw(`SELECT ai.id, ai.apiary_id, a.name AS apiary_name,
		            u.name AS invited_by_name, ai.created_at
		     FROM apiary_invitations ai
		     JOIN apiaries a ON a.id = ai.apiary_id
		     JOIN users u ON u.id = ai.invited_by_user_id
		     WHERE ai.invited_email = ? AND ai.status = 'pending'
		     ORDER BY ai.created_at DESC`, email).
		Scan(&views).Error
	if err != nil {
		return nil, err
	}
	return views, nil
}

// CountPendingByEmail returns the count of pending invitations for a user's email.
func (r *InvitationRepository) CountPendingByEmail(ctx context.Context, email string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Raw(`SELECT COUNT(*) FROM apiary_invitations WHERE invited_email = ? AND status = 'pending'`, email).
		Scan(&count).Error
	return int(count), err
}

// UpdateStatus changes the status of an invitation.
func (r *InvitationRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).
		Exec(`UPDATE apiary_invitations SET status = ? WHERE id = ?`, status, id).Error
}

// Delete removes an invitation record entirely.
func (r *InvitationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Exec(`DELETE FROM apiary_invitations WHERE id = ?`, id).Error
}

// AddMember inserts a user into apiary_members with role 'member'.
func (r *InvitationRepository) AddMember(ctx context.Context, apiaryID, userID int64) error {
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO apiary_members (apiary_id, user_id, role) VALUES (?, ?, 'member')
		      ON CONFLICT (apiary_id, user_id) DO NOTHING`, apiaryID, userID).Error
}

// IsMember reports whether a user is already a member of the apiary.
func (r *InvitationRepository) IsMember(ctx context.Context, apiaryID int64, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Raw(`SELECT COUNT(*) FROM apiary_members am
		     JOIN users u ON u.id = am.user_id
		     WHERE am.apiary_id = ? AND u.email = ?`, apiaryID, email).
		Scan(&count).Error
	return count > 0, err
}

// RemoveMember deletes a non-owner member from an apiary.
func (r *InvitationRepository) RemoveMember(ctx context.Context, apiaryID, userID int64) error {
	return r.db.WithContext(ctx).
		Exec(`DELETE FROM apiary_members WHERE apiary_id = ? AND user_id = ? AND role != 'owner'`,
			apiaryID, userID).Error
}
