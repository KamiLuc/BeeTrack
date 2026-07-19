-- +goose Up
CREATE TABLE honey_batch_qr_codes (
    id           BIGSERIAL    PRIMARY KEY,
    batch_id     BIGINT       NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE,
    qr_code_data VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX ON honey_batch_qr_codes(batch_id);

-- +goose Down
DROP TABLE honey_batch_qr_codes;
