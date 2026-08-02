-- +goose Up
ALTER TABLE inspections ADD COLUMN box_added BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE inspections DROP COLUMN box_added;
