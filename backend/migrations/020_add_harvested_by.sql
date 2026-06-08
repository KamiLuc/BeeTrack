-- +goose Up
ALTER TABLE harvests
    ADD COLUMN harvested_by BIGINT REFERENCES users(id);

CREATE INDEX ON harvests(harvested_by);

-- +goose Down
ALTER TABLE harvests DROP COLUMN harvested_by;
