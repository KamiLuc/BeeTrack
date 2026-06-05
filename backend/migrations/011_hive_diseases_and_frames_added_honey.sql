-- +goose Up
CREATE TABLE hive_diseases (
    id         BIGSERIAL   PRIMARY KEY,
    hive_id    BIGINT      NOT NULL REFERENCES hives(id) ON DELETE CASCADE,
    disease    VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON hive_diseases(hive_id);
ALTER TABLE inspections ADD COLUMN frames_added_honey INT;

-- +goose Down
ALTER TABLE inspections DROP COLUMN frames_added_honey;
DROP TABLE hive_diseases;
