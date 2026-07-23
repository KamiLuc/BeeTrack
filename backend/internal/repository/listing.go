package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

// ListingFilter holds optional criteria for searching listings.
type ListingFilter struct {
	Category           string
	Keyword            string
	PriceMin           *float64
	PriceMax           *float64
	PostedAfter        *string
	OwnerUserID        *int64
	ExcludeOwnerUserID *int64
	NearLat            *float64
	NearLng            *float64
	RadiusKm           *float64
	HasApiary          bool
	Status             string
	Limit              int
	Offset             int
}

// hasNearFilter reports whether f carries a complete distance filter — all of
// NearLat, NearLng, and RadiusKm must be set together, otherwise it's ignored.
func (f ListingFilter) hasNearFilter() bool {
	return f.NearLat != nil && f.NearLng != nil && f.RadiusKm != nil
}

// haversineKmExpr builds a SQL expression computing the great-circle distance in
// kilometers between (NearLat, NearLng) and the row's lat/lng columns (prefixed by p).
// Returns the expression and its three bind params (lat, lng, lat), in that order.
func (f ListingFilter) haversineKmExpr(p string) (string, []any) {
	expr := fmt.Sprintf(
		"6371 * acos(cos(radians(?)) * cos(radians(%[1]slat)) * cos(radians(%[1]slng) - radians(?)) + sin(radians(?)) * sin(radians(%[1]slat)))",
		p,
	)
	return expr, []any{*f.NearLat, *f.NearLng, *f.NearLat}
}

// ListingRepository persists marketplace listings and their images.
type ListingRepository struct {
	db *gorm.DB
}

// NewListingRepository creates a new ListingRepository backed by db.
func NewListingRepository(db *gorm.DB) *ListingRepository {
	return &ListingRepository{db: db}
}

