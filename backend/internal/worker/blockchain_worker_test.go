package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/model"
)

type mockJobRepo struct {
	next             *model.BlockchainJob
	claimErr         error
	submittingCertID *int64
	submittedCalled  bool
	confirmedCalled  bool
	failedErr        string
	failedNextRetry  time.Time
}

func (m *mockJobRepo) ClaimNext(ctx context.Context) (*model.BlockchainJob, error) {
	return m.next, m.claimErr
}
func (m *mockJobRepo) MarkSubmitting(ctx context.Context, id, certificationID int64) error {
	m.submittingCertID = &certificationID
	return nil
}
func (m *mockJobRepo) MarkSubmitted(ctx context.Context, id int64) error {
	m.submittedCalled = true
	return nil
}
func (m *mockJobRepo) MarkConfirmed(ctx context.Context, id int64) error {
	m.confirmedCalled = true
	return nil
}
func (m *mockJobRepo) MarkFailed(ctx context.Context, id int64, lastErr string, nextRetryAt time.Time) error {
	m.failedErr = lastErr
	m.failedNextRetry = nextRetryAt
	return nil
}

type mockCertRepo struct {
	latest  *model.HoneyBatchCertification
	nextID  int64
	created *model.HoneyBatchCertification
	updates []statusUpdate
}

type statusUpdate struct {
	id     int64
	status model.CertificationStatus
	fields map[string]any
}

func (m *mockCertRepo) GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error) {
	return m.latest, nil
}
func (m *mockCertRepo) Create(ctx context.Context, c *model.HoneyBatchCertification) error {
	m.nextID++
	c.ID = m.nextID
	m.created = c
	return nil
}
func (m *mockCertRepo) UpdateStatus(ctx context.Context, id int64, status model.CertificationStatus, fields map[string]any) error {
	m.updates = append(m.updates, statusUpdate{id: id, status: status, fields: fields})
	return nil
}

type mockBatchReader struct {
	batch *model.HoneyBatch
}

func (m *mockBatchReader) GetByIDIgnoringDeletion(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	return m.batch, nil
}

type mockWriter struct {
	txHash string
	err    error
}

func (m *mockWriter) CertifyBatch(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (string, error) {
	return m.txHash, m.err
}

type mockReader struct {
	record *blockchain.CertificationRecord
	err    error
}

func (m *mockReader) GetCertification(ctx context.Context, batchID int64) (*blockchain.CertificationRecord, error) {
	return m.record, m.err
}

func newTestBatch() *model.HoneyBatch {
	hash := "abcd000000000000000000000000000000000000000000000000000000000000"[:64]
	return &model.HoneyBatch{ID: 7, PDFFileHash: hash, MetadataHash: hash}
}

func TestProcessNextJob_NoJobAvailable(t *testing.T) {
	w := NewBlockchainWorker(&mockJobRepo{}, &mockCertRepo{}, &mockBatchReader{}, &mockWriter{}, &mockReader{}, 80002, "0xabc")

	processed, err := w.ProcessNextJob(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextJob() error = %v", err)
	}
	if processed {
		t.Error("expected processed=false when no job is claimable")
	}
}

func TestProcessNextJob_HappyPath(t *testing.T) {
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7}}
	certs := &mockCertRepo{}
	batches := &mockBatchReader{batch: newTestBatch()}
	writer := &mockWriter{txHash: "0xdeadbeef"}
	w := NewBlockchainWorker(jobs, certs, batches, writer, &mockReader{}, 80002, "0xabc")

	processed, err := w.ProcessNextJob(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextJob() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if certs.created == nil {
		t.Fatal("expected a certification row to be created")
	}
	if !jobs.submittedCalled {
		t.Error("expected job to be marked submitted")
	}
	if len(certs.updates) != 1 || certs.updates[0].status != model.CertificationStatusSubmitted {
		t.Errorf("expected certification updated to submitted, got %+v", certs.updates)
	}
	if certs.updates[0].fields["transaction_hash"] != "0xdeadbeef" {
		t.Errorf("expected transaction_hash to be recorded, got %+v", certs.updates[0].fields)
	}
}

