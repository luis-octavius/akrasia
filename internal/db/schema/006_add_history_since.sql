-- +goose Up
ALTER TABLE todos ADD COLUMN history_since DATE;

-- +goose Down
ALTER TABLE todos DROP COLUMN history_since;
