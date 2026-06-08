-- +goose Up
ALTER TABLE harvests
    ADD COLUMN notes TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE harvests DROP COLUMN notes;
