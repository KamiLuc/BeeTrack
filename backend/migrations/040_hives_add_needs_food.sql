-- +goose Up
ALTER TABLE hives ADD COLUMN needs_food BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE hives DROP COLUMN needs_food;
