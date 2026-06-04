-- +goose Up
CREATE TABLE inspections (
    id                      BIGSERIAL PRIMARY KEY,
    hive_id                 BIGINT       NOT NULL REFERENCES hives(id) ON DELETE CASCADE,
    inspected_by            BIGINT       NOT NULL REFERENCES users(id),
    inspected_at            TIMESTAMPTZ  NOT NULL,
    queen_status            VARCHAR(20),
    brood_pattern           VARCHAR(20),
    frames_honey            INT,
    frames_pollen           INT,
    varroa_count            INT,
    queen_cells_count       INT,
    aggressiveness          VARCHAR(20),
    frames_added_foundation INT,
    frames_added_drawn      INT,
    queen_added             BOOLEAN      NOT NULL DEFAULT FALSE,
    notes                   TEXT,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX ON inspections(hive_id);
CREATE INDEX ON inspections(inspected_at DESC);

-- +goose Down
DROP TABLE inspections;
