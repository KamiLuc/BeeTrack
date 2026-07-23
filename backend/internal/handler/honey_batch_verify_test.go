package handler

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/blockchain"
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
	h := NewHoneyBatchVerifyHandler(svc, nil)

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

func newVerifyPageService(t *testing.T, batch *model.HoneyBatch, cert *model.HoneyBatchCertification) *service.HoneyBatchService {
	return service.NewHoneyBatchService(
		&stubVerifyBatchRepo{batch: batch},
		&stubVerifyCertRepo{cert: cert},
		&stubVerifyCertRequestRepo{},
		&stubVerifyQRCodeRepo{},
		&stubVerifyJobRepo{},
		"https://example.com",
		t.TempDir(),
	)
}

func TestVerifyPage_FoundAndConfirmed(t *testing.T) {
	txHash := "0xabc123"
	blockNum := int64(999)
	batch := &model.HoneyBatch{
		ID:                1,
		VerificationToken: "tok-1",
		GatheringDate:     time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		HoneyType:         "Wildflower",
		ProcessingMethod:  "raw",
		AmountGrams:       1500,
		PDFFileHash:       "pdfhash123",
		MetadataHash:      "metahash456",
	}
	cert := &model.HoneyBatchCertification{
		Status:          model.CertificationStatusConfirmed,
		ContractAddress: "0xcontract",
		BlockNumber:     &blockNum,
		TransactionHash: &txHash,
	}
	svc := newVerifyPageService(t, batch, cert)
	h := NewHoneyBatchVerifyHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-1", nil)
	req.SetPathValue("token", "tok-1")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{"Wildflower", "Raw", "pdfhash123", "metahash456", "0xabc123", "0xcontract", "999", "https://amoy.polygonscan.com/tx/0xabc123", "Confirmed"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q, body:\n%s", want, body)
		}
	}
}

func TestVerifyPage_FoundNotYetCertified(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:                1,
		VerificationToken: "tok-2",
		GatheringDate:     time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		HoneyType:         "Buckwheat",
		ProcessingMethod:  "filtered",
		AmountGrams:       1000,
	}
	svc := newVerifyPageService(t, batch, nil)
	h := NewHoneyBatchVerifyHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-2", nil)
	req.SetPathValue("token", "tok-2")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Not certified") {
		t.Errorf("expected body to contain %q, body:\n%s", "Not certified", body)
	}
	if strings.Contains(body, "Smart contract address") {
		t.Errorf("expected no certification proof fields for uncertified batch, body:\n%s", body)
	}
}

func TestVerifyPage_UnknownTokenReturns404(t *testing.T) {
	svc := newVerifyPageService(t, nil, nil)
	h := NewHoneyBatchVerifyHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/verify/unknown", nil)
	req.SetPathValue("token", "unknown")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), "invalid or the batch no longer exists") {
		t.Errorf("expected not-found message in body, got:\n%s", rec.Body.String())
	}
}

func TestVerifyPage_LanguageSelection(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:                1,
		VerificationToken: "tok-3",
		GatheringDate:     time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		HoneyType:         "Lipowy",
		ProcessingMethod:  "raw",
		AmountGrams:       1000,
	}

	tests := []struct {
		name           string
		url            string
		acceptLanguage string
		wantHTMLLang   string
		wantSnippet    string
	}{
		{"lang=pl query param", "/verify/tok-3?lang=pl", "", `lang="pl"`, "Metoda przetwarzania"},
		{"lang=en query param", "/verify/tok-3?lang=en", "", `lang="en"`, "Processing method"},
		{"default no header", "/verify/tok-3", "", `lang="en"`, "Processing method"},
		{"accept-language pl header", "/verify/tok-3", "pl-PL,pl;q=0.9", `lang="pl"`, "Metoda przetwarzania"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newVerifyPageService(t, batch, nil)
			h := NewHoneyBatchVerifyHandler(svc, nil)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.SetPathValue("token", "tok-3")
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}
			rec := httptest.NewRecorder()

			h.VerifyPage(rec, req)

			body := rec.Body.String()
			if !strings.Contains(body, tt.wantHTMLLang) {
				t.Errorf("expected body to contain %q, body:\n%s", tt.wantHTMLLang, body)
			}
			if !strings.Contains(body, tt.wantSnippet) {
				t.Errorf("expected body to contain %q, body:\n%s", tt.wantSnippet, body)
			}
		})
	}
}

