package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockFeedingRepo struct {
	feeding     *model.Feeding
	feedings    []*model.Feeding
	bulkCreated []*model.Feeding
	created     *model.Feeding
	updated     *model.Feeding
	deletedID   int64
}

func (m *mockFeedingRepo) BulkCreate(ctx context.Context, feedings []*model.Feeding) error {
	for i, f := range feedings {
		f.ID = int64(i + 1)
	}
	m.bulkCreated = feedings
	return nil
}

func (m *mockFeedingRepo) Create(ctx context.Context, f *model.Feeding) error {
	f.ID = 1
	m.created = f
	return nil
}

func (m *mockFeedingRepo) Delete(ctx context.Context, feedingID int64) error {
	m.deletedID = feedingID
	return nil
}

func (m *mockFeedingRepo) GetByID(ctx context.Context, feedingID, hiveID int64) (*model.Feeding, error) {
	if m.feeding == nil || m.feeding.ID != feedingID || m.feeding.HiveID != hiveID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.feeding, nil
}

func (m *mockFeedingRepo) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	return int64(len(m.feedings)), nil
}

func (m *mockFeedingRepo) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Feeding, error) {
	return m.feedings, nil
}

func (m *mockFeedingRepo) Update(ctx context.Context, f *model.Feeding) error {
	m.updated = f
	return nil
}

func (m *mockFeedingRepo) DistinctFeedTypes(ctx context.Context, userID int64) ([]string, error) {
	return nil, nil
}

func (m *mockFeedingRepo) DistinctAmounts(ctx context.Context, userID int64) ([]string, error) {
	return nil, nil
}

func newFeedingSvc(repo *mockFeedingRepo) *FeedingService {
	hive := &model.Hive{ID: 10, ApiaryID: 1}
	return NewFeedingService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: hive},
		&mockBulkHiveReader{hives: []*model.Hive{hive}},
		repo,
	)
}

func validFeedingParams() FeedingParams {
	return FeedingParams{
		FedAt:    time.Now(),
		FeedType: "Sugar syrup (1:1)",
		Amount:   "2L",
		Notes:    "topped up the feeder",
	}
}

func TestFeedingCreate_FeedTypeTooLong(t *testing.T) {
	svc := newFeedingSvc(&mockFeedingRepo{})

	params := validFeedingParams()
	params.FeedType = strings.Repeat("a", 51)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrFeedTypeTooLong) {
		t.Errorf("expected ErrFeedTypeTooLong, got %v", err)
	}
}

func TestFeedingCreate_AmountTooLong(t *testing.T) {
	svc := newFeedingSvc(&mockFeedingRepo{})

	params := validFeedingParams()
	params.Amount = strings.Repeat("1", 21)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrAmountTooLong) {
		t.Errorf("expected ErrAmountTooLong, got %v", err)
	}
}

func TestFeedingCreate_NotesTooLong(t *testing.T) {
	svc := newFeedingSvc(&mockFeedingRepo{})

	params := validFeedingParams()
	params.Notes = strings.Repeat("a", 5001)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrFeedingNotesTooLong) {
		t.Errorf("expected ErrFeedingNotesTooLong, got %v", err)
	}
}

func TestFeedingCreate(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := newFeedingSvc(repo)

	f, err := svc.Create(context.Background(), 1, 1, 10, validFeedingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.FeedType != "Sugar syrup (1:1)" {
		t.Errorf("expected feed_type 'Sugar syrup (1:1)', got %s", f.FeedType)
	}
	if f.Amount != "2L" {
		t.Errorf("expected amount '2L', got %s", f.Amount)
	}
}

func TestFeedingCreate_MissingFedAt(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := newFeedingSvc(repo)

	params := validFeedingParams()
	params.FedAt = time.Time{}
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != ErrFedAtRequired {
		t.Errorf("expected ErrFedAtRequired, got %v", err)
	}
}

func TestFeedingCreate_MissingFeedType(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := newFeedingSvc(repo)

	params := validFeedingParams()
	params.FeedType = ""
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != ErrFeedTypeRequired {
		t.Errorf("expected ErrFeedTypeRequired, got %v", err)
	}
}

func TestFeedingGet_NotFound(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := newFeedingSvc(repo)

	_, err := svc.Get(context.Background(), 1, 1, 10, 99)
	if err != ErrFeedingNotFound {
		t.Errorf("expected ErrFeedingNotFound, got %v", err)
	}
}

func TestFeedingDelete(t *testing.T) {
	repo := &mockFeedingRepo{
		feeding: &model.Feeding{ID: 5, HiveID: 10},
	}
	svc := newFeedingSvc(repo)

	if err := svc.Delete(context.Background(), 1, 1, 10, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedID != 5 {
		t.Errorf("expected deletedID 5, got %d", repo.deletedID)
	}
}

func TestBulkFeed(t *testing.T) {
	hives := []*model.Hive{
		{ID: 10, ApiaryID: 1},
		{ID: 11, ApiaryID: 1},
		{ID: 12, ApiaryID: 1},
	}
	repo := &mockFeedingRepo{}
	svc := NewFeedingService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: hives[0]},
		&mockBulkHiveReader{hives: hives},
		repo,
	)

	count, err := svc.BulkFeed(context.Background(), 1, 1, nil, validFeedingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
	if len(repo.bulkCreated) != 3 {
		t.Errorf("expected 3 bulk created, got %d", len(repo.bulkCreated))
	}
}

func TestBulkFeed_SelectedHives(t *testing.T) {
	hives := []*model.Hive{
		{ID: 10, ApiaryID: 1},
		{ID: 11, ApiaryID: 1},
		{ID: 12, ApiaryID: 1},
	}
	repo := &mockFeedingRepo{}
	svc := NewFeedingService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: hives[0]},
		&mockBulkHiveReader{hives: hives},
		repo,
	)

	count, err := svc.BulkFeed(context.Background(), 1, 1, []int64{10, 12, 999}, validFeedingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	for _, f := range repo.bulkCreated {
		if f.HiveID != 10 && f.HiveID != 12 {
			t.Errorf("expected only hives 10 and 12, got hive_id %d", f.HiveID)
		}
	}
}

func TestBulkFeed_NoHives(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := NewFeedingService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{},
		&mockBulkHiveReader{hives: []*model.Hive{}},
		repo,
	)

	count, err := svc.BulkFeed(context.Background(), 1, 1, nil, validFeedingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestBulkFeed_MissingFeedType(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := newFeedingSvc(repo)

	params := validFeedingParams()
	params.FeedType = ""
	_, err := svc.BulkFeed(context.Background(), 1, 1, nil, params)
	if err != ErrFeedTypeRequired {
		t.Errorf("expected ErrFeedTypeRequired, got %v", err)
	}
}

func TestBulkFeed_ApiaryNotFound(t *testing.T) {
	repo := &mockFeedingRepo{}
	svc := NewFeedingService(
		&mockApiaryRepo{},
		&mockInspectionHiveReader{},
		&mockBulkHiveReader{},
		repo,
	)

	_, err := svc.BulkFeed(context.Background(), 1, 99, nil, validFeedingParams())
	if err != ErrApiaryNotFound {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestFeedingUpdate(t *testing.T) {
	repo := &mockFeedingRepo{
		feeding: &model.Feeding{ID: 5, HiveID: 10},
	}
	svc := newFeedingSvc(repo)

	params := validFeedingParams()
	params.FeedType = "Fondant"
	f, err := svc.Update(context.Background(), 1, 1, 10, 5, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.FeedType != "Fondant" {
		t.Errorf("expected feed_type Fondant, got %s", f.FeedType)
	}
}
