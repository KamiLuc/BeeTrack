package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

type mockHoneyBatchRepo struct {
	createdBatch        *model.HoneyBatch
	createdCertRequest  *model.HoneyBatchCertificationRequest
	nextID              int64
	err                 error

	byToken map[string]*model.HoneyBatch
	byID    map[int64]*model.HoneyBatch
	byUser  map[int64][]*model.HoneyBatch

	updatedGatheringDate    time.Time
	updatedAmountGrams      int64
	updatedProcessingMethod string
	updatedHoneyType        string
	updatedMetadataHash     string
	updatedLabPDFURL        string
	updatedPDFFilename      string
	updatedPDFFileHash      string
	deletedID               int64
}

func (m *mockHoneyBatchRepo) GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byID[id], nil
}

func (m *mockHoneyBatchRepo) CreateWithCertificationRequest(ctx context.Context, b *model.HoneyBatch, certRequest *model.HoneyBatchCertificationRequest) error {
	if m.err != nil {
		return m.err
	}
	if m.nextID == 0 {
		m.nextID = 1
	}
	b.ID = m.nextID
	b.MetadataHash = "computed"
	m.createdBatch = b
	if certRequest != nil {
		certRequest.BatchID = b.ID
	}
	m.createdCertRequest = certRequest
	return nil
}

func (m *mockHoneyBatchRepo) GetByVerificationToken(ctx context.Context, token string) (*model.HoneyBatch, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byToken[token], nil
}

func (m *mockHoneyBatchRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.HoneyBatch, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byUser[userID], nil
}

func (m *mockHoneyBatchRepo) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return int64(len(m.byUser[userID])), nil
}

func (m *mockHoneyBatchRepo) UpdateFields(ctx context.Context, batch *model.HoneyBatch) error {
	if m.err != nil {
		return m.err
	}
	m.updatedGatheringDate = batch.GatheringDate
	m.updatedAmountGrams = batch.AmountGrams
	m.updatedProcessingMethod = batch.ProcessingMethod
	m.updatedHoneyType = batch.HoneyType
	m.updatedMetadataHash = batch.MetadataHash
	m.updatedLabPDFURL = batch.LabPDFURL
	m.updatedPDFFilename = batch.PDFFilename
	m.updatedPDFFileHash = batch.PDFFileHash
	return nil
}

func (m *mockHoneyBatchRepo) SoftDelete(ctx context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	m.deletedID = id
	return nil
}

type mockHoneyBatchCertificationRepo struct {
	latest map[int64]*model.HoneyBatchCertification
	err    error
}

func (m *mockHoneyBatchCertificationRepo) GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.latest[batchID], nil
}

type mockHoneyBatchQRCodeRepo struct {
	byBatchID map[int64]*model.HoneyBatchQRCode
	created   *model.HoneyBatchQRCode
	nextID    int64
	err       error
}

func (m *mockHoneyBatchQRCodeRepo) GetByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchQRCode, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byBatchID[batchID], nil
}

func (m *mockHoneyBatchQRCodeRepo) Create(ctx context.Context, q *model.HoneyBatchQRCode) error {
	if m.err != nil {
		return m.err
	}
	m.nextID++
	q.ID = m.nextID
	m.created = q
	return nil
}

type mockHoneyBatchJobRepo struct {
	created *model.BlockchainJob
	err     error
	pending bool
}

func (m *mockHoneyBatchJobRepo) Create(ctx context.Context, j *model.BlockchainJob) error {
	if m.err != nil {
		return m.err
	}
	m.created = j
	return nil
}

func (m *mockHoneyBatchJobRepo) HasPendingJob(ctx context.Context, batchID int64) (bool, error) {
	return m.pending, nil
}

type mockCertRequestRepo struct {
	created *model.HoneyBatchCertificationRequest
	pending *model.HoneyBatchCertificationRequest
	latest  *model.HoneyBatchCertificationRequest
	err     error
}

