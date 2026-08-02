-- +goose Up
ALTER TABLE hives ADD COLUMN box_needs_adding BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE hives DROP COLUMN box_needs_adding;
