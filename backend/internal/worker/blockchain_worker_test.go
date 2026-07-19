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
	pendingCalled    bool
	confirmedCalled  bool
	revertedCalled   bool
	failedErr        string
	failedNextRetry  time.Time
	pendingList      []*model.BlockchainJob
	sweptCalled      bool
	sweptCount       int64
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
func (m *mockJobRepo) MarkPendingConfirmation(ctx context.Context, id int64) error {
	m.pendingCalled = true
	return nil
}
func (m *mockJobRepo) MarkConfirmed(ctx context.Context, id int64) error {
	m.confirmedCalled = true
	return nil
}
func (m *mockJobRepo) MarkReverted(ctx context.Context, id int64) error {
	m.revertedCalled = true
	return nil
}
func (m *mockJobRepo) MarkFailed(ctx context.Context, id int64, lastErr string, nextRetryAt time.Time) error {
	m.failedErr = lastErr
	m.failedNextRetry = nextRetryAt
	return nil
}
func (m *mockJobRepo) ListPendingConfirmation(ctx context.Context) ([]*model.BlockchainJob, error) {
	return m.pendingList, nil
}
func (m *mockJobRepo) SweepStuckSubmitting(ctx context.Context, olderThan time.Duration) (int64, error) {
	m.sweptCalled = true
	return m.sweptCount, nil
}

type mockCertRepo struct {
	latest  *model.HoneyBatchCertification
	byID    map[int64]*model.HoneyBatchCertification
	nextID  int64
	created *model.HoneyBatchCertification
	updates []statusUpdate
}

type statusUpdate struct {
	id     int64
	status model.CertificationStatus
	fields map[string]any
}

func (m *mockCertRepo) GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertification, error) {
	if m.byID == nil {
		return nil, nil
	}
	return m.byID[id], nil
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

	mined         bool
	reverted      bool
	blockNumber   uint64
	gasUsed       uint64
	confirmations uint64
	statusErr     error
}

func (m *mockReader) GetCertification(ctx context.Context, batchID int64) (*blockchain.CertificationRecord, error) {
	return m.record, m.err
}

func (m *mockReader) GetTransactionStatus(ctx context.Context, txHash string) (mined, reverted bool, blockNumber, gasUsed, confirmations uint64, err error) {
	return m.mined, m.reverted, m.blockNumber, m.gasUsed, m.confirmations, m.statusErr
}

func newTestBatch() *model.HoneyBatch {
	hash := "abcd000000000000000000000000000000000000000000000000000000000000"[:64]
	return &model.HoneyBatch{ID: 7, PDFFileHash: hash, MetadataHash: hash}
}

func newWorker(jobs JobRepository, certs CertificationRepository, batches BatchReader, writer CertifyWriter, reader CertificationReader) *BlockchainWorker {
	return NewBlockchainWorker(jobs, certs, batches, writer, reader, 80002, "0xabc", 12)
}

