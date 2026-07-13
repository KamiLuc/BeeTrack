package service

import (
	"context"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockFavoriteStore struct {
	added     *model.ListingFavorite
	removed   [2]int64
	listings  []*model.Listing
	removeHit bool
	addErr    error
	removeErr error
	listErr   error
}

func (m *mockFavoriteStore) Add(ctx context.Context, f *model.ListingFavorite) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.added = f
	return nil
}

func (m *mockFavoriteStore) Remove(ctx context.Context, userID, listingID int64) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	m.removed = [2]int64{userID, listingID}
	m.removeHit = true
	return nil
}

func (m *mockFavoriteStore) ListListingsByUserID(ctx context.Context, userID int64) ([]*model.Listing, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listings, nil
}

// mockFavoriteListingReader reuses mockListingReader's shape for listing lookups.
type mockFavoriteListingReader struct {
	listing *model.Listing
	err     error
}

func (m *mockFavoriteListingReader) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.listing == nil || m.listing.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return m.listing, nil
}

func newFavoriteSvc(store *mockFavoriteStore, reader *mockFavoriteListingReader) *ListingFavoriteService {
	return NewListingFavoriteService(store, reader)
}

func TestFavoriteAdd(t *testing.T) {
	store := &mockFavoriteStore{}
	reader := &mockFavoriteListingReader{listing: &model.Listing{ID: 5, UserID: 9}}
	svc := newFavoriteSvc(store, reader)

	if err := svc.Add(context.Background(), 3, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.added == nil || store.added.UserID != 3 || store.added.ListingID != 5 {
		t.Errorf("expected favorite (user 3, listing 5), got %+v", store.added)
	}
}

func TestFavoriteAdd_ListingNotFound(t *testing.T) {
	store := &mockFavoriteStore{}
	svc := newFavoriteSvc(store, &mockFavoriteListingReader{})

	err := svc.Add(context.Background(), 3, 5)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
	if store.added != nil {
		t.Error("expected no favorite to be added")
	}
}

func TestFavoriteAdd_HiddenNotOwner(t *testing.T) {
	store := &mockFavoriteStore{}
	reader := &mockFavoriteListingReader{listing: &model.Listing{ID: 5, UserID: 9, IsHidden: true}}
	svc := newFavoriteSvc(store, reader)

	err := svc.Add(context.Background(), 3, 5)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestFavoriteAdd_HiddenOwner(t *testing.T) {
	store := &mockFavoriteStore{}
	reader := &mockFavoriteListingReader{listing: &model.Listing{ID: 5, UserID: 3, IsHidden: true}}
	svc := newFavoriteSvc(store, reader)

	err := svc.Add(context.Background(), 3, 5)
	if err != ErrCannotFavoriteOwnListing {
		t.Errorf("expected ErrCannotFavoriteOwnListing, got %v", err)
	}
}

func TestFavoriteAdd_OwnListing(t *testing.T) {
	store := &mockFavoriteStore{}
	reader := &mockFavoriteListingReader{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newFavoriteSvc(store, reader)

	err := svc.Add(context.Background(), 3, 5)
	if err != ErrCannotFavoriteOwnListing {
		t.Errorf("expected ErrCannotFavoriteOwnListing, got %v", err)
	}
	if store.added != nil {
		t.Error("expected no favorite to be added")
	}
}

func TestFavoriteAdd_ReaderError(t *testing.T) {
	store := &mockFavoriteStore{}
	svc := newFavoriteSvc(store, &mockFavoriteListingReader{err: errors.New("db down")})

	err := svc.Add(context.Background(), 3, 5)
	if err == nil || errors.Is(err, ErrListingNotFound) {
		t.Errorf("expected wrapped db error, got %v", err)
	}
	if store.added != nil {
		t.Error("expected no favorite to be added")
	}
}

func TestFavoriteAdd_StoreError(t *testing.T) {
	store := &mockFavoriteStore{addErr: errors.New("insert failed")}
	reader := &mockFavoriteListingReader{listing: &model.Listing{ID: 5, UserID: 9}}
	svc := newFavoriteSvc(store, reader)

	if err := svc.Add(context.Background(), 3, 5); err == nil {
		t.Error("expected error from store.Add to propagate")
	}
}

func TestFavoriteRemove(t *testing.T) {
	store := &mockFavoriteStore{}
	svc := newFavoriteSvc(store, &mockFavoriteListingReader{})

	if err := svc.Remove(context.Background(), 3, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !store.removeHit || store.removed != [2]int64{3, 5} {
		t.Errorf("expected remove(3, 5), got %v", store.removed)
	}
}

func TestFavoriteRemove_StoreError(t *testing.T) {
	store := &mockFavoriteStore{removeErr: errors.New("delete failed")}
	svc := newFavoriteSvc(store, &mockFavoriteListingReader{})

	if err := svc.Remove(context.Background(), 3, 5); err == nil {
		t.Error("expected error from store.Remove to propagate")
	}
}

func TestFavoriteList(t *testing.T) {
	store := &mockFavoriteStore{listings: []*model.Listing{
		{ID: 1, Title: "A"},
		{ID: 2, Title: "B"},
	}}
	svc := newFavoriteSvc(store, &mockFavoriteListingReader{})

	listings, err := svc.List(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(listings) != 2 {
		t.Errorf("expected 2 favorites, got %d", len(listings))
	}
}

func TestFavoriteList_StoreError(t *testing.T) {
	store := &mockFavoriteStore{listErr: errors.New("query failed")}
	svc := newFavoriteSvc(store, &mockFavoriteListingReader{})

	listings, err := svc.List(context.Background(), 3)
	if err == nil {
		t.Error("expected error from store.ListListingsByUserID to propagate")
	}
	if listings != nil {
		t.Errorf("expected nil listings on error, got %v", listings)
	}
}
