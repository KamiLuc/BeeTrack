-- +goose Up
ALTER TABLE hives DROP COLUMN queenless;

-- +goose Down
ALTER TABLE hives ADD COLUMN queenless BOOLEAN NOT NULL DEFAULT FALSE;