func (m *mockCertRequestRepo) Create(ctx context.Context, req *model.HoneyBatchCertificationRequest) error {
	if m.err != nil {
		return m.err
	}
	m.created = req
	return nil
}

func (m *mockCertRequestRepo) GetPendingForBatch(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pending, nil
}

func (m *mockCertRequestRepo) GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertificationRequest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.latest, nil
}

const testAppURL = "https://app.example.com"
const testPDFHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd"

func newTestHoneyBatchService(t *testing.T) (*HoneyBatchService, *mockHoneyBatchRepo) {
	t.Helper()
	batches := &mockHoneyBatchRepo{}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	return NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir()), batches
}

func validCreateBatchRequest() CreateBatchRequest {
	return CreateBatchRequest{
		GatheringDate:    time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		AmountGrams:      15000,
		ProcessingMethod: "raw",
		HoneyType:        "Lipowy",
		PDFMimeType:      "application/pdf",
		PDFData:          []byte("%PDF-1.4 lab results"),
	}
}

func TestCreateBatch_NoJobByDefault(t *testing.T) {
	svc, repo := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()

	batch, err := svc.CreateBatch(context.Background(), 42, req)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	if batch.VerificationToken == "" {
		t.Error("expected a verification token to be generated")
	}
	if batch.PDFFileHash == "" {
		t.Error("expected the PDF to be hashed")
	}
	if batch.LabPDFURL == "" {
		t.Error("expected the PDF to be stored under a generated filename")
	}
	if repo.createdCertRequest != nil {
		t.Error("expected no certification request when RequestCertification is false")
	}
}

func TestCreateBatch_RequestCertification(t *testing.T) {
	svc, repo := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.RequestCertification = true

	batch, err := svc.CreateBatch(context.Background(), 42, req)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	if repo.createdCertRequest == nil {
		t.Fatal("expected a certification request when RequestCertification is true")
	}
	if repo.createdCertRequest.Status != model.CertificationRequestStatusPending {
		t.Errorf("expected request status pending, got %s", repo.createdCertRequest.Status)
	}
	if repo.createdCertRequest.BatchID != batch.ID {
		t.Errorf("expected request batch id %d, got %d", batch.ID, repo.createdCertRequest.BatchID)
	}
	if repo.createdCertRequest.RequestedBy != 42 {
		t.Errorf("expected requested_by 42, got %d", repo.createdCertRequest.RequestedBy)
	}
}

