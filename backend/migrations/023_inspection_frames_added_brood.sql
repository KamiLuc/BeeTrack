-- +goose Up
ALTER TABLE inspections ADD COLUMN frames_added_brood INT;
ALTER TABLE inspections RENAME COLUMN frames_added_honey TO frames_added_feed;

-- +goose Down
ALTER TABLE inspections RENAME COLUMN frames_added_feed TO frames_added_honey;
ALTER TABLE inspections DROP COLUMN frames_added_brood;