func TestPickVerifyPageLanguage(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		acceptLanguage string
		want           string
	}{
		{"query param pl wins over header", "/verify/x?lang=pl", "en-US", "pl"},
		{"query param en wins over header", "/verify/x?lang=en", "pl-PL", "en"},
		{"header pl with no query", "/verify/x", "pl-PL,pl;q=0.9", "pl"},
		{"header en with no query", "/verify/x", "en-US,en;q=0.9", "en"},
		{"no header no query defaults english", "/verify/x", "", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}
			got := pickVerifyPageLanguage(req)
			if got.HTMLLang != tt.want {
				t.Errorf("pickVerifyPageLanguage() HTMLLang = %q, want %q", got.HTMLLang, tt.want)
			}
		})
	}
}

func TestFormatKg(t *testing.T) {
	tests := []struct {
		name  string
		grams int64
		want  string
	}{
		{"whole kilogram", 2000, "2"},
		{"fractional kilogram", 1500, "1.5"},
		{"zero", 0, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatKg(tt.grams); got != tt.want {
				t.Errorf("formatKg(%d) = %q, want %q", tt.grams, got, tt.want)
			}
		})
	}
}

func TestFormatInt64(t *testing.T) {
	if got := formatInt64(12345); got != "12345" {
		t.Errorf("formatInt64(12345) = %q, want %q", got, "12345")
	}
}

type stubChainCertReader struct {
	record *blockchain.CertificationRecord
	err    error
}

func (r *stubChainCertReader) GetCertification(ctx context.Context, batchID int64) (*blockchain.CertificationRecord, error) {
	return r.record, r.err
}

func confirmedVerifyBatchAndCert(t *testing.T) (*model.HoneyBatch, *model.HoneyBatchCertification, [32]byte, [32]byte) {
	var pdfHashBytes, metadataHashBytes [32]byte
	for i := range pdfHashBytes {
		pdfHashBytes[i] = byte(i)
	}
	for i := range metadataHashBytes {
		metadataHashBytes[i] = byte(i + 1)
	}
	batch := &model.HoneyBatch{
		ID:                1,
		VerificationToken: "tok-chain",
		GatheringDate:     time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		HoneyType:         "Wildflower",
		ProcessingMethod:  "raw",
		AmountGrams:       1500,
		PDFFileHash:       hex.EncodeToString(pdfHashBytes[:]),
		MetadataHash:      hex.EncodeToString(metadataHashBytes[:]),
	}
	cert := &model.HoneyBatchCertification{Status: model.CertificationStatusConfirmed}
	return batch, cert, pdfHashBytes, metadataHashBytes
}

func TestCheckAgainstChain_Match(t *testing.T) {
	batch, _, pdfHashBytes, metadataHashBytes := confirmedVerifyBatchAndCert(t)
	h := &HoneyBatchVerifyHandler{certReader: &stubChainCertReader{
		record: &blockchain.CertificationRecord{PDFHash: pdfHashBytes, MetadataHash: metadataHashBytes},
	}}

	pdf, metadata := h.checkAgainstChain(context.Background(), batch)

	if pdf != chainCheckMatch {
		t.Errorf("pdf state = %v, want %v", pdf, chainCheckMatch)
	}
	if metadata != chainCheckMatch {
		t.Errorf("metadata state = %v, want %v", metadata, chainCheckMatch)
	}
}

func TestCheckAgainstChain_Mismatch(t *testing.T) {
	batch, _, _, metadataHashBytes := confirmedVerifyBatchAndCert(t)
	var tamperedPDFHash [32]byte
	for i := range tamperedPDFHash {
		tamperedPDFHash[i] = byte(255 - i)
	}
	h := &HoneyBatchVerifyHandler{certReader: &stubChainCertReader{
		record: &blockchain.CertificationRecord{PDFHash: tamperedPDFHash, MetadataHash: metadataHashBytes},
	}}

	pdf, metadata := h.checkAgainstChain(context.Background(), batch)

	if pdf != chainCheckMismatch {
		t.Errorf("pdf state = %v, want %v", pdf, chainCheckMismatch)
	}
	if metadata != chainCheckMatch {
		t.Errorf("metadata state = %v, want %v", metadata, chainCheckMatch)
	}
}

func TestCheckAgainstChain_ReaderError(t *testing.T) {
	batch, _, _, _ := confirmedVerifyBatchAndCert(t)
	h := &HoneyBatchVerifyHandler{certReader: &stubChainCertReader{err: errors.New("rpc timeout")}}

	pdf, metadata := h.checkAgainstChain(context.Background(), batch)

	if pdf != chainCheckUnavailable {
		t.Errorf("pdf state = %v, want %v", pdf, chainCheckUnavailable)
	}
	if metadata != chainCheckUnavailable {
		t.Errorf("metadata state = %v, want %v", metadata, chainCheckUnavailable)
	}
}