func TestCreateBatch_InvalidAmount(t *testing.T) {
	svc, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.AmountGrams = 0

	if _, err := svc.CreateBatch(context.Background(), 42, req); err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreateBatch_HoneyTypeRequired(t *testing.T) {
	svc, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.HoneyType = ""

	if _, err := svc.CreateBatch(context.Background(), 42, req); err != ErrHoneyTypeRequired {
		t.Errorf("expected ErrHoneyTypeRequired, got %v", err)
	}
}

func TestCreateBatch_InvalidProcessingMethod(t *testing.T) {
	svc, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.ProcessingMethod = "boiled"

	if _, err := svc.CreateBatch(context.Background(), 42, req); err != ErrInvalidProcessingMethod {
		t.Errorf("expected ErrInvalidProcessingMethod, got %v", err)
	}
}

func TestCreateBatch_PDFRequired(t *testing.T) {
	svc, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFData = nil
	req.RequestCertification = true

	if _, err := svc.CreateBatch(context.Background(), 42, req); err != ErrPDFRequired {
		t.Errorf("expected ErrPDFRequired, got %v", err)
	}
}

func TestCreateBatch_NoPDFAllowedWithoutCertification(t *testing.T) {
	svc, repo := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFData = nil
	req.PDFMimeType = ""

	batch, err := svc.CreateBatch(context.Background(), 42, req)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	if batch.PDFFileHash != "" || batch.LabPDFURL != "" {
		t.Errorf("expected no PDF stored, got hash=%q url=%q", batch.PDFFileHash, batch.LabPDFURL)
	}
	if repo.createdCertRequest != nil {
		t.Error("expected no certification request when RequestCertification is false")
	}
}

func TestCreateBatch_InvalidPDFType(t *testing.T) {
	svc, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFMimeType = "image/png"

	if _, err := svc.CreateBatch(context.Background(), 42, req); err != ErrInvalidPDFType {
		t.Errorf("expected ErrInvalidPDFType, got %v", err)
	}
}

func TestCreateBatch_PDFTooLarge(t *testing.T) {
	svc, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFData = make([]byte, maxLabPDFBytes+1)

	if _, err := svc.CreateBatch(context.Background(), 42, req); err != ErrPDFTooLarge {
		t.Errorf("expected ErrPDFTooLarge, got %v", err)
	}
}

func TestGetBatchWithVerification_NeverCertified(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1"}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}}
	certifications := &mockHoneyBatchCertificationRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	result, err := svc.GetBatchWithVerification(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("GetBatchWithVerification() error = %v", err)
	}
	if result.Batch != batch {
		t.Error("expected the returned batch to match")
	}
	if result.Certification != nil {
		t.Error("expected a nil certification for a never-certified batch")
	}
}

func TestGetBatchWithVerification_WithCertification(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1"}
	cert := &model.HoneyBatchCertification{ID: 5, BatchID: 1, Status: model.CertificationStatusConfirmed}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: cert}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	result, err := svc.GetBatchWithVerification(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("GetBatchWithVerification() error = %v", err)
	}
	if result.Certification != cert {
		t.Error("expected the latest certification to be returned")
	}
}

func TestGetBatchWithVerification_TokenNotFound(t *testing.T) {
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{}}
	certifications := &mockHoneyBatchCertificationRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatchWithVerification(context.Background(), "unknown-token"); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGetBatch_Owned(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, VerificationToken: "tok-1"}
	cert := &model.HoneyBatchCertification{ID: 5, BatchID: 1, Status: model.CertificationStatusQueued}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: cert}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	result, err := svc.GetBatch(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("GetBatch() error = %v", err)
	}
	if result.Batch != batch || result.Certification != cert {
		t.Error("expected the owned batch and its latest certification to be returned")
	}
}

func TestGetBatch_NotOwner(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatch(context.Background(), 999, 1); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound for a non-owner, got %v", err)
	}
}

func TestGetBatch_NotFound(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatch(context.Background(), 42, 999); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGetBatch_SynthesizesQueuedForPendingJob(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	jobs := &mockHoneyBatchJobRepo{pending: true}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, jobs, testAppURL, t.TempDir())

	result, err := svc.GetBatch(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("GetBatch() error = %v", err)
	}
	if result.Certification == nil || result.Certification.Status != model.CertificationStatusQueued {
		t.Errorf("expected a synthesized queued certification, got %+v", result.Certification)
	}
}

