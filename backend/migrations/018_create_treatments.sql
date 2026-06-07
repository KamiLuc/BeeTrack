-- +goose Up
CREATE TABLE treatments (
    id            BIGSERIAL PRIMARY KEY,
    hive_id       BIGINT      NOT NULL REFERENCES hives(id) ON DELETE CASCADE,
    treated_by    BIGINT      NOT NULL REFERENCES users(id),
    treated_at    TIMESTAMPTZ NOT NULL,
    medicine_name TEXT        NOT NULL,
    dose          TEXT        NOT NULL DEFAULT '1',
    notes         TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON treatments(hive_id);
CREATE INDEX ON treatments(treated_at DESC);

-- +goose Down
DROP TABLE treatments;