func TestVerifyPage_ChainCheckMatchRendersVerifiedBadge(t *testing.T) {
	batch, cert, pdfHashBytes, metadataHashBytes := confirmedVerifyBatchAndCert(t)
	svc := newVerifyPageService(t, batch, cert)
	reader := &stubChainCertReader{
		record: &blockchain.CertificationRecord{PDFHash: pdfHashBytes, MetadataHash: metadataHashBytes},
	}
	h := NewHoneyBatchVerifyHandler(svc, reader)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-chain", nil)
	req.SetPathValue("token", "tok-chain")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Matches the record on the blockchain") {
		t.Errorf("expected match badge, body:\n%s", body)
	}
	if strings.Contains(body, "Does not match the blockchain record") || strings.Contains(body, "Live blockchain check unavailable") {
		t.Errorf("did not expect mismatch/unavailable badge, body:\n%s", body)
	}
}

func TestVerifyPage_ChainCheckMismatchRendersMismatchBadge(t *testing.T) {
	batch, cert, _, metadataHashBytes := confirmedVerifyBatchAndCert(t)
	var tamperedPDFHash [32]byte
	for i := range tamperedPDFHash {
		tamperedPDFHash[i] = byte(255 - i)
	}
	svc := newVerifyPageService(t, batch, cert)
	reader := &stubChainCertReader{
		record: &blockchain.CertificationRecord{PDFHash: tamperedPDFHash, MetadataHash: metadataHashBytes},
	}
	h := NewHoneyBatchVerifyHandler(svc, reader)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-chain", nil)
	req.SetPathValue("token", "tok-chain")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Does not match the blockchain record") {
		t.Errorf("expected mismatch badge, body:\n%s", body)
	}
}

func TestVerifyPage_ChainCheckErrorRendersUnavailableBadge(t *testing.T) {
	batch, cert, _, _ := confirmedVerifyBatchAndCert(t)
	svc := newVerifyPageService(t, batch, cert)
	reader := &stubChainCertReader{err: errors.New("rpc timeout")}
	h := NewHoneyBatchVerifyHandler(svc, reader)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-chain", nil)
	req.SetPathValue("token", "tok-chain")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Live blockchain check unavailable") {
		t.Errorf("expected unavailable badge, body:\n%s", body)
	}
}

func TestVerifyPage_NilCertReaderRendersUnavailableBadgeWhenConfirmed(t *testing.T) {
	batch, cert, _, _ := confirmedVerifyBatchAndCert(t)
	svc := newVerifyPageService(t, batch, cert)
	h := NewHoneyBatchVerifyHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-chain", nil)
	req.SetPathValue("token", "tok-chain")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Live blockchain check unavailable") {
		t.Errorf("expected unavailable badge with nil certReader, body:\n%s", body)
	}
}

func TestVerifyPage_ChainCheckSkippedWhenNotConfirmed(t *testing.T) {
	batch, _, pdfHashBytes, metadataHashBytes := confirmedVerifyBatchAndCert(t)
	cert := &model.HoneyBatchCertification{Status: model.CertificationStatusPendingConfirmation}
	svc := newVerifyPageService(t, batch, cert)
	reader := &stubChainCertReader{
		record: &blockchain.CertificationRecord{PDFHash: pdfHashBytes, MetadataHash: metadataHashBytes},
	}
	h := NewHoneyBatchVerifyHandler(svc, reader)

	req := httptest.NewRequest(http.MethodGet, "/verify/tok-chain", nil)
	req.SetPathValue("token", "tok-chain")
	rec := httptest.NewRecorder()

	h.VerifyPage(rec, req)

	body := rec.Body.String()
	for _, unwanted := range []string{`class="check`, "Matches the record on the blockchain", "Does not match the blockchain record", "Live blockchain check unavailable"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("expected no chain check badge markup for unconfirmed batch, found %q in body:\n%s", unwanted, body)
		}
	}
}

func TestProcessingMethodLabel(t *testing.T) {
	if got := processingMethodLabel(verifyPageEN, "raw"); got != "Raw" {
		t.Errorf("processingMethodLabel(EN, raw) = %q, want %q", got, "Raw")
	}
	if got := processingMethodLabel(verifyPagePL, "filtered"); got != "Filtrowany" {
		t.Errorf("processingMethodLabel(PL, filtered) = %q, want %q", got, "Filtrowany")
	}
	if got := processingMethodLabel(verifyPageEN, "unknown-method"); got != "unknown-method" {
		t.Errorf("processingMethodLabel(EN, unknown-method) = %q, want fallback to input", got)
	}
}
