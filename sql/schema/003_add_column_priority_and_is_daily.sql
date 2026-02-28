-- +goose Up 
ALTER TABLE todos 
ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium';

ALTER TABLE todos 
ADD COLUMN is_daily BOOL NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE todos DROP COLUMN priority;
ALTER TABLE todos DROP COLUMN is_daily;
