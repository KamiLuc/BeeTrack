-- +goose Up
ALTER TABLE harvests
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE harvests DROP COLUMN IF EXISTS notes;
