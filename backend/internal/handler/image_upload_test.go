package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/token"
	"gorm.io/gorm"
)

const testUploadAuthSecret = "test-upload-secret"

// fakeApiaryMembershipReader is a minimal service.ApiaryMembershipReader for handler tests.
type fakeApiaryMembershipReader struct {
	apiary *model.Apiary
}

func (f *fakeApiaryMembershipReader) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	if f.apiary == nil {
		return nil, "", gorm.ErrRecordNotFound
	}
	return f.apiary, "owner", nil
}

// fakeInspectionHiveReader is a minimal service.InspectionHiveReader for handler tests.
type fakeInspectionHiveReader struct {
	hive *model.Hive
}

func (f *fakeInspectionHiveReader) GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error) {
	if f.hive == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.hive, nil
}

// fakeInspectionRepo is a minimal service.InspectionRepository for handler tests; only GetByID is exercised.
type fakeInspectionRepo struct {
	inspection *model.Inspection
}

func (f *fakeInspectionRepo) Create(ctx context.Context, insp *model.Inspection) error { return nil }
func (f *fakeInspectionRepo) CreateDisease(ctx context.Context, d *model.InspectionDisease) error {
	return nil
}
func (f *fakeInspectionRepo) Delete(ctx context.Context, inspectionID int64) error { return nil }
func (f *fakeInspectionRepo) DeleteDisease(ctx context.Context, diseaseID, inspectionID int64) error {
	return nil
}
func (f *fakeInspectionRepo) GetByID(ctx context.Context, inspectionID, hiveID int64) (*model.Inspection, error) {
	if f.inspection == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.inspection, nil
}
func (f *fakeInspectionRepo) GetDiseaseByID(ctx context.Context, diseaseID, inspectionID int64) (*model.InspectionDisease, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeInspectionRepo) CountByHiveID(ctx context.Context, hiveID int64) (int64, error) {
	return 0, nil
}
func (f *fakeInspectionRepo) ListByHiveID(ctx context.Context, hiveID int64, limit, offset int) ([]*model.Inspection, error) {
	return nil, nil
}
func (f *fakeInspectionRepo) ListDiseasesByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionDisease, error) {
	return nil, nil
}
func (f *fakeInspectionRepo) LastInspectionDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error) {
	return nil, nil
}
func (f *fakeInspectionRepo) ListDiseasesByInspectionIDs(ctx context.Context, ids []int64) ([]*model.InspectionDisease, error) {
	return nil, nil
}
func (f *fakeInspectionRepo) Update(ctx context.Context, insp *model.Inspection) error { return nil }

// fakeInspectionImageRepo is a minimal service.InspectionImageRepository for handler tests.
type fakeInspectionImageRepo struct{}

func (f *fakeInspectionImageRepo) Create(ctx context.Context, img *model.InspectionImage) error {
	img.ID = 1
	return nil
}
func (f *fakeInspectionImageRepo) GetByID(ctx context.Context, imageID, inspectionID int64) (*model.InspectionImage, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeInspectionImageRepo) ListByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error) {
	return nil, nil
}
func (f *fakeInspectionImageRepo) ListByInspectionIDForCleanup(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error) {
	return nil, nil
}
func (f *fakeInspectionImageRepo) CountByInspectionIDs(ctx context.Context, ids []int64) (map[int64]int, error) {
	return nil, nil
}
func (f *fakeInspectionImageRepo) Delete(ctx context.Context, imageID int64) error { return nil }

// fakeListingReader is a minimal service.ListingImageReader for handler tests.
type fakeListingReader struct {
	listing *model.Listing
}

func (f *fakeListingReader) GetByID(ctx context.Context, id int64) (*model.Listing, error) {
	if f.listing == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.listing, nil
}

func (f *fakeListingReader) Update(ctx context.Context, l *model.Listing) error {
	return nil
}

// fakeListingImageStore is a minimal service.ListingImageStore for handler tests.
type fakeListingImageStore struct{}

