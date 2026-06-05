-- +goose Up
ALTER TABLE inspections ADD COLUMN frames_brood INTEGER;

-- +goose Down
ALTER TABLE inspections DROP COLUMN frames_brood;
