package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/beetrack/backend/internal/model"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newVoiceTestRepo(t *testing.T) (*VoiceRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return NewVoiceRepository(gdb), mock
}

func TestVoiceRepository_CreateRecording(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "voice_recordings"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	rec := &model.VoiceRecording{UserID: 7, ApiaryID: 3, Status: model.VoiceRecordingStatusPending}
	if err := repo.CreateRecording(context.Background(), rec); err != nil {
		t.Fatalf("CreateRecording returned error: %v", err)
	}
	if rec.ID != 1 {
		t.Fatalf("expected id 1, got %d", rec.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_CountRecordingsByUserID(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "voice_recordings" WHERE user_id = $1 AND status IN ($2,$3,$4)`)).
		WithArgs(int64(7),
			model.VoiceRecordingStatusPending, model.VoiceRecordingStatusProcessing, model.VoiceRecordingStatusCompleted).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.CountRecordingsByUserID(context.Background(), 7)
	if err != nil {
		t.Fatalf("CountRecordingsByUserID returned error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_GetRecordingByID_Found(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_recordings"`)).
		WithArgs(int64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "apiary_id", "status", "created_at"}).
			AddRow(1, 7, 3, model.VoiceRecordingStatusCompleted, now))

	rec, err := repo.GetRecordingByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetRecordingByID returned error: %v", err)
	}
	if rec == nil || rec.ID != 1 || rec.UserID != 7 || rec.ApiaryID != 3 || rec.Status != model.VoiceRecordingStatusCompleted {
		t.Fatalf("unexpected recording: %+v", rec)
	}
}

func TestVoiceRepository_GetRecordingByID_NotFound(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_recordings"`)).
		WithArgs(int64(99), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "apiary_id", "status", "created_at"}))

	rec, err := repo.GetRecordingByID(context.Background(), 99)
	if err != nil {
		t.Fatalf("GetRecordingByID returned error: %v", err)
	}
	if rec != nil {
		t.Fatalf("expected nil recording, got %+v", rec)
	}
}

func TestVoiceRepository_ListRecordingsByApiaryID(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_recordings" WHERE apiary_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(int64(3), 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "apiary_id", "status", "created_at"}).
			AddRow(2, 7, 3, model.VoiceRecordingStatusProcessing, now).
			AddRow(1, 7, 3, model.VoiceRecordingStatusCompleted, now.Add(-time.Hour)))

	recs, err := repo.ListRecordingsByApiaryID(context.Background(), 3, 20, 0)
	if err != nil {
		t.Fatalf("ListRecordingsByApiaryID returned error: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 recordings, got %d", len(recs))
	}
	if recs[0].ID != 2 || recs[1].ID != 1 {
		t.Fatalf("recordings not returned in expected order: %+v", recs)
	}
}

func TestVoiceRepository_CountRecordingsByApiaryID(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "voice_recordings" WHERE apiary_id = $1`)).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := repo.CountRecordingsByApiaryID(context.Background(), 3)
	if err != nil {
		t.Fatalf("CountRecordingsByApiaryID returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_ListActionsByRecordingIDs(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_actions" WHERE voice_recording_id IN ($1,$2) ORDER BY voice_recording_id ASC, sequence ASC`)).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "voice_recording_id", "sequence", "status", "created_at"}).
			AddRow(1, 1, 1, model.VoiceActionStatusApplied, now).
			AddRow(2, 2, 1, model.VoiceActionStatusProposed, now))

	actions, err := repo.ListActionsByRecordingIDs(context.Background(), []int64{1, 2})
	if err != nil {
		t.Fatalf("ListActionsByRecordingIDs returned error: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
}

func TestVoiceRepository_ListActionsByRecordingIDs_Empty(t *testing.T) {
	repo, _ := newVoiceTestRepo(t)

	actions, err := repo.ListActionsByRecordingIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListActionsByRecordingIDs returned error: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("expected 0 actions, got %d", len(actions))
	}
}