func TestProcessNextJob_NoJobAvailable(t *testing.T) {
	w := newWorker(&mockJobRepo{}, &mockCertRepo{}, &mockBatchReader{}, &mockWriter{}, &mockReader{})

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
	w := newWorker(jobs, certs, batches, writer, &mockReader{})

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
	w := newWorker(jobs, certs, &mockBatchReader{batch: newTestBatch()}, writer, &mockReader{})

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
	w := newWorker(jobs, certs, &mockBatchReader{batch: newTestBatch()}, writer, reader)

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
	w := newWorker(jobs, certs, &mockBatchReader{batch: newTestBatch()}, writer, &mockReader{})

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
	w := newWorker(jobs, &mockCertRepo{}, &mockBatchReader{batch: newTestBatch()}, writer, &mockReader{})

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
	w := newWorker(jobs, certs, &mockBatchReader{batch: nil}, &mockWriter{}, &mockReader{})

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

func txHash(s string) *string { return &s }

func TestPollSubmittedJobs_NotYetMined(t *testing.T) {
	job := &model.BlockchainJob{ID: 1, BatchID: 7, CertificationID: idPtr(5)}
	jobs := &mockJobRepo{pendingList: []*model.BlockchainJob{job}}
	certs := &mockCertRepo{byID: map[int64]*model.HoneyBatchCertification{5: {ID: 5, TransactionHash: txHash("0xabc")}}}
	reader := &mockReader{mined: false}
	w := newWorker(jobs, certs, &mockBatchReader{}, &mockWriter{}, reader)

	if err := w.PollSubmittedJobs(context.Background()); err != nil {
		t.Fatalf("PollSubmittedJobs() error = %v", err)
	}
	if len(certs.updates) != 0 {
		t.Errorf("expected no updates while unmined, got %+v", certs.updates)
	}
}

func TestPollSubmittedJobs_MinedUnderConfirmations(t *testing.T) {
	job := &model.BlockchainJob{ID: 1, BatchID: 7, CertificationID: idPtr(5)}
	jobs := &mockJobRepo{pendingList: []*model.BlockchainJob{job}}
	certs := &mockCertRepo{byID: map[int64]*model.HoneyBatchCertification{5: {ID: 5, Status: model.CertificationStatusSubmitted, TransactionHash: txHash("0xabc")}}}
	reader := &mockReader{mined: true, confirmations: 3}
	w := newWorker(jobs, certs, &mockBatchReader{}, &mockWriter{}, reader)

	if err := w.PollSubmittedJobs(context.Background()); err != nil {
		t.Fatalf("PollSubmittedJobs() error = %v", err)
	}
	if !jobs.pendingCalled {
		t.Error("expected job marked pending_confirmation")
	}
	if len(certs.updates) != 1 || certs.updates[0].status != model.CertificationStatusPendingConfirmation {
		t.Errorf("expected certification marked pending_confirmation, got %+v", certs.updates)
	}
}

func TestPollSubmittedJobs_ConfirmedAfterEnoughConfirmations(t *testing.T) {
	job := &model.BlockchainJob{ID: 1, BatchID: 7, CertificationID: idPtr(5)}
	jobs := &mockJobRepo{pendingList: []*model.BlockchainJob{job}}
	certs := &mockCertRepo{byID: map[int64]*model.HoneyBatchCertification{5: {ID: 5, TransactionHash: txHash("0xabc")}}}
	reader := &mockReader{mined: true, confirmations: 12, blockNumber: 999, gasUsed: 21000}
	w := newWorker(jobs, certs, &mockBatchReader{}, &mockWriter{}, reader)

	if err := w.PollSubmittedJobs(context.Background()); err != nil {
		t.Fatalf("PollSubmittedJobs() error = %v", err)
	}
	if !jobs.confirmedCalled {
		t.Error("expected job marked confirmed")
	}
	if len(certs.updates) != 1 || certs.updates[0].status != model.CertificationStatusConfirmed {
		t.Errorf("expected certification marked confirmed, got %+v", certs.updates)
	}
	if certs.updates[0].fields["block_number"] != int64(999) {
		t.Errorf("expected block_number 999, got %+v", certs.updates[0].fields)
	}
}

func TestPollSubmittedJobs_Reverted(t *testing.T) {
	job := &model.BlockchainJob{ID: 1, BatchID: 7, CertificationID: idPtr(5)}
	jobs := &mockJobRepo{pendingList: []*model.BlockchainJob{job}}
	certs := &mockCertRepo{byID: map[int64]*model.HoneyBatchCertification{5: {ID: 5, TransactionHash: txHash("0xabc")}}}
	reader := &mockReader{mined: true, reverted: true}
	w := newWorker(jobs, certs, &mockBatchReader{}, &mockWriter{}, reader)

	if err := w.PollSubmittedJobs(context.Background()); err != nil {
		t.Fatalf("PollSubmittedJobs() error = %v", err)
	}
	if !jobs.revertedCalled {
		t.Error("expected job marked reverted")
	}
	if len(certs.updates) != 1 || certs.updates[0].status != model.CertificationStatusReverted {
		t.Errorf("expected certification marked reverted, got %+v", certs.updates)
	}
}

func TestPollSubmittedJobs_MissingCertificationIDCollectsError(t *testing.T) {
	job := &model.BlockchainJob{ID: 1, BatchID: 7}
	jobs := &mockJobRepo{pendingList: []*model.BlockchainJob{job}}
	w := newWorker(jobs, &mockCertRepo{}, &mockBatchReader{}, &mockWriter{}, &mockReader{})

	if err := w.PollSubmittedJobs(context.Background()); err == nil {
		t.Fatal("expected an error for a job with no certification id")
	}
}

// crashCertRepo is a CertificationRepository fake that keeps every row ever
// created (in creation order, like the append-only DB table) and can be told
// to fail one specific UpdateStatus call, leaving that row stuck mid-status
// — used to reproduce a worker crashing between broadcasting a transaction
// and recording it (HC-BE-25).
type crashCertRepo struct {
	rows       []*model.HoneyBatchCertification
	nextID     int64
	failUpdate int // 1-based UpdateStatus call number to fail; 0 = never
	updateCall int
}

func (r *crashCertRepo) GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertification, error) {
	for _, c := range r.rows {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (r *crashCertRepo) GetLatestByBatchID(ctx context.Context, batchID int64) (*model.HoneyBatchCertification, error) {
	var latest *model.HoneyBatchCertification
	for _, c := range r.rows {
		if c.BatchID == batchID {
			latest = c
		}
	}
	return latest, nil
}

func (r *crashCertRepo) Create(ctx context.Context, c *model.HoneyBatchCertification) error {
	r.nextID++
	c.ID = r.nextID
	r.rows = append(r.rows, c)
	return nil
}

func (r *crashCertRepo) UpdateStatus(ctx context.Context, id int64, status model.CertificationStatus, fields map[string]any) error {
	r.updateCall++
	if r.failUpdate != 0 && r.updateCall == r.failUpdate {
		return errors.New("simulated crash: connection lost before recording tx")
	}
	for _, c := range r.rows {
		if c.ID == id {
			c.Status = status
			if txHash, ok := fields["transaction_hash"].(string); ok {
				c.TransactionHash = &txHash
			}
		}
	}
	return nil
}

// TestProcessNextJob_MidBroadcastCrashRecovery reproduces HC-BE-25's
// idempotency guarantee: a worker that dies right after broadcasting a
// certify() transaction, but before recording the tx hash, must never let a
// retried job create more than one live certification. The transaction from
// the first attempt already succeeded on-chain, so the retry's own
// certify() call reverts as already-certified — exercising idempotency
// layers 1 (contract revert) and the resulting confirm-in-place handling
// together.
func TestProcessNextJob_MidBroadcastCrashRecovery(t *testing.T) {
	certs := &crashCertRepo{failUpdate: 1}
	jobs := &mockJobRepo{next: &model.BlockchainJob{ID: 1, BatchID: 7}}
	batches := &mockBatchReader{batch: newTestBatch()}
	w := newWorker(jobs, certs, batches, &mockWriter{txHash: "0xfirst"}, &mockReader{})

	if _, err := w.ProcessNextJob(context.Background()); err == nil {
		t.Fatal("expected the simulated post-broadcast crash to surface as an error")
	}
	if jobs.submittedCalled {
		t.Error("job should not be marked submitted when recording the tx crashes")
	}

	// Recovery: SweepStuckSubmitting (exercised separately) would reset the
	// stuck job back to queued; here we just simulate ClaimNext returning it
	// again for retry.
	jobs.next = &model.BlockchainJob{ID: 1, BatchID: 7, AttemptCount: 1}
	reader := &mockReader{record: &blockchain.CertificationRecord{Timestamp: time.Unix(2000, 0)}}
	w2 := newWorker(jobs, certs, batches, &mockWriter{err: blockchain.ErrAlreadyCertified}, reader)

	processed, err := w2.ProcessNextJob(context.Background())
	if err != nil {
		t.Fatalf("ProcessNextJob() retry error = %v", err)
	}
	if !processed {
		t.Fatal("expected the retried job to be processed")
	}

	if len(certs.rows) != 2 {
		t.Fatalf("expected two certification attempts recorded (one orphaned, one confirmed), got %d", len(certs.rows))
	}
	live := 0
	for _, c := range certs.rows {
		if c.Status.IsLive() {
			live++
		}
	}
	if live != 1 {
		t.Errorf("expected exactly one live certification after crash+recovery, got %d (rows: %+v)", live, certs.rows)
	}
}

func idPtr(id int64) *int64 { return &id }

func TestRun_TicksAndStopsOnContextCancel(t *testing.T) {
	jobs := &mockJobRepo{}
	certs := &mockCertRepo{}
	w := newWorker(jobs, certs, &mockBatchReader{}, &mockWriter{}, &mockReader{})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		w.Run(ctx, 10*time.Millisecond, 15*time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}

	if !jobs.sweptCalled {
		t.Error("expected the job loop to have ticked at least once")
	}
}
