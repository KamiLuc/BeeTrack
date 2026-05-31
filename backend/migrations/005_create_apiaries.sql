-- +goose Up
CREATE TABLE apiaries (
    id            BIGSERIAL        PRIMARY KEY,
    owner_user_id BIGINT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT             NOT NULL,
    lat           DOUBLE PRECISION,
    lng           DOUBLE PRECISION,
    grid_rows     INT              NOT NULL CHECK (grid_rows >= 1),
    grid_cols     INT              NOT NULL CHECK (grid_cols >= 1),
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE apiary_members (
    apiary_id BIGINT      NOT NULL REFERENCES apiaries(id) ON DELETE CASCADE,
    user_id   BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      TEXT        NOT NULL CHECK (role IN ('owner', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (apiary_id, user_id)
);

-- +goose Down
DROP TABLE apiary_members;
DROP TABLE apiaries;
