package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/token"
	"gorm.io/gorm"
)

const testListingAuthSecret = "test-listing-secret"

// captureListingStore is a minimal service.ListingStore that records the last
// filter it was searched with, for asserting on Search's viewer-based filtering.
type captureListingStore struct {
	lastFilter repository.ListingFilter
}

func (m *captureListingStore) Create(ctx context.Context, l *model.Listing) error { return nil }

func (m *captureListingStore) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *captureListingStore) Count(ctx context.Context, f repository.ListingFilter) (int64, error) {
	m.lastFilter = f
	return 0, nil
}

func (m *captureListingStore) List(ctx context.Context, f repository.ListingFilter) ([]*model.Listing, error) {
	m.lastFilter = f
	return nil, nil
}

func (m *captureListingStore) Update(ctx context.Context, l *model.Listing) error { return nil }

func (m *captureListingStore) SetHidden(ctx context.Context, id int64, hidden bool) error {
	return nil
}

func (m *captureListingStore) AddImages(ctx context.Context, images []model.ListingImage) error {
	return nil
}

func (m *captureListingStore) DeleteImages(ctx context.Context, listingID int64) error { return nil }

func (m *captureListingStore) Delete(ctx context.Context, id int64) error { return nil }

func TestListingJSON_IncludesApiaryFields(t *testing.T) {
	lat, lng := 52.2297, 21.0122
	l := &model.Listing{
		ID:              5,
		ApiaryID:        ptrInt64(1),
		ApiaryName:      "Home apiary",
		ApiaryLat:       &lat,
		ApiaryLng:       &lng,
		ApiaryHiveCount: 3,
	}

	got := listingJSON(l)

	if got["apiary_lat"] != &lat {
		t.Errorf("expected apiary_lat %v, got %v", &lat, got["apiary_lat"])
	}
	if got["apiary_lng"] != &lng {
		t.Errorf("expected apiary_lng %v, got %v", &lng, got["apiary_lng"])
	}
	if got["apiary_hive_count"] != 3 {
		t.Errorf("expected apiary_hive_count 3, got %v", got["apiary_hive_count"])
	}
}

func TestListingJSON_NilApiaryGPS(t *testing.T) {
	l := &model.Listing{ID: 5}

	got := listingJSON(l)

	if got["apiary_lat"] != (*float64)(nil) {
		t.Errorf("expected apiary_lat nil, got %v", got["apiary_lat"])
	}
	if got["apiary_lng"] != (*float64)(nil) {
		t.Errorf("expected apiary_lng nil, got %v", got["apiary_lng"])
	}
	if got["apiary_hive_count"] != 0 {
		t.Errorf("expected apiary_hive_count 0, got %v", got["apiary_hive_count"])
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestParseListingFilter_HasApiary(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "true sets HasApiary", url: "/listings?has_apiary=true", want: true},
		{name: "absent leaves HasApiary false", url: "/listings", want: false},
		{name: "other value leaves HasApiary false", url: "/listings?has_apiary=false", want: false},
		{name: "malformed value leaves HasApiary false", url: "/listings?has_apiary=1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)

			f := parseListingFilter(r)

			if f.HasApiary != tt.want {
				t.Errorf("expected HasApiary %v, got %v", tt.want, f.HasApiary)
			}
		})
	}
}

// searchAsUser issues a Search request through OptionalAuth, authenticated as
// userID when non-zero, and returns the filter the store was searched with.
func searchAsUser(t *testing.T, store *captureListingStore, path string, userID int64) repository.ListingFilter {
	t.Helper()
	svc := service.NewListingService(store, &fakeApiaryMembershipReader{})
	h := NewListingHandler(svc)
	handler := middleware.OptionalAuth(testListingAuthSecret)(http.HandlerFunc(h.Search))

	r := httptest.NewRequest("GET", path, nil)
	if userID != 0 {
		tok, err := token.NewAccessToken(userID, testListingAuthSecret, 5)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}
		r.Header.Set("Authorization", "Bearer "+tok)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	return store.lastFilter
}

func TestListingSearch_Authenticated_ExcludesOwnListings(t *testing.T) {
	store := &captureListingStore{}

	f := searchAsUser(t, store, "/listings", 7)

	if f.OwnerUserID != nil {
		t.Errorf("expected OwnerUserID unset, got %v", *f.OwnerUserID)
	}
	if f.ExcludeOwnerUserID == nil || *f.ExcludeOwnerUserID != 7 {
		t.Errorf("expected ExcludeOwnerUserID 7, got %v", f.ExcludeOwnerUserID)
	}
}

func TestListingSearch_Anonymous_DoesNotExclude(t *testing.T) {
	store := &captureListingStore{}

	f := searchAsUser(t, store, "/listings", 0)

	if f.ExcludeOwnerUserID != nil {
		t.Errorf("expected ExcludeOwnerUserID unset, got %v", *f.ExcludeOwnerUserID)
	}
}

func TestListingSearch_Mine_SetsOwnerNotExclude(t *testing.T) {
	store := &captureListingStore{}

	f := searchAsUser(t, store, "/listings?mine=true", 7)

	if f.OwnerUserID == nil || *f.OwnerUserID != 7 {
		t.Errorf("expected OwnerUserID 7, got %v", f.OwnerUserID)
	}
	if f.ExcludeOwnerUserID != nil {
		t.Errorf("expected ExcludeOwnerUserID unset, got %v", *f.ExcludeOwnerUserID)
	}
}
