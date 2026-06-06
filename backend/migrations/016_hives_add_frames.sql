-- +goose Up
ALTER TABLE hives ADD COLUMN frames INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE hives DROP COLUMN frames;
