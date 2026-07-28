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

func (r *AssistantRepository) ListMessageLogs(ctx context.Context, conversationID int64) ([]*model.AssistantMessageLog, error) {
	var logs []*model.AssistantMessageLog
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *AssistantRepository) CreateMessageLog(ctx context.Context, log *model.AssistantMessageLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AssistantRepository) CreateToolCall(ctx context.Context, call *model.AssistantToolCall) error {
	return r.db.WithContext(ctx).Create(call).Error
}
