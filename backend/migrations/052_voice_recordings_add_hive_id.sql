-- +goose Up
ALTER TABLE voice_recordings ADD COLUMN hive_id BIGINT REFERENCES hives(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE voice_recordings DROP COLUMN hive_id;
