package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

type mockHoneyBatchRepo struct {
	createdBatch *model.HoneyBatch
	createdJob   *model.BlockchainJob
	nextID       int64
	err          error

	byToken map[string]*model.HoneyBatch
	byID    map[int64]*model.HoneyBatch
	byUser  map[int64][]*model.HoneyBatch

	updatedHoneyType string
	deletedID        int64
}

func (m *mockHoneyBatchRepo) GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byID[id], nil
}

func (m *mockHoneyBatchRepo) CreateWithCertificationJob(ctx context.Context, b *model.HoneyBatch, job *model.BlockchainJob) error {
	if m.err != nil {
		return m.err
	}
	if m.nextID == 0 {
		m.nextID = 1
	}
	b.ID = m.nextID
	b.MetadataHash = "computed"
	m.createdBatch = b
	if job != nil {
		job.BatchID = b.ID
	}
	m.createdJob = job
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

func (m *mockHoneyBatchRepo) UpdateNotes(ctx context.Context, id int64, honeyType string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedHoneyType = honeyType
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

const testAppURL = "https://app.example.com"

func newTestHoneyBatchService(t *testing.T) (*HoneyBatchService, *mockApiaryMembershipReader, *mockHoneyBatchRepo) {
	t.Helper()
	apiaries := &mockApiaryMembershipReader{apiary: &model.Apiary{ID: 1}, role: "member"}
	batches := &mockHoneyBatchRepo{}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	return NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir()), apiaries, batches
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
	svc, _, repo := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()

	batch, err := svc.CreateBatch(context.Background(), 42, 1, req)
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
	if repo.createdJob != nil {
		t.Error("expected no blockchain job when RequestCertification is false")
	}
}

func TestCreateBatch_RequestCertification(t *testing.T) {
	svc, _, repo := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.RequestCertification = true

	batch, err := svc.CreateBatch(context.Background(), 42, 1, req)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	if repo.createdJob == nil {
		t.Fatal("expected a blockchain job when RequestCertification is true")
	}
	if repo.createdJob.Status != model.CertificationStatusQueued {
		t.Errorf("expected job status queued, got %s", repo.createdJob.Status)
	}
	if repo.createdJob.BatchID != batch.ID {
		t.Errorf("expected job batch id %d, got %d", batch.ID, repo.createdJob.BatchID)
	}
}

func TestCreateBatch_InvalidAmount(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.AmountGrams = 0

	if _, err := svc.CreateBatch(context.Background(), 42, 1, req); err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreateBatch_HoneyTypeRequired(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.HoneyType = ""

	if _, err := svc.CreateBatch(context.Background(), 42, 1, req); err != ErrHoneyTypeRequired {
		t.Errorf("expected ErrHoneyTypeRequired, got %v", err)
	}
}

func TestCreateBatch_InvalidProcessingMethod(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.ProcessingMethod = "boiled"

	if _, err := svc.CreateBatch(context.Background(), 42, 1, req); err != ErrInvalidProcessingMethod {
		t.Errorf("expected ErrInvalidProcessingMethod, got %v", err)
	}
}

func TestCreateBatch_PDFRequired(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFData = nil

	if _, err := svc.CreateBatch(context.Background(), 42, 1, req); err != ErrPDFRequired {
		t.Errorf("expected ErrPDFRequired, got %v", err)
	}
}

func TestCreateBatch_InvalidPDFType(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFMimeType = "image/png"

	if _, err := svc.CreateBatch(context.Background(), 42, 1, req); err != ErrInvalidPDFType {
		t.Errorf("expected ErrInvalidPDFType, got %v", err)
	}
}

func TestCreateBatch_PDFTooLarge(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()
	req.PDFData = make([]byte, maxLabPDFBytes+1)

	if _, err := svc.CreateBatch(context.Background(), 42, 1, req); err != ErrPDFTooLarge {
		t.Errorf("expected ErrPDFTooLarge, got %v", err)
	}
}

func TestCreateBatch_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHoneyBatchService(t)
	req := validCreateBatchRequest()

	if _, err := svc.CreateBatch(context.Background(), 42, 999, req); err != ErrApiaryNotFound {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestGetBatchWithVerification_NeverCertified(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1"}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}}
	certifications := &mockHoneyBatchCertificationRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

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
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, VerificationToken: "tok-1"}
	cert := &model.HoneyBatchCertification{ID: 5, BatchID: 1, Status: model.CertificationStatusConfirmed}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{"tok-1": batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: cert}}
	svc := NewHoneyBatchService(apiaries, batches, certifications, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	result, err := svc.GetBatchWithVerification(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("GetBatchWithVerification() error = %v", err)
	}
	if result.Certification != cert {
		t.Error("expected the latest certification to be returned")
	}
}

func TestGetBatchWithVerification_TokenNotFound(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byToken: map[string]*model.HoneyBatch{}}
	certifications := &mockHoneyBatchCertificationRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatchWithVerification(context.Background(), "unknown-token"); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGetBatch_Owned(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, UserID: 42, VerificationToken: "tok-1"}
	cert := &model.HoneyBatchCertification{ID: 5, BatchID: 1, Status: model.CertificationStatusQueued}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: cert}}
	svc := NewHoneyBatchService(apiaries, batches, certifications, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	result, err := svc.GetBatch(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("GetBatch() error = %v", err)
	}
	if result.Batch != batch || result.Certification != cert {
		t.Error("expected the owned batch and its latest certification to be returned")
	}
}

