-- +goose Up
CREATE TABLE honey_batch_certification_requests (
    id                BIGSERIAL   PRIMARY KEY,
    batch_id          BIGINT      NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE,
    requested_by      BIGINT      NOT NULL REFERENCES users(id),
    status            TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason  TEXT,
    reviewed_by       BIGINT      REFERENCES users(id),
    reviewed_at       TIMESTAMPTZ,
    blockchain_job_id BIGINT      REFERENCES blockchain_jobs(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON honey_batch_certification_requests(status, created_at DESC);
CREATE INDEX ON honey_batch_certification_requests(batch_id, created_at DESC);
CREATE UNIQUE INDEX honey_batch_certification_requests_pending_batch_id
    ON honey_batch_certification_requests(batch_id)
    WHERE status = 'pending';

-- +goose Down
DROP TABLE honey_batch_certification_requests;
