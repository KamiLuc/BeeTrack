package service

import (
	"context"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
	"gorm.io/gorm"
)

type mockListingStore struct {
	listing      *model.Listing
	listings     []*model.Listing
	created      *model.Listing
	updated      *model.Listing
	deletedID    int64
	hiddenID     int64
	hiddenValue  bool
	imagesDelIID int64
	addedImages  []model.ListingImage
}

func (m *mockListingStore) Create(ctx context.Context, l *model.Listing) error {
	l.ID = 1
	m.created = l
	return nil
}

func (m *mockListingStore) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	if m.listing == nil || m.listing.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return m.listing, nil
}

func (m *mockListingStore) Count(ctx context.Context, f repository.ListingFilter) (int64, error) {
	return int64(len(m.listings)), nil
}

func (m *mockListingStore) List(ctx context.Context, f repository.ListingFilter) ([]*model.Listing, error) {
	return m.listings, nil
}

func (m *mockListingStore) Update(ctx context.Context, l *model.Listing) error {
	m.updated = l
	return nil
}

func (m *mockListingStore) SetHidden(ctx context.Context, id int64, hidden bool) error {
	m.hiddenID = id
	m.hiddenValue = hidden
	return nil
}

func (m *mockListingStore) AddImages(ctx context.Context, images []model.ListingImage) error {
	m.addedImages = images
	return nil
}

func (m *mockListingStore) DeleteImages(ctx context.Context, listingID int64) error {
	m.imagesDelIID = listingID
	return nil
}

func (m *mockListingStore) Delete(ctx context.Context, id int64) error {
	m.deletedID = id
	return nil
}

func newListingSvc(store *mockListingStore) *ListingService {
	return NewListingService(store, &mockApiaryRepo{apiary: &model.Apiary{ID: 1}, role: "member"})
}

func validListingParams() ListingParams {
	return ListingParams{
		Title:    "Raw wildflower honey",
		Category: "HONEY",
	}
}

