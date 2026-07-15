package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/internal/validation"
	"gorm.io/gorm"
)

// maxListingImages is the maximum number of images allowed per listing.
const maxListingImages = 3

// maxListingPrice is the largest price the `price NUMERIC(10,2)` column can store
// (precision 10, scale 2 — an absolute value below 10^8).
const maxListingPrice = 100_000_000

var (
	ErrListingNotFound            = errors.New("listing not found")
	ErrListingTitleRequired       = errors.New("title is required")
	ErrListingCategoryInvalid     = errors.New("category is invalid")
	ErrListingTooManyImages       = errors.New("a listing may have at most 3 images")
	ErrListingTitleTooLong        = fmt.Errorf("title must be at most %d characters", validation.Medium.MaxLength())
	ErrListingDescriptionTooLong  = fmt.Errorf("description must be at most %d characters", validation.Large.MaxLength())
	ErrListingQuantityTooLong     = fmt.Errorf("quantity must be at most %d characters", validation.Small.MaxLength())
	ErrListingAddressTooLong      = fmt.Errorf("address must be at most %d characters", validation.Medium.MaxLength())
	ErrListingContactPhoneTooLong = fmt.Errorf("contact phone must be at most %d characters", validation.SuperSmall.MaxLength())
	ErrListingContactEmailTooLong = fmt.Errorf("contact email must be at most %d characters", validation.Medium.MaxLength())
	ErrListingPriceTooLarge       = errors.New("price must be less than 100,000,000")
	ErrListingLocationRequired    = errors.New("location (lat/lng) is required")
	ErrNotListingOwner            = errors.New("not the listing owner")
)

// validListingCategories is the set of accepted listing categories.
var validListingCategories = map[string]bool{
	"HONEY": true, "POLLEN": true, "BEE_COLONIES": true, "QUEEN_BEES": true,
	"BEEHIVES": true, "POPULATED_BEEHIVES": true, "EQUIPMENT": true,
	"EXTRACTION_EQUIPMENT": true, "FEED": true, "SUPPLIES": true,
	"WAX_FOUNDATION": true, "BEESWAX": true, "PROPOLIS": true,
	"SERVICES": true, "OTHER": true,
}

// ListingStore is the persistence interface for listings.
type ListingStore interface {
	Create(ctx context.Context, l *model.Listing) error
	GetByID(ctx context.Context, id int64) (*model.Listing, error)
	Count(ctx context.Context, f repository.ListingFilter) (int64, error)
	List(ctx context.Context, f repository.ListingFilter) ([]*model.Listing, error)
	Update(ctx context.Context, l *model.Listing) error
	SetHidden(ctx context.Context, id int64, hidden bool) error
	AddImages(ctx context.Context, images []model.ListingImage) error
	DeleteImages(ctx context.Context, listingID int64) error
	Delete(ctx context.Context, id int64) error
}

// ListingService handles business logic for marketplace listings.
type ListingService struct {
	listings ListingStore
	apiaries ApiaryMembershipReader
}

// NewListingService creates a ListingService with the given dependencies.
func NewListingService(listings ListingStore, apiaries ApiaryMembershipReader) *ListingService {
	return &ListingService{listings: listings, apiaries: apiaries}
}

// ListingParams holds the mutable fields for create and update operations.
type ListingParams struct {
	Title        string
	Description  string
	Category     string
	Price        *float64
	Quantity     string
	Address      string
	Lat          *float64
	Lng          *float64
	ApiaryID     *int64
	ContactPhone string
	ContactEmail string
	ImageURLs    []string
}

func validateListingParams(p ListingParams) error {
	if p.Title == "" {
		return ErrListingTitleRequired
	}
	if validation.TooLong(p.Title, validation.Medium) {
		return ErrListingTitleTooLong
	}
	if !validListingCategories[p.Category] {
		return ErrListingCategoryInvalid
	}
	if len(p.ImageURLs) > maxListingImages {
		return ErrListingTooManyImages
	}
	if validation.TooLong(p.Description, validation.Large) {
		return ErrListingDescriptionTooLong
	}
	if validation.TooLong(p.Quantity, validation.Small) {
		return ErrListingQuantityTooLong
	}
	if validation.TooLong(p.Address, validation.Medium) {
		return ErrListingAddressTooLong
	}
	if p.Lat == nil || p.Lng == nil {
		return ErrListingLocationRequired
	}
	if !validGPS(p.Lat, p.Lng) {
		return ErrInvalidGPS
	}
	if validation.TooLong(p.ContactPhone, validation.SuperSmall) {
		return ErrListingContactPhoneTooLong
	}
	if validation.TooLong(p.ContactEmail, validation.Medium) {
		return ErrListingContactEmailTooLong
	}
	if p.Price != nil && (*p.Price >= maxListingPrice || *p.Price <= -maxListingPrice) {
		return ErrListingPriceTooLarge
	}
	return nil
}

// defaultPrice returns price, or a pointer to 0 if price is unset — a listing with no price is free.
func defaultPrice(price *float64) *float64 {
	if price != nil {
		return price
	}
	zero := 0.0
	return &zero
}

