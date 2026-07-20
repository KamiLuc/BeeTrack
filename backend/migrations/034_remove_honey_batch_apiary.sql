-- +goose Up
-- Honey batches are no longer scoped to an apiary — a batch belongs to its
-- creator only, matching how certification/verification already work
-- (owner-scoped or token-scoped, never apiary-scoped).
ALTER TABLE honey_batches DROP COLUMN apiary_id;

-- +goose Down
-- Lossy: original apiary associations aren't recoverable, so existing rows
-- come back with no apiary set rather than their prior value.
ALTER TABLE honey_batches ADD COLUMN apiary_id BIGINT REFERENCES apiaries(id) ON DELETE CASCADE;
