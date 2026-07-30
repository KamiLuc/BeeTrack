-- +goose Up
CREATE TABLE voice_recordings (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         BIGINT      NOT NULL REFERENCES users(id),
    apiary_id       BIGINT      NOT NULL REFERENCES apiaries(id),
    status          TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'completed', 'accepted', 'rejected', 'failed', 'cancelled')),
    audio_path        TEXT,
    transcript        TEXT,
    detected_language TEXT,
    error_message     TEXT,
    retry_count       SMALLINT    NOT NULL DEFAULT 0,
    next_attempt_at   TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at      TIMESTAMPTZ
);

CREATE INDEX ON voice_recordings(apiary_id, created_at DESC);

CREATE TABLE voice_actions (
    id                  BIGSERIAL   PRIMARY KEY,
    voice_recording_id  BIGINT      NOT NULL REFERENCES voice_recordings(id) ON DELETE CASCADE,
    sequence            SMALLINT    NOT NULL,
    spoken_hive_name     TEXT,
    hive_id             BIGINT      REFERENCES hives(id),
    tool_name           TEXT,
    tool_arguments      JSONB,
    status              TEXT        NOT NULL DEFAULT 'proposed'
        CHECK (status IN ('proposed', 'applied', 'error')),
    previous_hive_state JSONB,
    result_type         TEXT
        CHECK (result_type IN ('inspection', 'treatment', 'harvest', 'feeding', 'hive_status', 'error')),
    result_record_id    BIGINT,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON voice_actions(voice_recording_id, sequence);

-- +goose Down
DROP TABLE voice_actions;
DROP TABLE voice_recordings;
