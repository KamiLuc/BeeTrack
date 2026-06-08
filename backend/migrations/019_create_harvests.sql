-- +goose Up
CREATE TABLE harvests (
    id           BIGSERIAL PRIMARY KEY,
    hive_id      BIGINT       NOT NULL REFERENCES hives(id) ON DELETE CASCADE,
    harvested_at TIMESTAMPTZ  NOT NULL,
    frames       INT          NOT NULL DEFAULT 0,
    half_frames  INT          NOT NULL DEFAULT 0,
    kilograms    NUMERIC(8,2) NOT NULL DEFAULT 0,
    notes        TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX ON harvests(hive_id);
CREATE INDEX ON harvests(harvested_at DESC);

-- +goose Down
DROP TABLE harvests;
