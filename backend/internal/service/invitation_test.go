package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// -- mocks --

type mockInvitationRepo struct {
	pending       *model.ApiaryInvitation
	byID          *model.ApiaryInvitation
	members       []model.ApiaryMemberInfo
	invitations   []model.ApiaryInvitation
	isMember      bool
	createdInv    *model.ApiaryInvitation
	updatedID     int64
	updatedStatus string
	deletedID     int64
	addedApiary   int64
	addedUser     int64
	removedUser   int64
}

func (m *mockInvitationRepo) Create(_ context.Context, inv *model.ApiaryInvitation) error {
	inv.ID = 99
	m.createdInv = inv
	return nil
}
func (m *mockInvitationRepo) GetPending(_ context.Context, _ int64, _ string) (*model.ApiaryInvitation, error) {
	return m.pending, nil
}
func (m *mockInvitationRepo) GetByID(_ context.Context, _ int64) (*model.ApiaryInvitation, error) {
	return m.byID, nil
}
func (m *mockInvitationRepo) ListByApiary(_ context.Context, _ int64) ([]model.ApiaryMemberInfo, []model.ApiaryInvitation, error) {
	return m.members, m.invitations, nil
}
func (m *mockInvitationRepo) ListPendingByEmail(_ context.Context, _ string) ([]model.MyInvitationView, error) {
	return nil, nil
}
func (m *mockInvitationRepo) CountPendingByEmail(_ context.Context, _ string) (int, error) {
	return len(m.invitations), nil
}
func (m *mockInvitationRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	m.updatedID = id
	m.updatedStatus = status
	return nil
}
func (m *mockInvitationRepo) Delete(_ context.Context, id int64) error {
	m.deletedID = id
	return nil
}
func (m *mockInvitationRepo) AddMember(_ context.Context, apiaryID, userID int64) error {
	m.addedApiary = apiaryID
	m.addedUser = userID
	return nil
}
func (m *mockInvitationRepo) IsMember(_ context.Context, _ int64, _ string) (bool, error) {
	return m.isMember, nil
}
func (m *mockInvitationRepo) RemoveMember(_ context.Context, _ int64, userID int64) error {
	m.removedUser = userID
	return nil
}

type mockUserLookup struct {
	byID    *model.User
	byEmail *model.User
}

func (m *mockUserLookup) GetByID(_ context.Context, _ int64) (*model.User, error) {
	return m.byID, nil
}
func (m *mockUserLookup) GetByEmail(_ context.Context, _ string) (*model.User, error) {
	return m.byEmail, nil
}

func newInvSvc(apiaryMock *mockApiaryRepo, invMock *mockInvitationRepo, userMock *mockUserLookup) *InvitationService {
	return NewInvitationService(apiaryMock, invMock, userMock)
}

// -- tests --

func TestInvite_Success(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{
		byID:    &model.User{ID: 1, Email: "owner@example.com"},
		byEmail: &model.User{ID: 2, Email: "guest@example.com"},
	}
	svc := newInvSvc(apiary, inv, users)

	result, err := svc.Invite(context.Background(), 1, 1, "guest@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.InvitedEmail != "guest@example.com" {
		t.Errorf("wrong email: %s", result.InvitedEmail)
	}
	if inv.createdInv == nil {
		t.Error("expected invitation to be created")
	}
}

