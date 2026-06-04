-- +goose Up
ALTER TABLE hives ADD COLUMN queenless BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE hives DROP COLUMN queenless;
