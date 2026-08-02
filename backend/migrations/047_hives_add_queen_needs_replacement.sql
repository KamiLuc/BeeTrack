-- +goose Up
ALTER TABLE hives ADD COLUMN queen_needs_replacement BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE hives DROP COLUMN queen_needs_replacement;