func imagesFromURLs(urls []string) []model.ListingImage {
	images := make([]model.ListingImage, len(urls))
	for i, url := range urls {
		images[i] = model.ListingImage{ImageURL: url, DisplayOrder: i}
	}
	return images
}

// checkApiaryAccess verifies the user may attach the given apiary, if one is set.
func (s *ListingService) checkApiaryAccess(ctx context.Context, apiaryID *int64, userID int64) error {
	if apiaryID == nil {
		return nil
	}
	if _, _, err := s.apiaries.GetMembership(ctx, *apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApiaryNotFound
		}
		return fmt.Errorf("get apiary: %w", err)
	}
	return nil
}

// Create validates params, verifies apiary access, and inserts a new listing with images.
func (s *ListingService) Create(ctx context.Context, userID int64, params ListingParams) (*model.Listing, error) {
	if err := validateListingParams(params); err != nil {
		return nil, err
	}
	if err := s.checkApiaryAccess(ctx, params.ApiaryID, userID); err != nil {
		return nil, err
	}
	l := &model.Listing{
		UserID:       userID,
		Title:        params.Title,
		Description:  params.Description,
		Category:     params.Category,
		Price:        defaultPrice(params.Price),
		Quantity:     params.Quantity,
		Address:      params.Address,
		Lat:          *params.Lat,
		Lng:          *params.Lng,
		ApiaryID:     params.ApiaryID,
		ContactPhone: params.ContactPhone,
		ContactEmail: params.ContactEmail,
		Images:       imagesFromURLs(params.ImageURLs),
	}
	if err := s.listings.Create(ctx, l); err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	return l, nil
}

// Get returns a single listing. Hidden listings are visible only to their owner.
// Pass viewerUserID 0 for an anonymous (public) viewer.
func (s *ListingService) Get(ctx context.Context, viewerUserID, listingID int64) (*model.Listing, error) {
	l, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if l.IsHidden && l.UserID != viewerUserID {
		return nil, ErrListingNotFound
	}
	return l, nil
}

// Search returns a paginated slice of listings matching the filter and the total count.
func (s *ListingService) Search(ctx context.Context, f repository.ListingFilter) ([]*model.Listing, int64, error) {
	total, err := s.listings.Count(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("count listings: %w", err)
	}
	listings, err := s.listings.List(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("list listings: %w", err)
	}
	return listings, total, nil
}

// Update validates params, verifies ownership, overwrites mutable fields, and replaces images.
func (s *ListingService) Update(ctx context.Context, userID, listingID int64, params ListingParams) (*model.Listing, error) {
	if err := validateListingParams(params); err != nil {
		return nil, err
	}
	l, err := s.ownedListing(ctx, userID, listingID)
	if err != nil {
		return nil, err
	}
	if err := s.checkApiaryAccess(ctx, params.ApiaryID, userID); err != nil {
		return nil, err
	}
	l.Title = params.Title
	l.Description = params.Description
	l.Category = params.Category
	l.Price = defaultPrice(params.Price)
	l.Quantity = params.Quantity
	l.Address = params.Address
	l.Lat = *params.Lat
	l.Lng = *params.Lng
	l.ApiaryID = params.ApiaryID
	l.ContactPhone = params.ContactPhone
	l.ContactEmail = params.ContactEmail
	if err := s.listings.Update(ctx, l); err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}
	if params.ImageURLs != nil {
		if err := s.listings.DeleteImages(ctx, listingID); err != nil {
			return nil, fmt.Errorf("delete listing images: %w", err)
		}
		images := imagesFromURLs(params.ImageURLs)
		for i := range images {
			images[i].ListingID = listingID
		}
		if err := s.listings.AddImages(ctx, images); err != nil {
			return nil, fmt.Errorf("add listing images: %w", err)
		}
		l.Images = images
	}
	return l, nil
}

// SetHidden toggles the visibility of a listing after verifying ownership.
func (s *ListingService) SetHidden(ctx context.Context, userID, listingID int64, hidden bool) error {
	if _, err := s.ownedListing(ctx, userID, listingID); err != nil {
		return err
	}
	if err := s.listings.SetHidden(ctx, listingID, hidden); err != nil {
		return fmt.Errorf("set hidden: %w", err)
	}
	return nil
}

// Delete removes a listing after verifying ownership.
func (s *ListingService) Delete(ctx context.Context, userID, listingID int64) error {
	if _, err := s.ownedListing(ctx, userID, listingID); err != nil {
		return err
	}
	if err := s.listings.Delete(ctx, listingID); err != nil {
		return fmt.Errorf("delete listing: %w", err)
	}
	return nil
}

// ownedListing fetches a listing and verifies it belongs to userID.
func (s *ListingService) ownedListing(ctx context.Context, userID, listingID int64) (*model.Listing, error) {
	l, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if l.UserID != userID {
		return nil, ErrNotListingOwner
	}
	return l, nil
}
