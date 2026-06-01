-- +goose Up
CREATE TABLE hives (
    id         BIGSERIAL    PRIMARY KEY,
    apiary_id  BIGINT       NOT NULL REFERENCES apiaries(id) ON DELETE CASCADE,
    name       TEXT         NOT NULL,
    type       TEXT         NOT NULL DEFAULT 'langstroth',
    active     BOOLEAN      NOT NULL DEFAULT TRUE,
    grid_row   INT          NOT NULL CHECK (grid_row >= 0),
    grid_col   INT          NOT NULL CHECK (grid_col >= 0),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (apiary_id, grid_row, grid_col)
);

-- +goose Down
DROP TABLE hives;
