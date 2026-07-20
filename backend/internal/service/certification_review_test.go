package service

import (
	"context"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockCertificationRequestStore struct {
	byID           map[int64]*model.HoneyBatchCertificationRequest
	pending        []*model.HoneyBatchCertificationRequest
	total          int64
	approvedID     int64
	approvedBy     int64
	approveErr     error
	rejectedID     int64
	rejectedBy     int64
	rejectedReason string
	rejectErr      error
	err            error
}

func (m *mockCertificationRequestStore) ListPending(ctx context.Context, limit, offset int) ([]*model.HoneyBatchCertificationRequest, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.pending, m.total, nil
}

func (m *mockCertificationRequestStore) GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertificationRequest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byID[id], nil
}

func (m *mockCertificationRequestStore) Approve(ctx context.Context, id, reviewerID int64) (*model.BlockchainJob, error) {
	if m.approveErr != nil {
		return nil, m.approveErr
	}
	m.approvedID = id
	m.approvedBy = reviewerID
	return &model.BlockchainJob{ID: 1, BatchID: 5, Status: model.CertificationStatusQueued}, nil
}

func (m *mockCertificationRequestStore) Reject(ctx context.Context, id, reviewerID int64, reason string) error {
	if m.rejectErr != nil {
		return m.rejectErr
	}
	m.rejectedID = id
	m.rejectedBy = reviewerID
	m.rejectedReason = reason
	return nil
}

func TestCertificationReview_ListPending(t *testing.T) {
	store := &mockCertificationRequestStore{
		pending: []*model.HoneyBatchCertificationRequest{{ID: 1}, {ID: 2}},
		total:   2,
	}
	svc := NewCertificationReviewService(store)

	items, total, err := svc.ListPending(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("ListPending() error = %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 items, got total=%d len=%d", total, len(items))
	}
}

func TestCertificationReview_Get_NotFound(t *testing.T) {
	store := &mockCertificationRequestStore{byID: map[int64]*model.HoneyBatchCertificationRequest{}}
	svc := NewCertificationReviewService(store)

	if _, err := svc.Get(context.Background(), 999); err != ErrCertificationRequestNotFound {
		t.Errorf("expected ErrCertificationRequestNotFound, got %v", err)
	}
}

func TestCertificationReview_Get_Found(t *testing.T) {
	req := &model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}
	store := &mockCertificationRequestStore{byID: map[int64]*model.HoneyBatchCertificationRequest{1: req}}
	svc := NewCertificationReviewService(store)

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != req {
		t.Error("expected the matching request to be returned")
	}
}

func TestCertificationReview_Approve_Success(t *testing.T) {
	store := &mockCertificationRequestStore{}
	svc := NewCertificationReviewService(store)

	if err := svc.Approve(context.Background(), 7, 1); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if store.approvedID != 1 || store.approvedBy != 7 {
		t.Errorf("expected Approve(1, 7), got Approve(%d, %d)", store.approvedID, store.approvedBy)
	}
}

func TestCertificationReview_Approve_NotPending(t *testing.T) {
	store := &mockCertificationRequestStore{approveErr: gorm.ErrRecordNotFound}
	svc := NewCertificationReviewService(store)

	if err := svc.Approve(context.Background(), 7, 1); err != ErrCertificationRequestNotPending {
		t.Errorf("expected ErrCertificationRequestNotPending, got %v", err)
	}
}

func TestCertificationReview_Approve_RepoError(t *testing.T) {
	store := &mockCertificationRequestStore{approveErr: errors.New("db down")}
	svc := NewCertificationReviewService(store)

	if err := svc.Approve(context.Background(), 7, 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCertificationReview_Reject_RequiresReason(t *testing.T) {
	store := &mockCertificationRequestStore{}
	svc := NewCertificationReviewService(store)

	if err := svc.Reject(context.Background(), 7, 1, ""); err != ErrRejectionReasonRequired {
		t.Errorf("expected ErrRejectionReasonRequired, got %v", err)
	}
	if store.rejectedID != 0 {
		t.Error("expected no reject call to the store")
	}
}

func TestCertificationReview_Reject_NotPending(t *testing.T) {
	store := &mockCertificationRequestStore{rejectErr: gorm.ErrRecordNotFound}
	svc := NewCertificationReviewService(store)

	if err := svc.Reject(context.Background(), 7, 1, "bad pdf"); err != ErrCertificationRequestNotPending {
		t.Errorf("expected ErrCertificationRequestNotPending, got %v", err)
	}
}

func TestCertificationReview_Reject_Success(t *testing.T) {
	store := &mockCertificationRequestStore{}
	svc := NewCertificationReviewService(store)

	if err := svc.Reject(context.Background(), 7, 1, "bad pdf"); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if store.rejectedID != 1 || store.rejectedBy != 7 || store.rejectedReason != "bad pdf" {
		t.Errorf("expected Reject(1, 7, %q), got Reject(%d, %d, %q)", "bad pdf", store.rejectedID, store.rejectedBy, store.rejectedReason)
	}
}