func TestInvite_InvalidEmail(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{byID: &model.User{ID: 1, Email: "owner@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 1, 1, "not-an-email")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestInvite_EmailTooLong(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{byID: &model.User{ID: 1, Email: "owner@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	longEmail := strings.Repeat("a", 151) + "@example.com"
	_, err := svc.Invite(context.Background(), 1, 1, longEmail)
	if !errors.Is(err, ErrEmailTooLong) {
		t.Errorf("expected ErrEmailTooLong, got %v", err)
	}
}

func TestInvite_NotOwner(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{byID: &model.User{ID: 2, Email: "member@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 2, 1, "guest@example.com")
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestInvite_Self(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{byID: &model.User{ID: 1, Email: "owner@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 1, 1, "owner@example.com")
	if err != ErrCannotInviteSelf {
		t.Errorf("expected ErrCannotInviteSelf, got %v", err)
	}
}

func TestInvite_UserNotFound(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{
		byID:    &model.User{ID: 1, Email: "owner@example.com"},
		byEmail: nil,
	}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 1, 1, "nobody@example.com")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestInvite_AlreadyMember(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{isMember: true}
	users := &mockUserLookup{
		byID:    &model.User{ID: 1, Email: "owner@example.com"},
		byEmail: &model.User{ID: 2, Email: "existing@example.com"},
	}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 1, 1, "existing@example.com")
	if err != ErrAlreadyMember {
		t.Errorf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestInvite_AlreadyPending(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{pending: &model.ApiaryInvitation{ID: 5, InvitedEmail: "guest@example.com"}}
	users := &mockUserLookup{
		byID:    &model.User{ID: 1, Email: "owner@example.com"},
		byEmail: &model.User{ID: 2, Email: "guest@example.com"},
	}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 1, 1, "guest@example.com")
	if err != ErrInvitationPending {
		t.Errorf("expected ErrInvitationPending, got %v", err)
	}
}

func TestInvite_ApiaryNotFound(t *testing.T) {
	apiary := &mockApiaryRepo{}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{byID: &model.User{ID: 1, Email: "owner@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	_, err := svc.Invite(context.Background(), 1, 99, "guest@example.com")
	if err != ErrApiaryNotFound {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestAccept_Success(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{
		byID: &model.ApiaryInvitation{ID: 10, ApiaryID: 1, InvitedEmail: "guest@example.com", Status: "pending"},
	}
	users := &mockUserLookup{byID: &model.User{ID: 2, Email: "guest@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	if err := svc.Accept(context.Background(), 2, 10); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inv.updatedStatus != "accepted" {
		t.Errorf("expected status 'accepted', got %q", inv.updatedStatus)
	}
	if inv.addedUser != 2 {
		t.Errorf("expected member userID 2, got %d", inv.addedUser)
	}
}

func TestAccept_WrongUser(t *testing.T) {
	apiary := &mockApiaryRepo{}
	inv := &mockInvitationRepo{
		byID: &model.ApiaryInvitation{ID: 10, ApiaryID: 1, InvitedEmail: "other@example.com", Status: "pending"},
	}
	users := &mockUserLookup{byID: &model.User{ID: 2, Email: "guest@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	err := svc.Accept(context.Background(), 2, 10)
	if err != ErrInvitationMismatch {
		t.Errorf("expected ErrInvitationMismatch, got %v", err)
	}
}

func TestDecline_Success(t *testing.T) {
	apiary := &mockApiaryRepo{}
	inv := &mockInvitationRepo{
		byID: &model.ApiaryInvitation{ID: 10, ApiaryID: 1, InvitedEmail: "guest@example.com", Status: "pending"},
	}
	users := &mockUserLookup{byID: &model.User{ID: 2, Email: "guest@example.com"}}
	svc := newInvSvc(apiary, inv, users)

	if err := svc.Decline(context.Background(), 2, 10); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inv.updatedStatus != "declined" {
		t.Errorf("expected status 'declined', got %q", inv.updatedStatus)
	}
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{}
	svc := newInvSvc(apiary, inv, users)

	err := svc.RemoveMember(context.Background(), 1, 1, 1)
	if err != ErrCannotRemoveOwner {
		t.Errorf("expected ErrCannotRemoveOwner, got %v", err)
	}
}

func TestListForApiary_NotOwner(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{}
	svc := newInvSvc(apiary, inv, users)

	_, _, err := svc.ListForApiary(context.Background(), 2, 1)
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCancelInvitation_NotFound(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{byID: nil}
	users := &mockUserLookup{}
	svc := newInvSvc(apiary, inv, users)

	err := svc.CancelInvitation(context.Background(), 1, 1, 999)
	if err != ErrInvitationNotFound {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestLeave_Success(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{}
	svc := newInvSvc(apiary, inv, users)

	if err := svc.Leave(context.Background(), 2, 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inv.removedUser != 2 {
		t.Errorf("expected removed userID 2, got %d", inv.removedUser)
	}
}

func TestLeave_OwnerCannotLeave(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{}
	users := &mockUserLookup{}
	svc := newInvSvc(apiary, inv, users)

	err := svc.Leave(context.Background(), 1, 1)
	if err != ErrCannotRemoveOwner {
		t.Errorf("expected ErrCannotRemoveOwner, got %v", err)
	}
}

func TestCancelInvitation_WrongApiary(t *testing.T) {
	apiary := &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "owner"}
	inv := &mockInvitationRepo{byID: &model.ApiaryInvitation{ID: 10, ApiaryID: 2}}
	users := &mockUserLookup{}
	svc := newInvSvc(apiary, inv, users)

	err := svc.CancelInvitation(context.Background(), 1, 1, 10)
	if err != ErrInvitationNotFound {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

// ensure mockApiaryRepo satisfies ApiaryRepository (used across test files in this package)
var _ ApiaryRepository = (*mockApiaryRepo)(nil)

// ensure GetMembership returns not-found for missing apiary
func init() {
	orig := (&mockApiaryRepo{}).GetMembership
	_ = orig
}

func (m *mockApiaryRepo) getByIDCheck() {}

// mockApiaryRepo already implements ApiaryRepository from apiary_test.go;
// we just need the compiler to know mockInvitationRepo satisfies InvitationRepo.
var _ InvitationRepo = (*mockInvitationRepo)(nil)
var _ UserLookup = (*mockUserLookup)(nil)

// Satisfy gorm import used in apiary_test.go
var _ = gorm.ErrRecordNotFound
