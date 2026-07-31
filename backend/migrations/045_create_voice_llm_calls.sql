-- +goose Up
CREATE TABLE voice_llm_calls (
    id                 BIGSERIAL   PRIMARY KEY,
    voice_recording_id BIGINT      NOT NULL REFERENCES voice_recordings(id) ON DELETE CASCADE,
    phase              TEXT        NOT NULL CHECK (phase IN ('resolve_hive', 'propose_actions')),
    request            JSONB       NOT NULL,
    response           JSONB,
    is_error           BOOLEAN     NOT NULL DEFAULT false,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON voice_llm_calls(voice_recording_id);

-- +goose Down
DROP TABLE voice_llm_calls;