// Create inserts a listing and its images in a single transaction.
func (r *ListingRepository) Create(ctx context.Context, l *model.Listing) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(l).Error; err != nil {
			return err
		}
		for i := range l.Images {
			l.Images[i].ListingID = l.ID
			if err := tx.Create(&l.Images[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID returns the listing with the given id, including its images and attached
// apiary's name, GPS, and hive count.
func (r *ListingRepository) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	type row struct {
		model.Listing
		ApiaryName      string   `gorm:"column:apiary_name"`
		ApiaryLat       *float64 `gorm:"column:apiary_lat"`
		ApiaryLng       *float64 `gorm:"column:apiary_lng"`
		ApiaryHiveCount int      `gorm:"column:apiary_hive_count"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("listings l").
		Select("l.*, a.name AS apiary_name, a.lat AS apiary_lat, a.lng AS apiary_lng, "+
			"(SELECT COUNT(*) FROM hives h WHERE h.apiary_id = a.id) AS apiary_hive_count").
		Joins("LEFT JOIN apiaries a ON a.id = l.apiary_id").
		Where("l.id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	listing := result.Listing
	listing.ApiaryName = result.ApiaryName
	listing.ApiaryLat = result.ApiaryLat
	listing.ApiaryLng = result.ApiaryLng
	listing.ApiaryHiveCount = result.ApiaryHiveCount
	images, err := r.listImages(ctx, id)
	if err != nil {
		return nil, err
	}
	listing.Images = images
	return &listing, nil
}

// GetByIDForReview returns the listing with the given id for the admin queue,
// including its images and the owner account's email.
func (r *ListingRepository) GetByIDForReview(ctx context.Context, id int64) (*model.Listing, error) {
	type row struct {
		model.Listing
		OwnerEmail string `gorm:"column:owner_email"`
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("listings l").
		Select("l.*, u.email AS owner_email").
		Joins("JOIN users u ON u.id = l.user_id").
		Where("l.id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	listing := result.Listing
	listing.OwnerEmail = result.OwnerEmail
	images, err := r.listImages(ctx, id)
	if err != nil {
		return nil, err
	}
	listing.Images = images
	return &listing, nil
}

// Count returns the number of listings matching the filter.
func (r *ListingRepository) Count(ctx context.Context, f ListingFilter) (int64, error) {
	var count int64
	err := r.applyFilter(r.db.WithContext(ctx).Model(&model.Listing{}), f).
		Count(&count).Error
	return count, err
}

// List returns listings matching the filter, ordered by distance ascending when a
// near-filter is active, otherwise by created_at descending; with pagination.
func (r *ListingRepository) List(ctx context.Context, f ListingFilter) ([]*model.Listing, error) {
	type row struct {
		model.Listing
		ApiaryName string   `gorm:"column:apiary_name"`
		DistanceKm *float64 `gorm:"column:distance_km"`
	}
	var rows []row
	selectCols := "l.*, a.name AS apiary_name"
	orderBy := "l.created_at DESC"
	var selectArgs []any
	if f.hasNearFilter() {
		distanceExpr, args := f.haversineKmExpr("l.")
		selectCols += ", " + distanceExpr + " AS distance_km"
		selectArgs = args
		orderBy = "distance_km ASC"
	}
	q := r.db.WithContext(ctx).
		Table("listings l").
		Select(selectCols, selectArgs...).
		Joins("LEFT JOIN apiaries a ON a.id = l.apiary_id")
	q = r.applyFilterAliased(q, f)
	err := q.Order(orderBy).
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	listings := make([]*model.Listing, len(rows))
	ids := make([]int64, len(rows))
	for i, row := range rows {
		l := row.Listing
		l.ApiaryName = row.ApiaryName
		l.DistanceKm = row.DistanceKm
		listings[i] = &l
		ids[i] = l.ID
	}
	imagesByListingID, err := r.listImagesForListingIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, l := range listings {
		l.Images = imagesByListingID[l.ID]
	}
	return listings, nil
}

// Update persists all mutable fields of l, including its moderation status —
// the caller (service layer) decides whether an edit's content actually
// changed enough to reset status back to pending for re-approval.
func (r *ListingRepository) Update(ctx context.Context, l *model.Listing) error {
	return r.db.WithContext(ctx).
		Model(l).
		Updates(map[string]any{
			"title":            l.Title,
			"description":      l.Description,
			"category":         l.Category,
			"price":            l.Price,
			"quantity":         l.Quantity,
			"address":          l.Address,
			"lat":              l.Lat,
			"lng":              l.Lng,
			"apiary_id":        l.ApiaryID,
			"contact_phone":    l.ContactPhone,
			"contact_email":    l.ContactEmail,
			"status":           l.Status,
			"rejection_reason": l.RejectionReason,
			"reviewed_by":      l.ReviewedBy,
			"reviewed_at":      l.ReviewedAt,
			"honey_batch_id":   l.HoneyBatchID,
			"updated_at":       gorm.Expr("NOW()"),
		}).Error
}

// FindByHoneyBatchID returns the listing that has batchID attached, or nil if none does.
func (r *ListingRepository) FindByHoneyBatchID(ctx context.Context, batchID int64) (*model.Listing, error) {
	var l model.Listing
	err := r.db.WithContext(ctx).
		Where("honey_batch_id = ?", batchID).
		First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// SetHidden toggles the is_hidden flag of the listing with the given id.
func (r *ListingRepository) SetHidden(ctx context.Context, id int64, hidden bool) error {
	return r.db.WithContext(ctx).
		Model(&model.Listing{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_hidden":  hidden,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// ListReview returns listings for the admin queue with each row's owner account
// email joined in, optionally filtered by status (empty means all statuses) and
// by keyword (matched against title and owner email), ordered by created_at
// according to sortDir ("asc" or "desc", default "asc").
func (r *ListingRepository) ListReview(ctx context.Context, status, keyword, sortDir string, limit, offset int) ([]*model.Listing, int64, error) {
	order := "l.created_at ASC"
	if sortDir == "desc" {
		order = "l.created_at DESC"
	}

	base := r.db.WithContext(ctx).
		Table("listings l").
		Joins("JOIN users u ON u.id = l.user_id")
	if status != "" {
		base = base.Where("l.status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("(l.title ILIKE ? OR u.email ILIKE ?)", like, like)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Model(&model.Listing{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		model.Listing
		OwnerEmail string `gorm:"column:owner_email"`
	}
	var rows []row
	if err := base.Session(&gorm.Session{}).
		Select("l.*, u.email AS owner_email").
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	listings := make([]*model.Listing, len(rows))
	for i, rr := range rows {
		l := rr.Listing
		l.OwnerEmail = rr.OwnerEmail
		listings[i] = &l
	}
	return listings, total, nil
}

// Approve sets a listing's status to approved and stamps the reviewer, also
// setting first_approved_at if this is its first approval.
func (r *ListingRepository) Approve(ctx context.Context, id, reviewerID int64) error {
	return r.db.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":            model.ListingStatusApproved,
			"rejection_reason":  nil,
			"reviewed_by":       reviewerID,
			"reviewed_at":       gorm.Expr("NOW()"),
			"first_approved_at": gorm.Expr("COALESCE(first_approved_at, NOW())"),
			"updated_at":        gorm.Expr("NOW()"),
		}).Error
}

// Reject sets a listing's status to rejected with the given reason and stamps the reviewer.
func (r *ListingRepository) Reject(ctx context.Context, id, reviewerID int64, reason string) error {
	return r.db.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":           model.ListingStatusRejected,
			"rejection_reason": reason,
			"reviewed_by":      reviewerID,
			"reviewed_at":      gorm.Expr("NOW()"),
			"updated_at":       gorm.Expr("NOW()"),
		}).Error
}

// Remove sets a live listing's status to removed with the given reason and stamps the reviewer.
func (r *ListingRepository) Remove(ctx context.Context, id, reviewerID int64, reason string) error {
	return r.db.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":           model.ListingStatusRemoved,
			"rejection_reason": reason,
			"reviewed_by":      reviewerID,
			"reviewed_at":      gorm.Expr("NOW()"),
			"updated_at":       gorm.Expr("NOW()"),
		}).Error
}

// Restore sets a removed listing's status back to approved and stamps the reviewer.
func (r *ListingRepository) Restore(ctx context.Context, id, reviewerID int64) error {
	return r.db.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":           model.ListingStatusApproved,
			"rejection_reason": nil,
			"reviewed_by":      reviewerID,
			"reviewed_at":      gorm.Expr("NOW()"),
			"updated_at":       gorm.Expr("NOW()"),
		}).Error
}

// AddImages inserts the given images.
func (r *ListingRepository) AddImages(ctx context.Context, images []model.ListingImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&images).Error
}

// CreateImage inserts a single listing image.
func (r *ListingRepository) CreateImage(ctx context.Context, img *model.ListingImage) error {
	return r.db.WithContext(ctx).Create(img).Error
}

// GetImageByID returns the image with the given id that belongs to listingID.
func (r *ListingRepository) GetImageByID(ctx context.Context, imageID, listingID int64) (*model.ListingImage, error) {
	var img model.ListingImage
	err := r.db.WithContext(ctx).
		Where("id = ? AND listing_id = ?", imageID, listingID).
		First(&img).Error
	return &img, err
}

// ListImagesByListingID returns the images for listingID ordered by display_order.
func (r *ListingRepository) ListImagesByListingID(ctx context.Context, listingID int64) ([]model.ListingImage, error) {
	return r.listImages(ctx, listingID)
}

// DeleteImage removes the image with the given id.
func (r *ListingRepository) DeleteImage(ctx context.Context, imageID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", imageID).
		Delete(&model.ListingImage{}).Error
}

// DeleteImages removes all images belonging to listingID.
func (r *ListingRepository) DeleteImages(ctx context.Context, listingID int64) error {
	return r.db.WithContext(ctx).
		Where("listing_id = ?", listingID).
		Delete(&model.ListingImage{}).Error
}

// Delete removes the listing with the given id; its images cascade in the database.
func (r *ListingRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.Listing{}).Error
}

// listImages returns the images for listingID ordered by display_order.
func (r *ListingRepository) listImages(ctx context.Context, listingID int64) ([]model.ListingImage, error) {
	var images []model.ListingImage
	err := r.db.WithContext(ctx).
		Where("listing_id = ?", listingID).
		Order("display_order ASC").
		Find(&images).Error
	return images, err
}

// listImagesForListingIDs returns the images for the given listing ids, grouped by listing id
// and ordered by display_order, in a single query.
func (r *ListingRepository) listImagesForListingIDs(ctx context.Context, listingIDs []int64) (map[int64][]model.ListingImage, error) {
	return fetchImagesForListingIDs(ctx, r.db, listingIDs)
}

// fetchImagesForListingIDs returns the images for the given listing ids, grouped by listing id
// and ordered by display_order, in a single query. Shared across repositories in this package
// that need to batch-populate model.Listing.Images (which gorm never scans automatically).
func fetchImagesForListingIDs(ctx context.Context, db *gorm.DB, listingIDs []int64) (map[int64][]model.ListingImage, error) {
	result := make(map[int64][]model.ListingImage, len(listingIDs))
	if len(listingIDs) == 0 {
		return result, nil
	}
	var images []model.ListingImage
	if err := db.WithContext(ctx).
		Where("listing_id IN ?", listingIDs).
		Order("display_order ASC").
		Find(&images).Error; err != nil {
		return nil, err
	}
	for _, img := range images {
		result[img.ListingID] = append(result[img.ListingID], img)
	}
	return result, nil
}

// applyFilter adds the filter's WHERE clauses to an unaliased listings query.
func (r *ListingRepository) applyFilter(q *gorm.DB, f ListingFilter) *gorm.DB {
	return r.buildFilter(q, f, "")
}

// applyFilterAliased adds the filter's WHERE clauses to a query using the "l" alias.
func (r *ListingRepository) applyFilterAliased(q *gorm.DB, f ListingFilter) *gorm.DB {
	return r.buildFilter(q, f, "l.")
}

// buildFilter applies shared listing filter conditions using the given column prefix.
func (r *ListingRepository) buildFilter(q *gorm.DB, f ListingFilter, p string) *gorm.DB {
	switch {
	case f.OwnerUserID != nil:
		q = q.Where(p+"user_id = ?", *f.OwnerUserID)
	case f.Status != "":
		q = q.Where(p+"status = ?", f.Status)
	default:
		q = q.Where(p+"is_hidden = FALSE").Where(p+"status = ?", model.ListingStatusApproved)
	}
	if f.Category != "" {
		q = q.Where(p+"category = ?", f.Category)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("("+p+"title ILIKE ? OR "+p+"description ILIKE ?)", like, like)
	}
	if f.PriceMin != nil {
		q = q.Where(p+"price >= ?", *f.PriceMin)
	}
	if f.PriceMax != nil {
		q = q.Where(p+"price <= ?", *f.PriceMax)
	}
	if f.PostedAfter != nil {
		q = q.Where(p+"created_at >= ?", *f.PostedAfter)
	}
	if f.hasNearFilter() {
		distanceExpr, args := f.haversineKmExpr(p)
		q = q.Where(distanceExpr+" <= ?", append(args, *f.RadiusKm)...)
	}
	if f.HasApiary {
		q = q.Where(p + "apiary_id IS NOT NULL")
	}
	if f.ExcludeOwnerUserID != nil {
		q = q.Where(p+"user_id != ?", *f.ExcludeOwnerUserID)
	}
	return q
}
