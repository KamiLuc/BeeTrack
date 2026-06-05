-- +goose Up
ALTER TABLE inspections
DROP COLUMN varroa_count;

-- +goose Down
ALTER TABLE inspections
ADD COLUMN varroa_count INT;