-- +goose Up
ALTER TABLE inspections ADD COLUMN colony_strength VARCHAR(20);

-- +goose Down
ALTER TABLE inspections DROP COLUMN colony_strength;
