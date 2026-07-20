-- +goose Up
ALTER TABLE honey_batches ADD COLUMN pdf_filename TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE honey_batches DROP COLUMN pdf_filename;