func TestVoiceRepository_UpdateRecording(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_recordings" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rec := &model.VoiceRecording{ID: 1, UserID: 7, ApiaryID: 3, Status: model.VoiceRecordingStatusCompleted}
	if err := repo.UpdateRecording(context.Background(), rec); err != nil {
		t.Fatalf("UpdateRecording returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_DeleteRecording(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "voice_recordings" WHERE "voice_recordings"."id" = $1`)).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.DeleteRecording(context.Background(), 1); err != nil {
		t.Fatalf("DeleteRecording returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_CreateLLMCall(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "voice_llm_calls"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	call := &model.VoiceLLMCall{
		VoiceRecordingID: 1,
		Phase:            model.VoiceLLMCallPhaseResolveHive,
		Request:          datatypes.JSON(`{"transcript":"hive three looked good"}`),
		Response:         datatypes.JSON(`{"outcome":"matched","hive_id":42}`),
	}
	if err := repo.CreateLLMCall(context.Background(), call); err != nil {
		t.Fatalf("CreateLLMCall returned error: %v", err)
	}
	if call.ID != 1 {
		t.Fatalf("expected id 1, got %d", call.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_CreateAction(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "voice_actions"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	action := &model.VoiceAction{
		VoiceRecordingID: 1,
		Sequence:         1,
		Status:           model.VoiceActionStatusProposed,
	}
	if err := repo.CreateAction(context.Background(), action); err != nil {
		t.Fatalf("CreateAction returned error: %v", err)
	}
	if action.ID != 1 {
		t.Fatalf("expected id 1, got %d", action.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_ListActionsByRecordingID(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_actions" WHERE voice_recording_id = $1 ORDER BY sequence ASC`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "voice_recording_id", "sequence", "status", "created_at"}).
			AddRow(1, 1, 1, model.VoiceActionStatusProposed, now).
			AddRow(2, 1, 2, model.VoiceActionStatusApplied, now.Add(time.Second)))

	actions, err := repo.ListActionsByRecordingID(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListActionsByRecordingID returned error: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].Sequence != 1 || actions[1].Sequence != 2 {
		t.Fatalf("actions not returned in expected order: %+v", actions)
	}
}

func TestVoiceRepository_ClaimNext_Found(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_recordings" WHERE status = $1 AND (next_attempt_at IS NULL OR next_attempt_at <= NOW()) ORDER BY created_at ASC,"voice_recordings"."id" LIMIT $2 FOR UPDATE SKIP LOCKED`)).
		WithArgs(model.VoiceRecordingStatusPending, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "apiary_id", "status", "created_at"}).
			AddRow(1, 7, 3, model.VoiceRecordingStatusPending, time.Now()))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_recordings" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rec, err := repo.ClaimNext(context.Background())
	if err != nil {
		t.Fatalf("ClaimNext returned error: %v", err)
	}
	if rec == nil || rec.ID != 1 || rec.Status != model.VoiceRecordingStatusProcessing {
		t.Fatalf("unexpected recording: %+v", rec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_ClaimNext_NoneAvailable(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_recordings" WHERE status = $1`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "apiary_id", "status", "created_at"}))
	mock.ExpectRollback()

	rec, err := repo.ClaimNext(context.Background())
	if err != nil {
		t.Fatalf("ClaimNext returned error: %v", err)
	}
	if rec != nil {
		t.Fatalf("expected nil recording, got %+v", rec)
	}
}

func TestVoiceRepository_MarkCompleted(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_recordings" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.MarkCompleted(context.Background(), 1, "hive three looked good", "en"); err != nil {
		t.Fatalf("MarkCompleted returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_MarkFailed(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_recordings" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.MarkFailed(context.Background(), 1, "POOR_AUDIO_QUALITY"); err != nil {
		t.Fatalf("MarkFailed returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_MarkRetry(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_recordings" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.MarkRetry(context.Background(), 1, "timeout", time.Now().Add(5*time.Second)); err != nil {
		t.Fatalf("MarkRetry returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_SweepStuckProcessing(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_recordings" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	n, err := repo.SweepStuckProcessing(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("SweepStuckProcessing returned error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 rows reset, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestVoiceRepository_UpdateAction(t *testing.T) {
	repo, mock := newVoiceTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "voice_actions" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	action := &model.VoiceAction{ID: 1, VoiceRecordingID: 1, Sequence: 1, Status: model.VoiceActionStatusApplied}
	if err := repo.UpdateAction(context.Background(), action); err != nil {
		t.Fatalf("UpdateAction returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