func TestProcessNextJob_SkipsWhenAlreadyLive(t *testing.T) {
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7}}
	certs := &mockCertRepo{latest: &model.HoneyBatchCertification{Status: model.CertificationStatusConfirmed}}
	writer := &mockWriter{}
	w := NewBlockchainWorker(jobs, certs, &mockBatchReader{batch: newTestBatch()}, writer, &mockReader{}, 80002, "0xabc")

	processed, err := w.ProcessNextJob(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextJob() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if !jobs.confirmedCalled {
		t.Error("expected job to be marked confirmed without resubmission")
	}
	if certs.created != nil {
		t.Error("expected no new certification row when a live one already exists")
	}
}

func TestProcessNextJob_AlreadyCertifiedRevert(t *testing.T) {
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7}}
	certs := &mockCertRepo{}
	writer := &mockWriter{err: blockchain.ErrAlreadyCertified}
	reader := &mockReader{record: &blockchain.CertificationRecord{Timestamp: time.Unix(1000, 0)}}
	w := NewBlockchainWorker(jobs, certs, &mockBatchReader{batch: newTestBatch()}, writer, reader, 80002, "0xabc")

	processed, err := w.ProcessNextJob(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextJob() error = %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if !jobs.confirmedCalled {
		t.Error("expected job to be marked confirmed on AlreadyCertified revert")
	}
	if len(certs.updates) != 1 || certs.updates[0].status != model.CertificationStatusConfirmed {
		t.Errorf("expected certification marked confirmed, got %+v", certs.updates)
	}
}

func TestProcessNextJob_WriterFailureSchedulesRetry(t *testing.T) {
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7, AttemptCount: 0}}
	certs := &mockCertRepo{}
	writer := &mockWriter{err: errors.New("rpc timeout")}
	w := NewBlockchainWorker(jobs, certs, &mockBatchReader{batch: newTestBatch()}, writer, &mockReader{}, 80002, "0xabc")

	if _, err := w.ProcessNextJob(context.Background()); err == nil {
		t.Fatal("expected ProcessNextJob to surface the writer error")
	}
	if jobs.failedErr != "rpc timeout" {
		t.Errorf("expected last_error to be recorded, got %q", jobs.failedErr)
	}
	if !jobs.failedNextRetry.After(time.Now()) || jobs.failedNextRetry.After(time.Now().Add(2*time.Second)) {
		t.Errorf("expected next retry ~1s out for attempt 1, got %v", jobs.failedNextRetry)
	}
	if len(certs.updates) != 1 || certs.updates[0].status != model.CertificationStatusFailed {
		t.Errorf("expected certification marked failed, got %+v", certs.updates)
	}
}

func TestProcessNextJob_ExhaustedAttemptsGoesFarFuture(t *testing.T) {
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7, AttemptCount: maxAttempts - 1}}
	writer := &mockWriter{err: errors.New("still failing")}
	w := NewBlockchainWorker(jobs, &mockCertRepo{}, &mockBatchReader{batch: newTestBatch()}, writer, &mockReader{}, 80002, "0xabc")

	if _, err := w.ProcessNextJob(context.Background()); err == nil {
		t.Fatal("expected an error")
	}
	if jobs.failedNextRetry.Before(time.Now().AddDate(50, 0, 0)) {
		t.Errorf("expected a far-future next retry after exhausting attempts, got %v", jobs.failedNextRetry)
	}
}

func TestProcessNextJob_BatchNotFound(t *testing.T) {
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7}}
	certs := &mockCertRepo{}
	w := NewBlockchainWorker(jobs, certs, &mockBatchReader{batch: nil}, &mockWriter{}, &mockReader{}, 80002, "0xabc")

	if _, err := w.ProcessNextJob(context.Background()); err == nil {
		t.Fatal("expected an error when the batch is missing")
	}
	if certs.created != nil {
		t.Error("expected no certification row to be created when the batch is missing")
	}
}

func TestBackoffDuration(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{10, 8 * time.Second},
	}
	for _, tt := range tests {
		if got := backoffDuration(tt.attempt); got != tt.want {
			t.Errorf("backoffDuration(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}
