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

func newAssistantTestRepo(t *testing.T) (*AssistantRepository, sqlmock.Sqlmock) {
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
	return NewAssistantRepository(gdb), mock
}

func TestAssistantRepository_CreateConversation(t *testing.T) {
	repo, mock := newAssistantTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "assistant_conversations"`)).
		WithArgs(int64(7), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	conv, err := repo.CreateConversation(context.Background(), 7)
	if err != nil {
		t.Fatalf("CreateConversation returned error: %v", err)
	}
	if conv.ID != 1 || conv.UserID != 7 {
		t.Fatalf("unexpected conversation: %+v", conv)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAssistantRepository_GetConversationByID_Found(t *testing.T) {
	repo, mock := newAssistantTestRepo(t)

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "assistant_conversations"`)).
		WithArgs(int64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "created_at"}).
			AddRow(1, 7, now))

	conv, err := repo.GetConversationByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetConversationByID returned error: %v", err)
	}
	if conv == nil || conv.ID != 1 || conv.UserID != 7 {
		t.Fatalf("unexpected conversation: %+v", conv)
	}
}

func TestAssistantRepository_GetConversationByID_NotFound(t *testing.T) {
	repo, mock := newAssistantTestRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "assistant_conversations"`)).
		WithArgs(int64(99), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "created_at"}))

	conv, err := repo.GetConversationByID(context.Background(), 99)
	if err != nil {
		t.Fatalf("GetConversationByID returned error: %v", err)
	}
	if conv != nil {
		t.Fatalf("expected nil conversation, got %+v", conv)
	}
}

func TestAssistantRepository_CreateAndListMessageLogs(t *testing.T) {
	repo, mock := newAssistantTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "assistant_message_logs"`)).
		WithArgs(int64(1), model.AssistantMessageRoleUser, "hello", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	log1 := &model.AssistantMessageLog{ConversationID: 1, Role: model.AssistantMessageRoleUser, Content: "hello"}
	if err := repo.CreateMessageLog(context.Background(), log1); err != nil {
		t.Fatalf("CreateMessageLog returned error: %v", err)
	}
	if log1.ID != 1 {
		t.Fatalf("expected id 1, got %d", log1.ID)
	}

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "assistant_message_logs"`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "conversation_id", "role", "content", "created_at"}).
			AddRow(1, 1, model.AssistantMessageRoleUser, "hello", now).
			AddRow(2, 1, model.AssistantMessageRoleAssistant, "hi there", now.Add(time.Second)))

	logs, err := repo.ListMessageLogs(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListMessageLogs returned error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
	if logs[0].Role != model.AssistantMessageRoleUser || logs[1].Role != model.AssistantMessageRoleAssistant {
		t.Fatalf("logs not returned in expected order: %+v", logs)
	}
}

func TestAssistantRepository_CreateToolCall(t *testing.T) {
	repo, mock := newAssistantTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "assistant_tool_calls"`)).
		WithArgs(int64(5), "list_hives", "{}", `{"ok":true}`, false, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	call := &model.AssistantToolCall{
		MessageID: 5,
		ToolName:  "list_hives",
		Input:     []byte(`{}`),
		Result:    []byte(`{"ok":true}`),
		IsError:   false,
	}
	if err := repo.CreateToolCall(context.Background(), call); err != nil {
		t.Fatalf("CreateToolCall returned error: %v", err)
	}
	if call.ID != 1 {
		t.Fatalf("expected id 1, got %d", call.ID)
	}
}