func TestListBatches(t *testing.T) {
	b1 := &model.HoneyBatch{ID: 1, UserID: 42}
	b2 := &model.HoneyBatch{ID: 2, UserID: 42}
	cert1 := &model.HoneyBatchCertification{ID: 10, BatchID: 1, Status: model.CertificationStatusConfirmed}
	batches := &mockHoneyBatchRepo{byUser: map[int64][]*model.HoneyBatch{42: {b1, b2}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: cert1}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	items, total, err := svc.ListBatches(context.Background(), 42, 20, 0)
	if err != nil {
		t.Fatalf("ListBatches() error = %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 batches, got total=%d len=%d", total, len(items))
	}
	if items[0].Certification != cert1 {
		t.Error("expected the first item's latest certification to be attached")
	}
	if items[1].Certification != nil {
		t.Error("expected a nil certification for a never-certified batch")
	}
}

func validUpdateBatchRequest() UpdateBatchRequest {
	return UpdateBatchRequest{
		GatheringDate:    time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		AmountGrams:      20000,
		ProcessingMethod: "filtered",
		HoneyType:        "Gryczany",
	}
}

func TestUpdateBatch_Owned(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, HoneyType: "Lipowy", AmountGrams: 1000, ProcessingMethod: "raw"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	updated, err := svc.UpdateBatch(context.Background(), 42, 1, validUpdateBatchRequest())
	if err != nil {
		t.Fatalf("UpdateBatch() error = %v", err)
	}
	if updated.HoneyType != "Gryczany" || updated.AmountGrams != 20000 || updated.ProcessingMethod != "filtered" {
		t.Errorf("expected updated fields to be applied, got %+v", updated)
	}
	if batches.updatedHoneyType != "Gryczany" || batches.updatedAmountGrams != 20000 || batches.updatedProcessingMethod != "filtered" {
		t.Errorf("expected repository update to be called with the new fields, got honeyType=%s amount=%d method=%s",
			batches.updatedHoneyType, batches.updatedAmountGrams, batches.updatedProcessingMethod)
	}
	if batches.updatedMetadataHash == "" {
		t.Error("expected metadata_hash to be recomputed")
	}
}

func TestUpdateBatch_KeepsExistingPDFWhenNoneProvided(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, LabPDFURL: "old.pdf", PDFFilename: "old.pdf", PDFFileHash: "oldhash"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	updated, err := svc.UpdateBatch(context.Background(), 42, 1, validUpdateBatchRequest())
	if err != nil {
		t.Fatalf("UpdateBatch() error = %v", err)
	}
	if updated.LabPDFURL != "old.pdf" || updated.PDFFilename != "old.pdf" || updated.PDFFileHash != "oldhash" {
		t.Errorf("expected existing PDF fields to be preserved, got %+v", updated)
	}
	if batches.updatedLabPDFURL != "old.pdf" || batches.updatedPDFFileHash != "oldhash" {
		t.Error("expected repository update to persist the unchanged PDF fields")
	}
}

func TestUpdateBatch_ReplacesPDF(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, LabPDFURL: "old.pdf", PDFFilename: "old.pdf", PDFFileHash: "oldhash"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.PDFData = []byte("%PDF-1.4 new lab results")
	req.PDFMimeType = "application/pdf"
	req.PDFFilename = "new.pdf"

	updated, err := svc.UpdateBatch(context.Background(), 42, 1, req)
	if err != nil {
		t.Fatalf("UpdateBatch() error = %v", err)
	}
	if updated.PDFFilename != "new.pdf" {
		t.Errorf("expected PDFFilename new.pdf, got %s", updated.PDFFilename)
	}
	if updated.PDFFileHash == "oldhash" || updated.PDFFileHash == "" {
		t.Errorf("expected a freshly computed PDF hash, got %s", updated.PDFFileHash)
	}
	if updated.LabPDFURL == "old.pdf" || updated.LabPDFURL == "" {
		t.Errorf("expected a new generated filename, got %s", updated.LabPDFURL)
	}
}

func TestUpdateBatch_RemovesPDF(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, LabPDFURL: "old.pdf", PDFFilename: "old.pdf", PDFFileHash: "oldhash"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.RemovePDF = true

	updated, err := svc.UpdateBatch(context.Background(), 42, 1, req)
	if err != nil {
		t.Fatalf("UpdateBatch() error = %v", err)
	}
	if updated.LabPDFURL != "" || updated.PDFFilename != "" || updated.PDFFileHash != "" {
		t.Errorf("expected PDF fields to be cleared, got %+v", updated)
	}
	if batches.updatedLabPDFURL != "" || batches.updatedPDFFileHash != "" {
		t.Error("expected repository update to persist the cleared PDF fields")
	}
}

func TestUpdateBatch_PDFDataTakesPrecedenceOverRemovePDF(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, LabPDFURL: "old.pdf", PDFFilename: "old.pdf", PDFFileHash: "oldhash"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.RemovePDF = true
	req.PDFData = []byte("%PDF-1.4 new lab results")
	req.PDFMimeType = "application/pdf"
	req.PDFFilename = "new.pdf"

	updated, err := svc.UpdateBatch(context.Background(), 42, 1, req)
	if err != nil {
		t.Fatalf("UpdateBatch() error = %v", err)
	}
	if updated.PDFFilename != "new.pdf" || updated.PDFFileHash == "" {
		t.Errorf("expected the new PDF to win over RemovePDF, got %+v", updated)
	}
}

func TestUpdateBatch_InvalidPDFType(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.PDFData = []byte("not a pdf")
	req.PDFMimeType = "image/png"

	if _, err := svc.UpdateBatch(context.Background(), 42, 1, req); err != ErrInvalidPDFType {
		t.Errorf("expected ErrInvalidPDFType, got %v", err)
	}
}

func TestUpdateBatch_PDFTooLarge(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.PDFData = make([]byte, maxLabPDFBytes+1)
	req.PDFMimeType = "application/pdf"

	if _, err := svc.UpdateBatch(context.Background(), 42, 1, req); err != ErrPDFTooLarge {
		t.Errorf("expected ErrPDFTooLarge, got %v", err)
	}
}

func TestUpdateBatch_NotOwner(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.UpdateBatch(context.Background(), 999, 1, validUpdateBatchRequest()); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestUpdateBatch_HoneyTypeRequired(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.HoneyType = ""
	if _, err := svc.UpdateBatch(context.Background(), 42, 1, req); err != ErrHoneyTypeRequired {
		t.Errorf("expected ErrHoneyTypeRequired, got %v", err)
	}
}

func TestUpdateBatch_InvalidAmount(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.AmountGrams = 0
	if _, err := svc.UpdateBatch(context.Background(), 42, 1, req); err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestUpdateBatch_InvalidProcessingMethod(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	req := validUpdateBatchRequest()
	req.ProcessingMethod = "boiled"
	if _, err := svc.UpdateBatch(context.Background(), 42, 1, req); err != ErrInvalidProcessingMethod {
		t.Errorf("expected ErrInvalidProcessingMethod, got %v", err)
	}
}

func TestUpdateBatch_LockedWhenCertified(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusConfirmed}}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.UpdateBatch(context.Background(), 42, 1, validUpdateBatchRequest()); err != ErrBatchLocked {
		t.Errorf("expected ErrBatchLocked, got %v", err)
	}
}

func TestUpdateBatch_LockedWhenFailedAttemptExists(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusFailed}}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.UpdateBatch(context.Background(), 42, 1, validUpdateBatchRequest()); err != ErrBatchLocked {
		t.Errorf("expected ErrBatchLocked, got %v", err)
	}
}

