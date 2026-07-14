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

type mockHarvestRepo struct {
	harvest   *model.Harvest
	harvests  []*model.Harvest
	created   *model.Harvest
	updated   *model.Harvest
	deletedID int64
}

func (m *mockHarvestRepo) Create(ctx context.Context, h *model.Harvest) error {
	h.ID = 1
	m.created = h
	return nil
}

func (m *mockHarvestRepo) Delete(ctx context.Context, harvestID int64) error {
	m.deletedID = harvestID
	return nil
}

func (m *mockHarvestRepo) GetByID(ctx context.Context, harvestID, hiveID int64) (*model.Harvest, error) {
	if m.harvest == nil || m.harvest.ID != harvestID || m.harvest.HiveID != hiveID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.harvest, nil
}

func (m *mockHarvestRepo) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	return int64(len(m.harvests)), nil
}

func (m *mockHarvestRepo) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Harvest, error) {
	return m.harvests, nil
}

func (m *mockHarvestRepo) Update(ctx context.Context, h *model.Harvest) error {
	m.updated = h
	return nil
}

func newHarvestSvc(repo *mockHarvestRepo) *HarvestService {
	return NewHarvestService(
		&mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"},
		&mockInspectionHiveReader{hive: &model.Hive{ID: 10, ApiaryID: 1}},
		repo,
	)
}

func validHarvestParams() HarvestParams {
	return HarvestParams{
		HarvestedAt: time.Now(),
		Frames:      5,
		HalfFrames:  2,
		Kilograms:   12.50,
	}
}

func TestHarvestCreate(t *testing.T) {
	repo := &mockHarvestRepo{}
	svc := newHarvestSvc(repo)

	h, err := svc.Create(context.Background(), 1, 1, 10, validHarvestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Frames != 5 {
		t.Errorf("expected frames 5, got %d", h.Frames)
	}
	if h.HalfFrames != 2 {
		t.Errorf("expected half_frames 2, got %d", h.HalfFrames)
	}
	if h.Kilograms != 12.50 {
		t.Errorf("expected kilograms 12.50, got %f", h.Kilograms)
	}
	if h.HarvestedBy != 1 {
		t.Errorf("expected harvested_by 1, got %d", h.HarvestedBy)
	}
}

func TestHarvestCreate_NotesTooLong(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.Notes = strings.Repeat("a", 5001)
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrHarvestNotesTooLong) {
		t.Errorf("expected ErrHarvestNotesTooLong, got %v", err)
	}
}

func TestHarvestCreate_FramesTooLarge(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.Frames = 100
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrHarvestFramesInvalid) {
		t.Errorf("expected ErrHarvestFramesInvalid, got %v", err)
	}
}

func TestHarvestCreate_FramesAtMax(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.Frames = 99
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHarvestCreate_FramesNegative(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.Frames = -1
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrHarvestFramesInvalid) {
		t.Errorf("expected ErrHarvestFramesInvalid, got %v", err)
	}
}

func TestHarvestCreate_HalfFramesTooLarge(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.HalfFrames = 100
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrHarvestHalfFramesInvalid) {
		t.Errorf("expected ErrHarvestHalfFramesInvalid, got %v", err)
	}
}

func TestHarvestCreate_MissingDate(t *testing.T) {
	repo := &mockHarvestRepo{}
	svc := newHarvestSvc(repo)

	_, err := svc.Create(context.Background(), 1, 1, 10, HarvestParams{})
	if err != ErrHarvestedAtRequired {
		t.Errorf("expected ErrHarvestedAtRequired, got %v", err)
	}
}

func TestHarvestCreate_ZeroFrames(t *testing.T) {
	repo := &mockHarvestRepo{}
	svc := newHarvestSvc(repo)

	_, err := svc.Create(context.Background(), 1, 1, 10, HarvestParams{
		HarvestedAt: time.Now(),
		Frames:      0,
		HalfFrames:  0,
		Kilograms:   1.0,
	})
	if err != ErrHarvestFramesRequired {
		t.Errorf("expected ErrHarvestFramesRequired, got %v", err)
	}
}

func TestHarvestCreate_ZeroKilograms(t *testing.T) {
	repo := &mockHarvestRepo{}
	svc := newHarvestSvc(repo)

	_, err := svc.Create(context.Background(), 1, 1, 10, HarvestParams{
		HarvestedAt: time.Now(),
		Frames:      5,
		HalfFrames:  0,
		Kilograms:   0,
	})
	if err != ErrHarvestKilogramsRequired {
		t.Errorf("expected ErrHarvestKilogramsRequired, got %v", err)
	}
}

func TestHarvestCreate_KilogramsTooLarge(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.Kilograms = 1001
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if !errors.Is(err, ErrHarvestKilogramsTooLarge) {
		t.Errorf("expected ErrHarvestKilogramsTooLarge, got %v", err)
	}
}

func TestHarvestCreate_KilogramsAtMax(t *testing.T) {
	svc := newHarvestSvc(&mockHarvestRepo{})

	params := validHarvestParams()
	params.Kilograms = 1000
	_, err := svc.Create(context.Background(), 1, 1, 10, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHarvestList(t *testing.T) {
	repo := &mockHarvestRepo{
		harvests: []*model.Harvest{
			{ID: 1, HiveID: 10, Kilograms: 10.0},
			{ID: 2, HiveID: 10, Kilograms: 8.5},
		},
	}
	svc := newHarvestSvc(repo)

	harvests, total, err := svc.List(context.Background(), 1, 1, 10, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(harvests) != 2 {
		t.Errorf("expected 2 harvests, got %d", len(harvests))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
}

func TestHarvestUpdate(t *testing.T) {
	existing := &model.Harvest{ID: 1, HiveID: 10, Frames: 3, Kilograms: 9.0}
	repo := &mockHarvestRepo{harvest: existing}
	svc := newHarvestSvc(repo)

	params := HarvestParams{
		HarvestedAt: time.Now(),
		Frames:      7,
		HalfFrames:  1,
		Kilograms:   20.25,
	}
	h, err := svc.Update(context.Background(), 1, 1, 10, 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Frames != 7 {
		t.Errorf("expected frames 7, got %d", h.Frames)
	}
	if h.Kilograms != 20.25 {
		t.Errorf("expected kilograms 20.25, got %f", h.Kilograms)
	}
}

func TestHarvestDelete(t *testing.T) {
	existing := &model.Harvest{ID: 1, HiveID: 10}
	repo := &mockHarvestRepo{harvest: existing}
	svc := newHarvestSvc(repo)

	err := svc.Delete(context.Background(), 1, 1, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedID != 1 {
		t.Errorf("expected deletedID 1, got %d", repo.deletedID)
	}
}

func TestHarvestDelete_NotFound(t *testing.T) {
	repo := &mockHarvestRepo{}
	svc := newHarvestSvc(repo)

	err := svc.Delete(context.Background(), 1, 1, 10, 99)
	if err != ErrHarvestNotFound {
		t.Errorf("expected ErrHarvestNotFound, got %v", err)
	}
}
