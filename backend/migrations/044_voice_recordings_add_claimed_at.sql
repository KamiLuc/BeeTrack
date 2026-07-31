-- +goose Up
ALTER TABLE voice_recordings ADD COLUMN claimed_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE voice_recordings DROP COLUMN claimed_at;
