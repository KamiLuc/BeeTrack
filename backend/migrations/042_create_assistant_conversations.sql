-- +goose Up
CREATE TABLE assistant_conversations (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON assistant_conversations(user_id, created_at DESC);

CREATE TABLE assistant_message_logs (
    id              BIGSERIAL   PRIMARY KEY,
    conversation_id BIGINT      NOT NULL REFERENCES assistant_conversations(id) ON DELETE CASCADE,
    role            TEXT        NOT NULL CHECK (role IN ('user', 'assistant')),
    content         TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON assistant_message_logs(conversation_id, created_at);

CREATE TABLE assistant_tool_calls (
    id         BIGSERIAL   PRIMARY KEY,
    message_id BIGINT      NOT NULL REFERENCES assistant_message_logs(id) ON DELETE CASCADE,
    tool_name  TEXT        NOT NULL,
    input      JSONB       NOT NULL,
    result     JSONB,
    is_error   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON assistant_tool_calls(message_id);

-- +goose Down
DROP TABLE assistant_tool_calls;
DROP TABLE assistant_message_logs;
DROP TABLE assistant_conversations;
