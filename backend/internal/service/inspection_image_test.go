package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockInspectionImageRepo struct {
	images    []*model.InspectionImage
	image     *model.InspectionImage
	created   *model.InspectionImage
	deletedID int64
	counts    map[int64]int
}

func (m *mockInspectionImageRepo) Create(ctx context.Context, img *model.InspectionImage) error {
	img.ID = 1
	m.created = img
	return nil
}

func (m *mockInspectionImageRepo) GetByID(ctx context.Context, imageID, inspectionID int64) (*model.InspectionImage, error) {
	if m.image == nil || m.image.ID != imageID || m.image.InspectionID != inspectionID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.image, nil
}

func (m *mockInspectionImageRepo) ListByInspectionID(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error) {
	return m.images, nil
}

func (m *mockInspectionImageRepo) ListByInspectionIDForCleanup(ctx context.Context, inspectionID int64) ([]*model.InspectionImage, error) {
	return m.images, nil
}

func (m *mockInspectionImageRepo) CountByInspectionIDs(ctx context.Context, ids []int64) (map[int64]int, error) {
	if m.counts != nil {
		return m.counts, nil
	}
	return map[int64]int{}, nil
}

func (m *mockInspectionImageRepo) Delete(ctx context.Context, imageID int64) error {
	m.deletedID = imageID
	return nil
}

func newTestImageService(t *testing.T) (*InspectionImageService, *mockApiaryMembershipReader, *mockInspectionHiveReader, *mockInspectionRepo, *mockInspectionImageRepo, string) {
	t.Helper()
	dir := t.TempDir()
	apiaryMock := &mockApiaryMembershipReader{}
	hiveMock := &mockInspectionHiveReader{}
	inspMock := &mockInspectionRepo{}
	imgMock := &mockInspectionImageRepo{}
	svc := NewInspectionImageService(apiaryMock, hiveMock, inspMock, imgMock, dir)
	return svc, apiaryMock, hiveMock, inspMock, imgMock, dir
}

func TestUploadImage_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock, imgMock, _ := newTestImageService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	data := []byte{0xFF, 0xD8, 0xFF} // fake JPEG bytes
	img, err := svc.Upload(context.Background(), 1, 1, 10, 5, "image/jpeg", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if img.MimeType != "image/jpeg" {
		t.Errorf("unexpected mime type: %s", img.MimeType)
	}
	if imgMock.created == nil {
		t.Error("expected Create to be called")
	}
}

func TestUploadImage_InvalidType(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock, _, _ := newTestImageService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	_, err := svc.Upload(context.Background(), 1, 1, 10, 5, "image/gif", []byte{1})
	if !errors.Is(err, ErrInvalidImageType) {
		t.Errorf("expected ErrInvalidImageType, got %v", err)
	}
}

func TestUploadImage_TooLarge(t *testing.T) {
	svc, _, _, _, _, _ := newTestImageService(t)

	data := make([]byte, maxImageBytes+1)
	_, err := svc.Upload(context.Background(), 1, 1, 10, 5, "image/jpeg", data)
	if !errors.Is(err, ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge, got %v", err)
	}
}

func TestUploadImage_ApiaryNotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTestImageService(t)

	_, err := svc.Upload(context.Background(), 1, 1, 10, 5, "image/jpeg", []byte{1})
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock, imgMock, dir := newTestImageService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	filename := "testfile.jpg"
	_ = os.WriteFile(filepath.Join(dir, filename), []byte{1}, 0o644)
	imgMock.image = &model.InspectionImage{ID: 7, InspectionID: 5, Filename: filename, MimeType: "image/jpeg"}

	if err := svc.Delete(context.Background(), 1, 1, 10, 5, 7); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if imgMock.deletedID != 7 {
		t.Errorf("expected deletedID=7, got %d", imgMock.deletedID)
	}
	if _, err := os.Stat(filepath.Join(dir, filename)); !os.IsNotExist(err) {
		t.Error("expected file to be removed from disk")
	}
}

func TestDeleteImage_NotFound(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock, _, _ := newTestImageService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}

	err := svc.Delete(context.Background(), 1, 1, 10, 5, 99)
	if !errors.Is(err, ErrImageNotFound) {
		t.Errorf("expected ErrImageNotFound, got %v", err)
	}
}

func TestUploadImage_MaxImagesReached(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock, imgMock, _ := newTestImageService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}
	imgMock.images = make([]*model.InspectionImage, maxImagesPerInspection)

	_, err := svc.Upload(context.Background(), 1, 1, 10, 5, "image/jpeg", []byte{0xFF})
	if !errors.Is(err, ErrMaxImagesReached) {
		t.Errorf("expected ErrMaxImagesReached, got %v", err)
	}
}

func TestListImages_Success(t *testing.T) {
	svc, apiaryMock, hiveMock, inspMock, imgMock, _ := newTestImageService(t)
	apiaryMock.apiary = &model.Apiary{ID: 1}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	inspMock.inspection = &model.Inspection{ID: 5, HiveID: 10}
	imgMock.images = []*model.InspectionImage{
		{ID: 1, InspectionID: 5, Filename: "a.jpg", MimeType: "image/jpeg"},
		{ID: 2, InspectionID: 5, Filename: "b.png", MimeType: "image/png"},
	}

	imgs, err := svc.List(context.Background(), 1, 1, 10, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(imgs) != 2 {
		t.Errorf("expected 2 images, got %d", len(imgs))
	}
}

func TestCountForInspections_Success(t *testing.T) {
	svc, _, _, _, imgMock, _ := newTestImageService(t)
	imgMock.counts = map[int64]int{5: 3, 6: 0}

	counts, err := svc.CountForInspections(context.Background(), []int64{5, 6})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if counts[5] != 3 {
		t.Errorf("expected 3 photos for inspection 5, got %d", counts[5])
	}
	if counts[6] != 0 {
		t.Errorf("expected 0 photos for inspection 6, got %d", counts[6])
	}
}

func TestCountForInspections_Empty(t *testing.T) {
	svc, _, _, _, _, _ := newTestImageService(t)

	counts, err := svc.CountForInspections(context.Background(), []int64{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map, got %v", counts)
	}
}

func TestDeleteFilesForInspection(t *testing.T) {
	svc, _, _, _, imgMock, dir := newTestImageService(t)

	f1, f2 := "img1.jpg", "img2.jpg"
	_ = os.WriteFile(filepath.Join(dir, f1), []byte{1}, 0o644)
	_ = os.WriteFile(filepath.Join(dir, f2), []byte{2}, 0o644)
	imgMock.images = []*model.InspectionImage{
		{ID: 1, InspectionID: 5, Filename: f1},
		{ID: 2, InspectionID: 5, Filename: f2},
	}

	svc.DeleteFilesForInspection(context.Background(), 5)

	for _, f := range []string{f1, f2} {
		if _, err := os.Stat(filepath.Join(dir, f)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", f)
		}
	}
}
