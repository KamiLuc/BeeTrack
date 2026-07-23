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

// maxListingsPerUser is the maximum number of listings a single user may create.
const maxListingsPerUser = 20

// maxListingPrice is the largest price the `price NUMERIC(10,2)` column can store
// (precision 10, scale 2 — an absolute value below 10^8).
const maxListingPrice = 100_000_000

var (
	ErrListingNotFound            = errors.New("listing not found")
	ErrListingTitleRequired       = errors.New("title is required")
	ErrListingCategoryInvalid     = errors.New("category is invalid")
	ErrListingTooManyImages       = errors.New("a listing may have at most 3 images")
	ErrListingPhotoRequired       = errors.New("a listing must have at least one photo")
	ErrListingLimitReached        = fmt.Errorf("a user may have at most %d listings", maxListingsPerUser)
	ErrListingTitleTooLong        = fmt.Errorf("title must be at most %d characters", validation.Medium.MaxLength())
	ErrListingDescriptionTooLong  = fmt.Errorf("description must be at most %d characters", validation.Large.MaxLength())
	ErrListingQuantityTooLong     = fmt.Errorf("quantity must be at most %d characters", validation.Small.MaxLength())
	ErrListingAddressTooLong      = fmt.Errorf("address must be at most %d characters", validation.Medium.MaxLength())
	ErrListingContactPhoneTooLong = fmt.Errorf("contact phone must be at most %d characters", validation.SuperSmall.MaxLength())
	ErrListingContactEmailTooLong = fmt.Errorf("contact email must be at most %d characters", validation.Medium.MaxLength())
	ErrListingPriceTooLarge       = errors.New("price must be less than 100,000,000")
	ErrListingLocationRequired    = errors.New("location (lat/lng) is required")
	ErrNotListingOwner            = errors.New("not the listing owner")
	ErrHoneyBatchCategoryMismatch = errors.New("attaching a honey batch requires the HONEY category")
	ErrHoneyBatchAlreadyAttached  = errors.New("honey batch is already attached to another listing")
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
	FindByHoneyBatchID(ctx context.Context, batchID int64) (*model.Listing, error)
}

// HoneyBatchReader is the read-only interface into honey batch state used by
// ListingService to validate and display a listing's attached honey batch.
// Satisfied by *HoneyBatchService.
type HoneyBatchReader interface {
	RequireCertifiedOwnedBatch(ctx context.Context, userID, batchID int64) error
	GetBatchByID(ctx context.Context, batchID int64) (*BatchVerification, error)
	VerificationURL(token string) string
	PublicPDFURL(token string) string
}

// ListingService handles business logic for marketplace listings.
type ListingService struct {
	listings     ListingStore
	apiaries     ApiaryMembershipReader
	honeyBatches HoneyBatchReader
}

