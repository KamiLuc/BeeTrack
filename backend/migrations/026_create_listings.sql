-- +goose Up
CREATE TABLE listings (
    id            BIGSERIAL      PRIMARY KEY,
    user_id       BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         TEXT           NOT NULL,
    description   TEXT           NOT NULL DEFAULT '',
    category      TEXT           NOT NULL CHECK (category IN (
                      'HONEY', 'POLLEN', 'BEE_COLONIES', 'QUEEN_BEES', 'BEEHIVES',
                      'POPULATED_BEEHIVES', 'EQUIPMENT', 'EXTRACTION_EQUIPMENT', 'FEED',
                      'SUPPLIES', 'WAX_FOUNDATION', 'BEESWAX', 'PROPOLIS', 'SERVICES', 'OTHER'
                  )),
    price         NUMERIC(10,2),
    quantity      TEXT           NOT NULL DEFAULT '',
    address       TEXT           NOT NULL DEFAULT '',
    apiary_id     BIGINT         REFERENCES apiaries(id) ON DELETE SET NULL,
    contact_phone TEXT           NOT NULL DEFAULT '',
    contact_email TEXT           NOT NULL DEFAULT '',
    is_hidden     BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX ON listings(user_id);
CREATE INDEX ON listings(category);
CREATE INDEX ON listings(created_at DESC);
CREATE INDEX ON listings(is_hidden);

CREATE TABLE listing_images (
    id            BIGSERIAL   PRIMARY KEY,
    listing_id    BIGINT      NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    image_url     TEXT        NOT NULL,
    display_order INT         NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON listing_images(listing_id);

CREATE TABLE listing_favorites (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id BIGINT      NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, listing_id)
);

CREATE INDEX ON listing_favorites(user_id);

-- +goose Down
DROP TABLE listing_favorites;
DROP TABLE listing_images;
DROP TABLE listings;