func TestUpdateBatch_LockedWhenJobPending(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	jobs := &mockHoneyBatchJobRepo{pending: true}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, jobs, testAppURL, t.TempDir())

	if _, err := svc.UpdateBatch(context.Background(), 42, 1, validUpdateBatchRequest()); err != ErrBatchLocked {
		t.Errorf("expected ErrBatchLocked, got %v", err)
	}
}

func TestUpdateBatch_LockedWhenCertificationRequestPending(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certRequests := &mockCertRequestRepo{pending: &model.HoneyBatchCertificationRequest{ID: 9, BatchID: 1, Status: model.CertificationRequestStatusPending}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.UpdateBatch(context.Background(), 42, 1, validUpdateBatchRequest()); err != ErrBatchLocked {
		t.Errorf("expected ErrBatchLocked, got %v", err)
	}
}

func TestDeleteBatch_Owned(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.DeleteBatch(context.Background(), 42, 1); err != nil {
		t.Fatalf("DeleteBatch() error = %v", err)
	}
	if batches.deletedID != 1 {
		t.Errorf("expected batch 1 to be soft-deleted, got %d", batches.deletedID)
	}
}

func TestDeleteBatch_NotOwner(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.DeleteBatch(context.Background(), 999, 1); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
	if batches.deletedID != 0 {
		t.Error("expected no delete to occur for a non-owner")
	}
}

func TestRetryCertification_NeverCertified(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, PDFFileHash: testPDFHash}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != nil {
		t.Fatalf("RetryCertification() error = %v", err)
	}
	if certRequests.created == nil || certRequests.created.BatchID != 1 {
		t.Fatalf("expected a certification request created for batch 1, got %+v", certRequests.created)
	}
	if certRequests.created.Status != model.CertificationRequestStatusPending {
		t.Errorf("expected request status pending, got %s", certRequests.created.Status)
	}
	if certRequests.created.RequestedBy != 42 {
		t.Errorf("expected requested_by 42, got %d", certRequests.created.RequestedBy)
	}
}