// NewListingService creates a ListingService with the given dependencies.
func NewListingService(listings ListingStore, apiaries ApiaryMembershipReader, honeyBatches HoneyBatchReader) *ListingService {
	return &ListingService{listings: listings, apiaries: apiaries, honeyBatches: honeyBatches}
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
	HoneyBatchID *int64
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
	if p.HoneyBatchID != nil && p.Category != "HONEY" {
		return ErrHoneyBatchCategoryMismatch
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

// checkListingLimit verifies userID has not reached maxListingsPerUser listings yet.
func (s *ListingService) checkListingLimit(ctx context.Context, userID int64) error {
	count, err := s.listings.Count(ctx, repository.ListingFilter{OwnerUserID: &userID})
	if err != nil {
		return fmt.Errorf("count listings: %w", err)
	}
	if count >= maxListingsPerUser {
		return ErrListingLimitReached
	}
	return nil
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

// checkHoneyBatchAccess verifies the user may attach the given honey batch, if
// one is set — it must be owned by userID and already have a confirmed
// on-chain certification.
func (s *ListingService) checkHoneyBatchAccess(ctx context.Context, batchID *int64, userID int64) error {
	if batchID == nil {
		return nil
	}
	return s.honeyBatches.RequireCertifiedOwnedBatch(ctx, userID, *batchID)
}

// checkHoneyBatchAvailable verifies the given honey batch (if set) isn't
// already attached to a different listing — a batch may back at most one
// live listing at a time. excludeListingID is the listing being updated (0
// when creating), so re-saving a listing with its own already-attached batch
// is not treated as a conflict.
func (s *ListingService) checkHoneyBatchAvailable(ctx context.Context, batchID *int64, excludeListingID int64) error {
	if batchID == nil {
		return nil
	}
	existing, err := s.listings.FindByHoneyBatchID(ctx, *batchID)
	if err != nil {
		return fmt.Errorf("find listing by honey batch: %w", err)
	}
	if existing != nil && existing.ID != excludeListingID {
		return ErrHoneyBatchAlreadyAttached
	}
	return nil
}

// withHoneyBatch populates l's transient HoneyBatch* display fields from its
// attached batch, if any.
func (s *ListingService) withHoneyBatch(ctx context.Context, l *model.Listing) (*model.Listing, error) {
	if l.HoneyBatchID == nil {
		return l, nil
	}
	bv, err := s.honeyBatches.GetBatchByID(ctx, *l.HoneyBatchID)
	if err != nil {
		return nil, fmt.Errorf("get attached honey batch: %w", err)
	}
	l.HoneyBatchHoneyType = bv.Batch.HoneyType
	gatheringDate := bv.Batch.GatheringDate
	l.HoneyBatchGatheringDate = &gatheringDate
	amount := bv.Batch.AmountGrams
	l.HoneyBatchAmountGrams = &amount
	l.HoneyBatchProcessingMethod = bv.Batch.ProcessingMethod
	if bv.Certification != nil {
		l.HoneyBatchCertificationStatus = string(bv.Certification.Status)
		l.HoneyBatchHasPDF = bv.Certification.Status == model.CertificationStatusConfirmed && bv.Batch.PDFFilename != ""
	}
	l.HoneyBatchVerificationURL = s.honeyBatches.VerificationURL(bv.Batch.VerificationToken)
	if l.HoneyBatchHasPDF {
		l.HoneyBatchPDFURL = s.honeyBatches.PublicPDFURL(bv.Batch.VerificationToken)
	}
	return l, nil
}

// Create validates params, verifies apiary and honey batch access, and inserts a new listing with images.
func (s *ListingService) Create(ctx context.Context, userID int64, params ListingParams) (*model.Listing, error) {
	if err := validateListingParams(params); err != nil {
		return nil, err
	}
	if err := s.checkListingLimit(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.checkApiaryAccess(ctx, params.ApiaryID, userID); err != nil {
		return nil, err
	}
	if err := s.checkHoneyBatchAccess(ctx, params.HoneyBatchID, userID); err != nil {
		return nil, err
	}
	if err := s.checkHoneyBatchAvailable(ctx, params.HoneyBatchID, 0); err != nil {
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
		Status:       model.ListingStatusPending,
		Images:       imagesFromURLs(params.ImageURLs),
		HoneyBatchID: params.HoneyBatchID,
	}
	if err := s.listings.Create(ctx, l); err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	return l, nil
}

// Get returns a single listing, with its attached honey batch's details
// populated if one is set. Hidden listings are visible only to their owner.
// Pass viewerUserID 0 for an anonymous (public) viewer.
func (s *ListingService) Get(ctx context.Context, viewerUserID, listingID int64) (*model.Listing, error) {
	l, err := s.listings.GetByID(ctx, listingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if (l.IsHidden || l.Status != model.ListingStatusApproved) && l.UserID != viewerUserID {
		return nil, ErrListingNotFound
	}
	return s.withHoneyBatch(ctx, l)
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

// listingContentChanged reports whether params differ from l's current mutable
// fields. An edit that resubmits identical content must not re-enter moderation.
func listingContentChanged(l *model.Listing, params ListingParams) bool {
	if l.Title != params.Title ||
		l.Description != params.Description ||
		l.Category != params.Category ||
		l.Quantity != params.Quantity ||
		l.Address != params.Address ||
		l.ContactPhone != params.ContactPhone ||
		l.ContactEmail != params.ContactEmail ||
		l.Lat != *params.Lat ||
		l.Lng != *params.Lng ||
		!int64PtrEqual(l.ApiaryID, params.ApiaryID) ||
		!int64PtrEqual(l.HoneyBatchID, params.HoneyBatchID) ||
		!float64PtrEqual(l.Price, defaultPrice(params.Price)) {
		return true
	}
	return params.ImageURLs != nil && !imageURLsEqual(l.Images, params.ImageURLs)
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func float64PtrEqual(a, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func imageURLsEqual(images []model.ListingImage, urls []string) bool {
	if len(images) != len(urls) {
		return false
	}
	for i, img := range images {
		if img.ImageURL != urls[i] {
			return false
		}
	}
	return true
}

// Update validates params, verifies ownership, overwrites mutable fields, and replaces images.
// Content that is identical to what's already stored does not re-enter moderation.
func (s *ListingService) Update(ctx context.Context, userID, listingID int64, params ListingParams) (*model.Listing, error) {
	if err := validateListingParams(params); err != nil {
		return nil, err
	}
	l, err := s.ownedListing(ctx, userID, listingID)
	if err != nil {
		return nil, err
	}
	if params.ImageURLs != nil && len(params.ImageURLs) == 0 {
		return nil, ErrListingPhotoRequired
	}
	if err := s.checkApiaryAccess(ctx, params.ApiaryID, userID); err != nil {
		return nil, err
	}
	if err := s.checkHoneyBatchAccess(ctx, params.HoneyBatchID, userID); err != nil {
		return nil, err
	}
	if err := s.checkHoneyBatchAvailable(ctx, params.HoneyBatchID, listingID); err != nil {
		return nil, err
	}
	if listingContentChanged(l, params) {
		l.Status = model.ListingStatusPending
		l.RejectionReason = nil
		l.ReviewedBy = nil
		l.ReviewedAt = nil
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
	l.HoneyBatchID = params.HoneyBatchID
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
