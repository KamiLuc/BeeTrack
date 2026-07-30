package repository

import (
	"context"
	"errors"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type AssistantRepository struct {
	db *gorm.DB
}

func NewAssistantRepository(db *gorm.DB) *AssistantRepository {
	return &AssistantRepository{db: db}
}

func (r *AssistantRepository) CreateConversation(ctx context.Context, userID int64) (*model.AssistantConversation, error) {
	conv := &model.AssistantConversation{UserID: userID}
	if err := r.db.WithContext(ctx).Create(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

// Returns nil, nil if not found — callers must still check UserID themselves before treating this as an ownership check.
func (r *AssistantRepository) GetConversationByID(ctx context.Context, id int64) (*model.AssistantConversation, error) {
	var conv model.AssistantConversation
	err := r.db.WithContext(ctx).First(&conv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListConversations returns userID's conversations, newest first, each paired with its first user message
// (as a preview) and total message count, for the chat history sidebar.
func (r *AssistantRepository) ListConversations(ctx context.Context, userID int64) ([]*model.AssistantConversationSummary, error) {
	var summaries []*model.AssistantConversationSummary
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.id AS id, c.created_at AS created_at, COALESCE(m.content, '') AS preview,
			(SELECT COUNT(*) FROM assistant_message_logs WHERE conversation_id = c.id) AS message_count
		FROM assistant_conversations c
		LEFT JOIN LATERAL (
			SELECT content FROM assistant_message_logs
			WHERE conversation_id = c.id AND role = ?
			ORDER BY created_at ASC LIMIT 1
		) m ON true
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC
	`, model.AssistantMessageRoleUser, userID).Scan(&summaries).Error
	return summaries, err
}

// DeleteConversation removes conversationID's row; its message logs and tool calls cascade via FK.
func (r *AssistantRepository) DeleteConversation(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.AssistantConversation{}, id).Error
}

func (r *AssistantRepository) ListMessageLogs(ctx context.Context, conversationID int64) ([]*model.AssistantMessageLog, error) {
	var logs []*model.AssistantMessageLog
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

// CountMessageLogs backs the per-conversation message cap.
func (r *AssistantRepository) CountMessageLogs(ctx context.Context, conversationID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AssistantMessageLog{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error
	return count, err
}

func (r *AssistantRepository) CreateMessageLog(ctx context.Context, log *model.AssistantMessageLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AssistantRepository) CreateToolCall(ctx context.Context, call *model.AssistantToolCall) error {
	return r.db.WithContext(ctx).Create(call).Error
}
