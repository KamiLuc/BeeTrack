package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	AssistantMessageRoleUser      = "user"
	AssistantMessageRoleAssistant = "assistant"
)

type AssistantConversation struct {
	ID        int64
	UserID    int64
	CreatedAt time.Time
}

type AssistantMessageLog struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
	CreatedAt      time.Time
}

type AssistantToolCall struct {
	ID        int64
	MessageID int64
	ToolName  string
	Input     datatypes.JSON
	Result    datatypes.JSON
	IsError   bool
	CreatedAt time.Time
}