func TestRetryCertification_AfterFailed(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, PDFFileHash: testPDFHash}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusFailed}}}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, certifications, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != nil {
		t.Fatalf("RetryCertification() error = %v", err)
	}
	if certRequests.created == nil {
		t.Fatal("expected a certification request created after a failed attempt")
	}
}

func TestRetryCertification_AfterReverted(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, PDFFileHash: testPDFHash}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusReverted}}}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, certifications, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != nil {
		t.Fatalf("RetryCertification() error = %v", err)
	}
	if certRequests.created == nil {
		t.Fatal("expected a certification request created after a reverted attempt")
	}
}

func TestRetryCertification_AlreadyLive(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, PDFFileHash: testPDFHash}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusConfirmed}}}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, certifications, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != ErrBatchAlreadyCertified {
		t.Errorf("expected ErrBatchAlreadyCertified, got %v", err)
	}
	if certRequests.created != nil {
		t.Error("expected no request to be created when already live")
	}
}

func TestRetryCertification_AlreadyPendingJob(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, PDFFileHash: testPDFHash}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	jobs := &mockHoneyBatchJobRepo{pending: true}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, certRequests, &mockHoneyBatchQRCodeRepo{}, jobs, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != ErrBatchAlreadyCertified {
		t.Errorf("expected ErrBatchAlreadyCertified, got %v", err)
	}
	if certRequests.created != nil {
		t.Error("expected no request to be created while a job is still queued/in-flight")
	}
}

func TestRetryCertification_ReviewRequestAlreadyPending(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, PDFFileHash: testPDFHash}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certRequests := &mockCertRequestRepo{pending: &model.HoneyBatchCertificationRequest{ID: 9, BatchID: 1, Status: model.CertificationRequestStatusPending}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != ErrCertificationRequestPending {
		t.Errorf("expected ErrCertificationRequestPending, got %v", err)
	}
	if certRequests.created != nil {
		t.Error("expected no second request to be created while one is still pending review")
	}
}

func TestRetryCertification_NotOwner(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 999, 1); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
	if certRequests.created != nil {
		t.Error("expected no request to be created for a non-owner")
	}
}

func TestRetryCertification_NoPDF(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certRequests := &mockCertRequestRepo{}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, certRequests, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if err := svc.RetryCertification(context.Background(), 42, 1); err != ErrBatchHasNoPDF {
		t.Errorf("expected ErrBatchHasNoPDF, got %v", err)
	}
	if certRequests.created != nil {
		t.Error("expected no request to be created when the batch has no PDF")
	}
}

func TestGetBatchPDF_Owned(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42, LabPDFURL: "abc.pdf"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, "/data/pdfs")

	path, err := svc.GetBatchPDF(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("GetBatchPDF() error = %v", err)
	}
	if want := svc.FilePath("abc.pdf"); path != want {
		t.Errorf("GetBatchPDF() = %q, want %q", path, want)
	}
}

