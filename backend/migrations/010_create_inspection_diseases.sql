-- +goose Up
CREATE TABLE inspection_diseases (
    id            BIGSERIAL PRIMARY KEY,
    inspection_id BIGINT       NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
    disease       VARCHAR(50)  NOT NULL,
    notes         TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX ON inspection_diseases(inspection_id);

-- +goose Down
DROP TABLE inspection_diseases;
