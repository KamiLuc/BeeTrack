-- +goose Up
CREATE TABLE feedings (
    id         BIGSERIAL PRIMARY KEY,
    hive_id    BIGINT      NOT NULL REFERENCES hives(id) ON DELETE CASCADE,
    fed_by     BIGINT      NOT NULL REFERENCES users(id),
    fed_at     TIMESTAMPTZ NOT NULL,
    feed_type  TEXT        NOT NULL,
    amount     TEXT        NOT NULL DEFAULT '',
    notes      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON feedings(hive_id);
CREATE INDEX ON feedings(fed_at DESC);

-- +goose Down
DROP TABLE feedings;