func TestGetBatch_NotOwner(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatch(context.Background(), 999, 1); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound for a non-owner, got %v", err)
	}
}

func TestGetBatch_NotFound(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if _, err := svc.GetBatch(context.Background(), 42, 999); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestListBatches(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	b1 := &model.HoneyBatch{ID: 1, UserID: 42}
	b2 := &model.HoneyBatch{ID: 2, UserID: 42}
	cert1 := &model.HoneyBatchCertification{ID: 10, BatchID: 1, Status: model.CertificationStatusConfirmed}
	batches := &mockHoneyBatchRepo{byUser: map[int64][]*model.HoneyBatch{42: {b1, b2}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: cert1}}
	svc := NewHoneyBatchService(apiaries, batches, certifications, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

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

func TestUpdateHoneyType_Owned(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, UserID: 42, HoneyType: "Lipowy"}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	updated, err := svc.UpdateHoneyType(context.Background(), 42, 1, "Gryczany")
	if err != nil {
		t.Fatalf("UpdateHoneyType() error = %v", err)
	}
	if updated.HoneyType != "Gryczany" {
		t.Errorf("expected honey_type Gryczany, got %s", updated.HoneyType)
	}
	if batches.updatedHoneyType != "Gryczany" {
		t.Errorf("expected repository update to be called with Gryczany, got %s", batches.updatedHoneyType)
	}
}

func TestUpdateHoneyType_NotOwner(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if _, err := svc.UpdateHoneyType(context.Background(), 999, 1, "Gryczany"); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestUpdateHoneyType_Required(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if _, err := svc.UpdateHoneyType(context.Background(), 42, 1, ""); err != ErrHoneyTypeRequired {
		t.Errorf("expected ErrHoneyTypeRequired, got %v", err)
	}
}

func TestDeleteBatch_Owned(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if err := svc.DeleteBatch(context.Background(), 42, 1); err != nil {
		t.Fatalf("DeleteBatch() error = %v", err)
	}
	if batches.deletedID != 1 {
		t.Errorf("expected batch 1 to be soft-deleted, got %d", batches.deletedID)
	}
}

func TestDeleteBatch_NotOwner(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: batch}}
	svc := NewHoneyBatchService(apiaries, batches, &mockHoneyBatchCertificationRepo{}, &mockHoneyBatchQRCodeRepo{}, testAppURL, t.TempDir())

	if err := svc.DeleteBatch(context.Background(), 999, 1); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
	if batches.deletedID != 0 {
		t.Error("expected no delete to occur for a non-owner")
	}
}

func TestGenerateQRCodeData_FirstCallCreatesRow(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusConfirmed}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

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
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{}
	certifications := &mockHoneyBatchCertificationRepo{}
	existing := &model.HoneyBatchQRCode{ID: 9, BatchID: 1, QRCodeData: testAppURL + "/verify/tok-1"}
	qrCodes := &mockHoneyBatchQRCodeRepo{byBatchID: map[int64]*model.HoneyBatchQRCode{1: existing}}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

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
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{}}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 999); err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGenerateQRCodeData_NotCertifiedYet(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusSubmitted}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
	if qrCodes.created != nil {
		t.Error("expected no row to be persisted for an uncertified batch")
	}
}

func TestGenerateQRCodeData_NeverCertified(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
}

func TestGenerateQRCodeData_Reverted(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{latest: map[int64]*model.HoneyBatchCertification{1: {Status: model.CertificationStatusReverted}}}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err != ErrBatchNotCertified {
		t.Errorf("expected ErrBatchNotCertified, got %v", err)
	}
	if qrCodes.created != nil {
		t.Error("expected no row to be persisted for a reverted certification")
	}
}

func TestGenerateQRCodeData_CertificationRepoError(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{1: {ID: 1, VerificationToken: "tok-1"}}}
	certifications := &mockHoneyBatchCertificationRepo{err: errors.New("db down")}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGenerateQRCodeData_QRRepoLookupError(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{err: errors.New("db down")}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGenerateQRCodeData_BatchRepoError(t *testing.T) {
	apiaries := &mockApiaryMembershipReader{}
	batches := &mockHoneyBatchRepo{err: errors.New("db down")}
	certifications := &mockHoneyBatchCertificationRepo{}
	qrCodes := &mockHoneyBatchQRCodeRepo{}
	svc := NewHoneyBatchService(apiaries, batches, certifications, qrCodes, testAppURL, t.TempDir())

	if _, err := svc.GenerateQRCodeData(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}
