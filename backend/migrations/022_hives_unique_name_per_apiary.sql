-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS uq_hives_apiary_name ON hives (apiary_id, LOWER(name));

-- +goose Down
DROP INDEX IF EXISTS uq_hives_apiary_name;
