package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/validation"
	"gorm.io/gorm"
)

type mockListingModerationStore struct {
	byID           map[int64]*model.Listing
	pending        []*model.Listing
	total          int64
	approvedID     int64
	approvedBy     int64
	rejectedID     int64
	rejectedBy     int64
	rejectedReason string
	removedID      int64
	removedBy      int64
	removedReason  string
	restoredID     int64
	restoredBy     int64
	err            error
}

func (m *mockListingModerationStore) ListReview(ctx context.Context, status, keyword, sortDir string, limit, offset int) ([]*model.Listing, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.pending, m.total, nil
}

func (m *mockListingModerationStore) GetByIDForReview(ctx context.Context, id int64) (*model.Listing, error) {
	if m.err != nil {
		return nil, m.err
	}
	l, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return l, nil
}

func (m *mockListingModerationStore) Approve(ctx context.Context, id, reviewerID int64) error {
	if m.err != nil {
		return m.err
	}
	m.approvedID = id
	m.approvedBy = reviewerID
	return nil
}

func (m *mockListingModerationStore) Reject(ctx context.Context, id, reviewerID int64, reason string) error {
	if m.err != nil {
		return m.err
	}
	m.rejectedID = id
	m.rejectedBy = reviewerID
	m.rejectedReason = reason
	return nil
}

func (m *mockListingModerationStore) Remove(ctx context.Context, id, reviewerID int64, reason string) error {
	if m.err != nil {
		return m.err
	}
	m.removedID = id
	m.removedBy = reviewerID
	m.removedReason = reason
	return nil
}

func (m *mockListingModerationStore) Restore(ctx context.Context, id, reviewerID int64) error {
	if m.err != nil {
		return m.err
	}
	m.restoredID = id
	m.restoredBy = reviewerID
	return nil
}

func TestListingModeration_List(t *testing.T) {
	store := &mockListingModerationStore{
		pending: []*model.Listing{{ID: 1}, {ID: 2}},
		total:   2,
	}
	svc := NewListingModerationService(store)

	items, total, err := svc.List(context.Background(), "pending", "", "asc", 20, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 items, got total=%d len=%d", total, len(items))
	}
}

func TestListingModeration_Get_NotFound(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{}}
	svc := NewListingModerationService(store)

	if _, err := svc.Get(context.Background(), 999); err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingModeration_Approve_Success(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusPending}}}
	svc := NewListingModerationService(store)

	if err := svc.Approve(context.Background(), 7, 1); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if store.approvedID != 1 || store.approvedBy != 7 {
		t.Errorf("expected Approve(1, 7), got Approve(%d, %d)", store.approvedID, store.approvedBy)
	}
}

func TestListingModeration_Approve_NotPending(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusApproved}}}
	svc := NewListingModerationService(store)

	if err := svc.Approve(context.Background(), 7, 1); err != ErrListingNotPending {
		t.Errorf("expected ErrListingNotPending, got %v", err)
	}
}

func TestListingModeration_Approve_NotFound(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{}}
	svc := NewListingModerationService(store)

	if err := svc.Approve(context.Background(), 7, 999); err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingModeration_Reject_RequiresReason(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusPending}}}
	svc := NewListingModerationService(store)

	if err := svc.Reject(context.Background(), 7, 1, ""); err != ErrRejectionReasonRequired {
		t.Errorf("expected ErrRejectionReasonRequired, got %v", err)
	}
	if store.rejectedID != 0 {
		t.Error("expected no reject call to the store")
	}
}

func TestListingModeration_Reject_NotPending(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusRejected}}}
	svc := NewListingModerationService(store)

	if err := svc.Reject(context.Background(), 7, 1, "bad photos"); err != ErrListingNotPending {
		t.Errorf("expected ErrListingNotPending, got %v", err)
	}
}

func TestListingModeration_Reject_Success(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusPending}}}
	svc := NewListingModerationService(store)

	if err := svc.Reject(context.Background(), 7, 1, "bad photos"); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if store.rejectedID != 1 || store.rejectedBy != 7 || store.rejectedReason != "bad photos" {
		t.Errorf("expected Reject(1, 7, %q), got Reject(%d, %d, %q)", "bad photos", store.rejectedID, store.rejectedBy, store.rejectedReason)
	}
}

func TestListingModeration_Reject_WhitespaceOnlyReason(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusPending}}}
	svc := NewListingModerationService(store)

	if err := svc.Reject(context.Background(), 7, 1, "   "); err != ErrRejectionReasonRequired {
		t.Errorf("expected ErrRejectionReasonRequired, got %v", err)
	}
	if store.rejectedID != 0 {
		t.Error("expected no reject call to the store")
	}
}

func TestListingModeration_Reject_ReasonLength(t *testing.T) {
	tests := []struct {
		name    string
		reason  string
		wantErr error
	}{
		{"too short", "ab", ErrRejectionReasonTooShort},
		{"exactly min length", "abc", nil},
		{"too long", strings.Repeat("a", validation.Large.MaxLength()+1), ErrRejectionReasonTooLong},
		{"exactly max length", strings.Repeat("a", validation.Large.MaxLength()), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusPending}}}
			svc := NewListingModerationService(store)

			err := svc.Reject(context.Background(), 7, 1, tt.reason)
			if err != tt.wantErr {
				t.Errorf("Reject() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && store.rejectedID != 0 {
				t.Error("expected no reject call to the store")
			}
		})
	}
}

func TestListingModeration_RepoError(t *testing.T) {
	store := &mockListingModerationStore{err: errors.New("db down")}
	svc := NewListingModerationService(store)

	if _, _, err := svc.List(context.Background(), "pending", "", "asc", 20, 0); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestListingModeration_Remove_Success(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusApproved}}}
	svc := NewListingModerationService(store)

	if err := svc.Remove(context.Background(), 7, 1, "seller requested takedown"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if store.removedID != 1 || store.removedBy != 7 || store.removedReason != "seller requested takedown" {
		t.Errorf("expected Remove(1, 7, %q), got Remove(%d, %d, %q)", "seller requested takedown", store.removedID, store.removedBy, store.removedReason)
	}
}

func TestListingModeration_Remove_NotApproved(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusPending}}}
	svc := NewListingModerationService(store)

	if err := svc.Remove(context.Background(), 7, 1, "seller requested takedown"); err != ErrListingNotApproved {
		t.Errorf("expected ErrListingNotApproved, got %v", err)
	}
}

func TestListingModeration_Remove_RequiresReason(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusApproved}}}
	svc := NewListingModerationService(store)

	if err := svc.Remove(context.Background(), 7, 1, ""); err != ErrRejectionReasonRequired {
		t.Errorf("expected ErrRejectionReasonRequired, got %v", err)
	}
	if store.removedID != 0 {
		t.Error("expected no remove call to the store")
	}
}

func TestListingModeration_Restore_Success(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusRemoved}}}
	svc := NewListingModerationService(store)

	if err := svc.Restore(context.Background(), 7, 1); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if store.restoredID != 1 || store.restoredBy != 7 {
		t.Errorf("expected Restore(1, 7), got Restore(%d, %d)", store.restoredID, store.restoredBy)
	}
}

func TestListingModeration_Restore_NotRemoved(t *testing.T) {
	store := &mockListingModerationStore{byID: map[int64]*model.Listing{1: {ID: 1, Status: model.ListingStatusApproved}}}
	svc := NewListingModerationService(store)

	if err := svc.Restore(context.Background(), 7, 1); err != ErrListingNotRemoved {
		t.Errorf("expected ErrListingNotRemoved, got %v", err)
	}
}
