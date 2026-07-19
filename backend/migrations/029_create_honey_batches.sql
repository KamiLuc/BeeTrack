-- +goose Up
CREATE TABLE honey_batches (
    id                 BIGSERIAL   PRIMARY KEY,
    user_id            BIGINT      NOT NULL REFERENCES users(id),
    apiary_id          BIGINT      NOT NULL REFERENCES apiaries(id) ON DELETE CASCADE,
    verification_token UUID        NOT NULL,
    gathering_date     DATE        NOT NULL,
    amount_grams       BIGINT      NOT NULL CHECK (amount_grams > 0),
    processing_method  VARCHAR(20) NOT NULL CHECK (processing_method IN ('raw', 'filtered', 'pasteurized')),
    honey_type         TEXT        NOT NULL,
    lab_pdf_url        TEXT        NOT NULL,
    pdf_file_hash      CHAR(64)    NOT NULL,
    metadata_hash      CHAR(64)    NOT NULL,
    deleted_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (verification_token)
);

CREATE INDEX ON honey_batches(user_id, created_at DESC);
CREATE INDEX ON honey_batches(apiary_id, created_at DESC);

-- +goose Down
DROP TABLE honey_batches;
