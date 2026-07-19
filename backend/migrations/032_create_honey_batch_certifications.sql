-- +goose Up
CREATE TABLE honey_batch_certifications (
    id                     BIGSERIAL   PRIMARY KEY,
    batch_id               BIGINT      NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE,
    chain_id               INT         NOT NULL,
    contract_address       CHAR(42)    NOT NULL,
    transaction_hash       CHAR(66),
    block_number           BIGINT,
    status                 TEXT        NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'submitting', 'submitted', 'pending_confirmation', 'confirmed', 'failed', 'reverted')),
    gas_used               BIGINT,
    confirmation_timestamp TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON honey_batch_certifications(batch_id, created_at DESC);
CREATE UNIQUE INDEX honey_batch_certifications_live_batch_id
    ON honey_batch_certifications(batch_id)
    WHERE status IN ('submitted', 'pending_confirmation', 'confirmed');

ALTER TABLE blockchain_jobs
    ADD CONSTRAINT blockchain_jobs_certification_id_fkey
    FOREIGN KEY (certification_id) REFERENCES honey_batch_certifications(id);

-- +goose Down
ALTER TABLE blockchain_jobs DROP CONSTRAINT blockchain_jobs_certification_id_fkey;
DROP TABLE honey_batch_certifications;
