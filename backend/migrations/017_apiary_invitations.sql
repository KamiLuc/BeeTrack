-- +goose Up
CREATE TABLE apiary_invitations (
    id                  BIGSERIAL    PRIMARY KEY,
    apiary_id           BIGINT       NOT NULL REFERENCES apiaries(id) ON DELETE CASCADE,
    invited_by_user_id  BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_email       TEXT         NOT NULL,
    status              TEXT         NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_apiary_invitations_pending
    ON apiary_invitations (apiary_id, invited_email)
    WHERE status = 'pending';

-- +goose Down
DROP TABLE apiary_invitations;
