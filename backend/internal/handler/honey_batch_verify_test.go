package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
)

func TestSanitizeFilenamePart(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercase word", "wildflower", "wildflower"},
		{"uppercase", "Wildflower", "wildflower"},
		{"spaces and punctuation", "Lipowy, miód!", "lipowy-mi-d"},
		{"leading and trailing unsafe chars", "  -Buckwheat- ", "buckwheat"},
		{"collapses runs of unsafe chars", "multi   flower--type", "multi-flower-type"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilenamePart(tt.in); got != tt.want {
				t.Errorf("sanitizeFilenamePart(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestQRCodeDownloadFilename(t *testing.T) {
	batch := &model.HoneyBatch{
		GatheringDate: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		HoneyType:     "Wildflower",
		AmountGrams:   1500,
	}

	got := qrCodeDownloadFilename(batch)
	want := "2024-05-01_wildflower_1.5kg.png"
	if got != want {
		t.Errorf("qrCodeDownloadFilename() = %q, want %q", got, want)
	}
}

func TestQRCodeDownloadFilename_WholeKilogramAndUnsafeHoneyType(t *testing.T) {
	batch := &model.HoneyBatch{
		GatheringDate: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		HoneyType:     "Rzepakowy!",
		AmountGrams:   2000,
	}

	got := qrCodeDownloadFilename(batch)
	want := "2023-12-25_rzepakowy_2kg.png"
	if got != want {
		t.Errorf("qrCodeDownloadFilename() = %q, want %q", got, want)
	}
}

type stubVerifyBatchRepo struct {
	batch *model.HoneyBatch
}

func (r *stubVerifyBatchRepo) CreateWithCertificationRequest(context.Context, *model.HoneyBatch, *model.HoneyBatchCertificationRequest) error {
	return nil
}
func (r *stubVerifyBatchRepo) GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	return r.batch, nil
}
func (r *stubVerifyBatchRepo) GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error) {
	return r.batch, nil
}
func (r *stubVerifyBatchRepo) ListByUserID(context.Context, int64, int, int) ([]*model.HoneyBatch, error) {
	return nil, nil
}
func (r *stubVerifyBatchRepo) CountByUserID(context.Context, int64) (int64, error)   { return 0, nil }
func (r *stubVerifyBatchRepo) UpdateFields(context.Context, *model.HoneyBatch) error { return nil }
func (r *stubVerifyBatchRepo) SoftDelete(context.Context, int64) error               { return nil }
func (r *stubVerifyBatchRepo) HardDelete(context.Context, int64) error               { return nil }

type stubVerifyCertRequestRepo struct{}

func (r *stubVerifyCertRequestRepo) Create(context.Context, *model.HoneyBatchCertificationRequest) error {
	return nil
}
func (r *stubVerifyCertRequestRepo) GetPendingForBatch(context.Context, int64) (*model.HoneyBatchCertificationRequest, error) {
	return nil, nil
}
func (r *stubVerifyCertRequestRepo) GetLatestByBatchID(context.Context, int64) (*model.HoneyBatchCertificationRequest, error) {
	return nil, nil
}

type stubVerifyCertRepo struct {
	cert *model.HoneyBatchCertification
}

func (r *stubVerifyCertRepo) GetLatestByBatchID(context.Context, int64) (*model.HoneyBatchCertification, error) {
	return r.cert, nil
}

type stubVerifyQRCodeRepo struct{}

func (r *stubVerifyQRCodeRepo) GetByBatchID(context.Context, int64) (*model.HoneyBatchQRCode, error) {
	return nil, nil
}
func (r *stubVerifyQRCodeRepo) Create(context.Context, *model.HoneyBatchQRCode) error { return nil }

type stubVerifyJobRepo struct{}

func (r *stubVerifyJobRepo) Create(context.Context, *model.BlockchainJob) error { return nil }
func (r *stubVerifyJobRepo) HasPendingJob(context.Context, int64) (bool, error) { return false, nil }

func TestQRCodeDownload_ContentDisposition(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:                1,
		VerificationToken: "tok-1",
		GatheringDate:     time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		HoneyType:         "Wildflower",
		AmountGrams:       1500,
	}
	cert := &model.HoneyBatchCertification{Status: model.CertificationStatusConfirmed}

	svc := service.NewHoneyBatchService(
		&stubVerifyBatchRepo{batch: batch},
		&stubVerifyCertRepo{cert: cert},
		&stubVerifyCertRequestRepo{},
		&stubVerifyQRCodeRepo{},
		&stubVerifyJobRepo{},
		"https://example.com",
		t.TempDir(),
	)
	h := NewHoneyBatchVerifyHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/verify/tok-1/qr-code/download", nil)
	req.SetPathValue("token", "tok-1")
	rec := httptest.NewRecorder()

	h.QRCodeDownload(rec, req)

	wantDisposition := `attachment; filename="2024-05-01_wildflower_1.5kg.png"`
	if got := rec.Header().Get("Content-Disposition"); got != wantDisposition {
		t.Errorf("Content-Disposition = %q, want %q", got, wantDisposition)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
