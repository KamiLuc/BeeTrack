-- +goose Up
CREATE TABLE inspection_images (
    id BIGSERIAL PRIMARY KEY,
    inspection_id BIGINT NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX inspection_images_inspection_id_idx ON inspection_images(inspection_id);

-- +goose Down
DROP TABLE inspection_images;
