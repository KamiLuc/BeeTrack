-- +goose Up
ALTER TABLE hives ADD COLUMN queen_needs_replacement_since TIMESTAMPTZ;
ALTER TABLE hives ADD COLUMN ready_for_harvest_since TIMESTAMPTZ;
ALTER TABLE hives ADD COLUMN needs_food_since TIMESTAMPTZ;
ALTER TABLE hives ADD COLUMN box_needs_adding_since TIMESTAMPTZ;

-- +goose Down
ALTER TABLE hives DROP COLUMN queen_needs_replacement_since;
ALTER TABLE hives DROP COLUMN ready_for_harvest_since;
ALTER TABLE hives DROP COLUMN needs_food_since;
ALTER TABLE hives DROP COLUMN box_needs_adding_since;
