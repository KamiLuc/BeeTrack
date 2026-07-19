-- +goose Up
CREATE TABLE blockchain_jobs (
    id               BIGSERIAL   PRIMARY KEY,
    batch_id         BIGINT      NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE,
    job_type         TEXT        NOT NULL,
    status           TEXT        NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'submitting', 'submitted', 'pending_confirmation', 'confirmed', 'failed', 'reverted')),
    attempt_count    INT         NOT NULL DEFAULT 0,
    next_retry_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error       TEXT,
    certification_id BIGINT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON blockchain_jobs(status, next_retry_at);
CREATE INDEX ON blockchain_jobs(batch_id);

-- +goose Down
DROP TABLE blockchain_jobs;
