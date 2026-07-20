-- +goose Up
-- CHAR(64) right-pads short values with spaces on storage/read, which would
-- turn an empty "no PDF yet" hash into 64 spaces instead of "". VARCHAR(64)
-- doesn't pad, so an empty string round-trips as empty.
ALTER TABLE honey_batches ALTER COLUMN pdf_file_hash TYPE VARCHAR(64);

-- +goose Down
ALTER TABLE honey_batches ALTER COLUMN pdf_file_hash TYPE CHAR(64);
