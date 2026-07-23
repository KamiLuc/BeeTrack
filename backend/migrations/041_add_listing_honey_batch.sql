-- +goose Up
ALTER TABLE listings ADD COLUMN honey_batch_id BIGINT REFERENCES honey_batches(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX idx_listings_honey_batch_id_unique ON listings(honey_batch_id) WHERE honey_batch_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_listings_honey_batch_id_unique;
ALTER TABLE listings DROP COLUMN honey_batch_id;