func TestGetBatchPDF_NotOwner(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatchPDF(context.Background(), 999, 1); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGetBatchPDF_NoPDF(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatchPDF(context.Background(), 42, 1); err != ErrBatchHasNoPDF {
		t.Errorf("expected ErrBatchHasNoPDF, got %v", err)
	}
}

func TestGetBatchPDFByToken_Confirmed(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1", LabPDFURL: "abc.pdf"}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusConfirmed}}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, "/data/pdfs")

	path, err := svc.GetBatchPDFByToken(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("GetBatchPDFByToken() error = %v", err)
	}
	if want := svc.FilePath("abc.pdf"); path != want {
		t.Errorf("GetBatchPDFByToken() = %q, want %q", path, want)
	}
}

func TestGetBatchPDFByToken_NotCertified(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1"}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatchPDFByToken(context.Background(), "tok-1"); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
}

func TestGetBatchPDFByToken_NotFound(t *testing.T) {
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatchPDFByToken(context.Background(), "unknown"); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGenerateQRCodeDataByToken_Confirmed(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1"}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}, byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusConfirmed}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	data, err := svc.GenerateQRCodeDataByToken(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("GenerateQRCodeDataByToken() error = %v", err)
	}
	want := testAppURL + "/verify/tok-1"
	if data != want {
		t.Errorf("GenerateQRCodeDataByToken() = %q, want %q", data, want)
	}
}

func TestGenerateQRCodeDataByToken_TokenNotFound(t *testing.T) {
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{}}
	svc := NewHoneyBatchService(batches, &mockHoneyBatchCertificationRepo{}, &mockCertRequestRepo{}, &mockHoneyBatchQRCodeRepo{}, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeDataByToken(context.Background(), "unknown"); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGenerateQRCodeData_FirstCallCreatesRow(t *testing.T) {
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusConfirmed}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	data, err := svc.GenerateQRCodeData(context.Background(), 1)
	if err != nil {
		t.Fatalf("GenerateQRCodeData() error = %v", err)
	}
	want := testAppURL + "/verify/tok-1"
	if data != want {
		t.Errorf("GenerateQRCodeData() = %q, want %q", data, want)
	}
	if qrCodes.created == nil || qrCodes.created.QRCodeData != want {
		t.Errorf("expected a QR code row to be persisted with data %q, got %+v", want, qrCodes.created)
	}
}

func TestGenerateQRCodeData_ReusesExistingRow(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	certifications := &mockHoneyBatchCertificationRepo{}
	existing := &model.HoneyBatchQRCode{ID: 9, BatchID: 1, QRCodeData: testAppURL + "/verify/tok-1"}
	qrCodes := &mockHoneyBatchQRCodeRepo{byBatchID: map[int64]*model.HoneyBatchQRCode{1: existing}}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	data, err := svc.GenerateQRCodeData(context.Background(), 1)
	if err != nil {
		t.Fatalf("GenerateQRCodeData() error = %v", err)
	}
	if data != existing.QRCodeData {
		t.Errorf("GenerateQRCodeData() = %q, want %q", data, existing.QRCodeData)
	}
	if qrCodes.created != nil {
		t.Error("expected no new row when one already exists")
	}
}

func TestGenerateQRCodeData_BatchNotFound(t *testing.T) {
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{}}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 999); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGenerateQRCodeData_NotCertifiedYet(t *testing.T) {
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusSubmitted}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
	if qrCodes.created != nil {
		t.Error("expected no row to be persisted for an uncertified batch")
	}
}

func TestGenerateQRCodeData_NeverCertified(t *testing.T) {
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
}

func TestGenerateQRCodeData_Reverted(t *testing.T) {
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusReverted}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
	if qrCodes.created != nil {
		t.Error("expected no row to be persisted for a reverted certification")
	}
}

func TestGenerateQRCodeData_CertificationRepoError(t *testing.T) {
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{err: errors.New("db down")}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGenerateQRCodeData_QRRepoLookupError(t *testing.T) {
	batches := &mockHoneyBatchRepo{}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{err: errors.New("db down")}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGenerateQRCodeData_BatchRepoError(t *testing.T) {
	batches := &mockHoneyBatchRepo{err: errors.New("db down")}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(batches, certifications, &mockCertRequestRepo{}, qrCodes, &mockHoneyBatchJobRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}
