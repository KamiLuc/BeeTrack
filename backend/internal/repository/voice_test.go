package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/beetrack/backend/internal/model"
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
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "voice_recordings" WHERE user_id = $1`)).
		WithArgs(int64(7)).
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
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "voice_recordings" WHERE apiary_id = $1 ORDER BY created_at DESC`)).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "apiary_id", "status", "created_at"}).
			AddRow(2, 7, 3, model.VoiceRecordingStatusProcessing, now).
			AddRow(1, 7, 3, model.VoiceRecordingStatusCompleted, now.Add(-time.Hour)))

	recs, err := repo.ListRecordingsByApiaryID(context.Background(), 3)
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
