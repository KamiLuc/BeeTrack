-- +goose Up
ALTER TABLE hives DROP COLUMN frames;

-- +goose Down
ALTER TABLE hives ADD COLUMN frames INT NOT NULL DEFAULT 0;
