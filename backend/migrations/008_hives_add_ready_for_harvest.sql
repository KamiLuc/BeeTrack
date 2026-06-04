-- +goose Up
ALTER TABLE hives ADD COLUMN ready_for_harvest BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE hives DROP COLUMN ready_for_harvest;
