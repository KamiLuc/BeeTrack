-- +goose Up
ALTER TABLE inspections RENAME COLUMN frames_honey TO frames_feed;

-- +goose Down
ALTER TABLE inspections RENAME COLUMN frames_feed TO frames_honey;