func (f *fakeListingImageStore) CreateImage(ctx context.Context, img *model.ListingImage) error {
	img.ID = 1
	return nil
}
func (f *fakeListingImageStore) GetImageByID(ctx context.Context, imageID, listingID int64) (*model.ListingImage, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeListingImageStore) ListImagesByListingID(ctx context.Context, listingID int64) ([]model.ListingImage, error) {
	return nil, nil
}
func (f *fakeListingImageStore) DeleteImage(ctx context.Context, imageID int64) error { return nil }

// newMultipartUploadRequest builds a POST request with a single "image" multipart field of size bytes.
func newMultipartUploadRequest(t *testing.T, url, contentType string, size int) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="image"; filename="test.jpg"`)
	header.Set("Content-Type", contentType)
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(make([]byte, size))); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

// authedRequest attaches a valid Bearer token for userID to req.
func authedRequest(t *testing.T, req *http.Request, userID int64) *http.Request {
	t.Helper()
	tok, err := token.NewAccessToken(userID, testUploadAuthSecret, 5)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	return req
}

func decodeErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Code
}

func TestInspectionImageUpload_Success(t *testing.T) {
	svc := service.NewInspectionImageService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeInspectionHiveReader{hive: &model.Hive{ID: 10, ApiaryID: 1}},
		&fakeInspectionRepo{inspection: &model.Inspection{ID: 5, HiveID: 10}},
		&fakeInspectionImageRepo{},
		t.TempDir(),
	)
	h := NewInspectionImageHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newMultipartUploadRequest(t, "/apiaries/1/hives/10/inspections/5/images", "image/jpeg", 1024)
	req.SetPathValue("id", "1")
	req.SetPathValue("hiveId", "10")
	req.SetPathValue("inspectionId", "5")
	req = authedRequest(t, req, 1)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInspectionImageUpload_TooLarge(t *testing.T) {
	svc := service.NewInspectionImageService(
		&fakeApiaryMembershipReader{apiary: &model.Apiary{ID: 1}},
		&fakeInspectionHiveReader{hive: &model.Hive{ID: 10, ApiaryID: 1}},
		&fakeInspectionRepo{inspection: &model.Inspection{ID: 5, HiveID: 10}},
		&fakeInspectionImageRepo{},
		t.TempDir(),
	)
	h := NewInspectionImageHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newMultipartUploadRequest(t, "/apiaries/1/hives/10/inspections/5/images", "image/jpeg", 12*1024*1024)
	req.SetPathValue("id", "1")
	req.SetPathValue("hiveId", "10")
	req.SetPathValue("inspectionId", "5")
	req = authedRequest(t, req, 1)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "IMAGE_TOO_LARGE" {
		t.Errorf("expected code IMAGE_TOO_LARGE, got %q", code)
	}
}

func TestListingImageUpload_Success_Handler(t *testing.T) {
	svc := service.NewListingImageService(
		&fakeListingReader{listing: &model.Listing{ID: 5, UserID: 3}},
		&fakeListingImageStore{},
		t.TempDir(),
	)
	h := NewListingImageHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newMultipartUploadRequest(t, "/api/v1/listings/5/images", "image/jpeg", 1024)
	req.SetPathValue("id", "5")
	req = authedRequest(t, req, 3)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListingImageUpload_TooLarge(t *testing.T) {
	svc := service.NewListingImageService(
		&fakeListingReader{listing: &model.Listing{ID: 5, UserID: 3}},
		&fakeListingImageStore{},
		t.TempDir(),
	)
	h := NewListingImageHandler(svc)
	handler := middleware.Auth(testUploadAuthSecret)(http.HandlerFunc(h.Upload))

	req := newMultipartUploadRequest(t, "/api/v1/listings/5/images", "image/jpeg", 12*1024*1024)
	req.SetPathValue("id", "5")
	req = authedRequest(t, req, 3)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := decodeErrorCode(t, rec); code != "IMAGE_TOO_LARGE" {
		t.Errorf("expected code IMAGE_TOO_LARGE, got %q", code)
	}
}
