-- +goose Up
ALTER TABLE listings ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE listings ADD COLUMN rejection_reason TEXT;
ALTER TABLE listings ADD COLUMN first_approved_at TIMESTAMPTZ;
ALTER TABLE listings ADD COLUMN reviewed_by BIGINT REFERENCES users(id);
ALTER TABLE listings ADD COLUMN reviewed_at TIMESTAMPTZ;

-- Existing listings predate the moderation workflow — grandfather them in as
-- approved so this migration doesn't retroactively de-list anything.
UPDATE listings SET status = 'approved', first_approved_at = created_at;

CREATE INDEX ON listings(status, created_at DESC);

-- +goose Down
ALTER TABLE listings DROP COLUMN status;
ALTER TABLE listings DROP COLUMN rejection_reason;
ALTER TABLE listings DROP COLUMN first_approved_at;
ALTER TABLE listings DROP COLUMN reviewed_by;
ALTER TABLE listings DROP COLUMN reviewed_at;