func TestListingCreate(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	l, err := svc.Create(context.Background(), 7, validListingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.UserID != 7 {
		t.Errorf("expected user_id 7, got %d", l.UserID)
	}
	if l.Title != "Raw wildflower honey" {
		t.Errorf("unexpected title %q", l.Title)
	}
	if store.created == nil {
		t.Error("expected listing to be persisted")
	}
}

func TestListingCreate_MissingTitle(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	_, err := svc.Create(context.Background(), 1, ListingParams{Category: "HONEY"})
	if err != ErrListingTitleRequired {
		t.Errorf("expected ErrListingTitleRequired, got %v", err)
	}
}

func TestListingCreate_InvalidCategory(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	_, err := svc.Create(context.Background(), 1, ListingParams{Title: "x", Category: "GOLD"})
	if err != ErrListingCategoryInvalid {
		t.Errorf("expected ErrListingCategoryInvalid, got %v", err)
	}
}

func TestListingCreate_TooManyImages(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.ImageURLs = []string{"a", "b", "c", "d"}
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingTooManyImages {
		t.Errorf("expected ErrListingTooManyImages, got %v", err)
	}
}

func TestListingCreate_DescriptionAtMaxLength(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	params := validListingParams()
	params.Description = strings.Repeat("a", 500)
	_, err := svc.Create(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListingCreate_DescriptionTooLong(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.Description = strings.Repeat("a", 501)
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingDescriptionTooLong {
		t.Errorf("expected ErrListingDescriptionTooLong, got %v", err)
	}
}

func TestListingCreate_DescriptionTooLongMultiByteRunes(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.Description = strings.Repeat("蜂", 501)
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingDescriptionTooLong {
		t.Errorf("expected ErrListingDescriptionTooLong, got %v", err)
	}
}

func TestListingCreate_TitleTooLong(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.Title = strings.Repeat("a", 151)
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingTitleTooLong {
		t.Errorf("expected ErrListingTitleTooLong, got %v", err)
	}
}

func TestListingCreate_TitleAtMaxLength(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.Title = strings.Repeat("a", 150)
	_, err := svc.Create(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListingCreate_QuantityTooLong(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.Quantity = strings.Repeat("a", 51)
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingQuantityTooLong {
		t.Errorf("expected ErrListingQuantityTooLong, got %v", err)
	}
}

func TestListingCreate_AddressTooLong(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.Address = strings.Repeat("a", 151)
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingAddressTooLong {
		t.Errorf("expected ErrListingAddressTooLong, got %v", err)
	}
}

func TestListingCreate_ContactPhoneTooLong(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.ContactPhone = strings.Repeat("1", 21)
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingContactPhoneTooLong {
		t.Errorf("expected ErrListingContactPhoneTooLong, got %v", err)
	}
}

func TestListingCreate_ContactEmailTooLong(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	params.ContactEmail = strings.Repeat("a", 151) + "@example.com"
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingContactEmailTooLong {
		t.Errorf("expected ErrListingContactEmailTooLong, got %v", err)
	}
}

func TestListingCreate_PriceAtMax(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	params := validListingParams()
	price := 99_999_999.99
	params.Price = &price
	_, err := svc.Create(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListingCreate_PriceTooLarge(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	price := 883_154_044.0
	params.Price = &price
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingPriceTooLarge {
		t.Errorf("expected ErrListingPriceTooLarge, got %v", err)
	}
}

func TestListingCreate_PriceTooNegative(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	params := validListingParams()
	price := -883_154_044.0
	params.Price = &price
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrListingPriceTooLarge {
		t.Errorf("expected ErrListingPriceTooLarge, got %v", err)
	}
}

func TestListingCreate_WithImages(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	params := validListingParams()
	params.ImageURLs = []string{"a.jpg", "b.jpg"}
	l, err := svc.Create(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(l.Images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(l.Images))
	}
	if l.Images[1].DisplayOrder != 1 {
		t.Errorf("expected display_order 1, got %d", l.Images[1].DisplayOrder)
	}
}

func TestListingCreate_DefaultsMissingPriceToZero(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	l, err := svc.Create(context.Background(), 1, validListingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Price == nil || *l.Price != 0 {
		t.Errorf("expected price defaulted to 0, got %v", l.Price)
	}
}

func TestListingCreate_KeepsExplicitPrice(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	price := 12.5
	params := validListingParams()
	params.Price = &price
	l, err := svc.Create(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Price == nil || *l.Price != 12.5 {
		t.Errorf("expected price 12.5, got %v", l.Price)
	}
}

func TestListingCreate_UnknownApiary(t *testing.T) {
	store := &mockListingStore{}
	// mockApiaryRepo returns NotFound when apiary id doesn't match the seeded one (id 1).
	svc := newListingSvc(store)

	other := int64(99)
	params := validListingParams()
	params.ApiaryID = &other
	_, err := svc.Create(context.Background(), 1, params)
	if err != ErrApiaryNotFound {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestListingGet_HiddenVisibleToOwner(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3, IsHidden: true}}
	svc := newListingSvc(store)

	l, err := svc.Get(context.Background(), 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.ID != 5 {
		t.Errorf("expected listing 5, got %d", l.ID)
	}
}

func TestListingGet_HiddenHiddenFromOthers(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3, IsHidden: true}}
	svc := newListingSvc(store)

	_, err := svc.Get(context.Background(), 99, 5)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingGet_HiddenHiddenFromAnonymous(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3, IsHidden: true}}
	svc := newListingSvc(store)

	_, err := svc.Get(context.Background(), 0, 5)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingGet_NotFound(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	_, err := svc.Get(context.Background(), 1, 5)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingSearch(t *testing.T) {
	store := &mockListingStore{listings: []*model.Listing{
		{ID: 1, Title: "A"},
		{ID: 2, Title: "B"},
	}}
	svc := newListingSvc(store)

	listings, total, err := svc.Search(context.Background(), repository.ListingFilter{Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(listings) != 2 || total != 2 {
		t.Errorf("expected 2 listings and total 2, got %d and %d", len(listings), total)
	}
}

func TestListingUpdate(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3, Title: "old"}}
	svc := newListingSvc(store)

	params := validListingParams()
	params.Title = "new title"
	l, err := svc.Update(context.Background(), 3, 5, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Title != "new title" {
		t.Errorf("expected updated title, got %q", l.Title)
	}
	if store.updated == nil {
		t.Error("expected update to be persisted")
	}
}

func TestListingUpdate_DefaultsMissingPriceToZero(t *testing.T) {
	price := 9.0
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3, Price: &price}}
	svc := newListingSvc(store)

	l, err := svc.Update(context.Background(), 3, 5, validListingParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Price == nil || *l.Price != 0 {
		t.Errorf("expected price defaulted to 0, got %v", l.Price)
	}
}

func TestListingUpdate_KeepsExplicitPrice(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	price := 30.0
	params := validListingParams()
	params.Price = &price
	l, err := svc.Update(context.Background(), 3, 5, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Price == nil || *l.Price != 30.0 {
		t.Errorf("expected price 30.0, got %v", l.Price)
	}
}

func TestListingUpdate_NotFound(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	_, err := svc.Update(context.Background(), 3, 5, validListingParams())
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingUpdate_MissingTitle(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	_, err := svc.Update(context.Background(), 3, 5, ListingParams{Category: "HONEY"})
	if err != ErrListingTitleRequired {
		t.Errorf("expected ErrListingTitleRequired, got %v", err)
	}
}

func TestListingUpdate_InvalidCategory(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	_, err := svc.Update(context.Background(), 3, 5, ListingParams{Title: "x", Category: "GOLD"})
	if err != ErrListingCategoryInvalid {
		t.Errorf("expected ErrListingCategoryInvalid, got %v", err)
	}
}

func TestListingUpdate_TooManyImages(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	params := validListingParams()
	params.ImageURLs = []string{"a", "b", "c", "d"}
	_, err := svc.Update(context.Background(), 3, 5, params)
	if err != ErrListingTooManyImages {
		t.Errorf("expected ErrListingTooManyImages, got %v", err)
	}
}

func TestListingUpdate_UnknownApiary(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	other := int64(99)
	params := validListingParams()
	params.ApiaryID = &other
	_, err := svc.Update(context.Background(), 3, 5, params)
	if err != ErrApiaryNotFound {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestListingUpdate_ValidatesBeforeOwnershipCheck(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	_, err := svc.Update(context.Background(), 3, 5, ListingParams{})
	if err != ErrListingTitleRequired {
		t.Errorf("expected ErrListingTitleRequired, got %v", err)
	}
}

func TestListingUpdate_NotOwner(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	_, err := svc.Update(context.Background(), 99, 5, validListingParams())
	if err != ErrNotListingOwner {
		t.Errorf("expected ErrNotListingOwner, got %v", err)
	}
}

func TestListingUpdate_DescriptionAtMaxLength(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	params := validListingParams()
	params.Description = strings.Repeat("a", 500)
	_, err := svc.Update(context.Background(), 3, 5, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListingUpdate_DescriptionTooLong(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	params := validListingParams()
	params.Description = strings.Repeat("a", 501)
	_, err := svc.Update(context.Background(), 3, 5, params)
	if err != ErrListingDescriptionTooLong {
		t.Errorf("expected ErrListingDescriptionTooLong, got %v", err)
	}
}

func TestListingUpdate_ReplacesImages(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	params := validListingParams()
	params.ImageURLs = []string{"new.jpg"}
	_, err := svc.Update(context.Background(), 3, 5, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.imagesDelIID != 5 {
		t.Errorf("expected old images deleted for listing 5, got %d", store.imagesDelIID)
	}
	if len(store.addedImages) != 1 || store.addedImages[0].ListingID != 5 {
		t.Errorf("expected new image attached to listing 5, got %+v", store.addedImages)
	}
}

func TestListingSetHidden(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	if err := svc.SetHidden(context.Background(), 3, 5, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.hiddenID != 5 || !store.hiddenValue {
		t.Errorf("expected listing 5 hidden=true, got %d hidden=%v", store.hiddenID, store.hiddenValue)
	}
}

func TestListingSetHidden_NotOwner(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	err := svc.SetHidden(context.Background(), 99, 5, true)
	if err != ErrNotListingOwner {
		t.Errorf("expected ErrNotListingOwner, got %v", err)
	}
}

func TestListingSetHidden_NotFound(t *testing.T) {
	svc := newListingSvc(&mockListingStore{})

	err := svc.SetHidden(context.Background(), 3, 5, true)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestListingDelete(t *testing.T) {
	store := &mockListingStore{listing: &model.Listing{ID: 5, UserID: 3}}
	svc := newListingSvc(store)

	if err := svc.Delete(context.Background(), 3, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.deletedID != 5 {
		t.Errorf("expected deletedID 5, got %d", store.deletedID)
	}
}

func TestListingDelete_NotFound(t *testing.T) {
	store := &mockListingStore{}
	svc := newListingSvc(store)

	err := svc.Delete(context.Background(), 3, 5)
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}
