-- +goose Up 
ALTER TABLE todos 
ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium';

ALTER TABLE todos 
ADD COLUMN is_daily BOOL NOT NULL DEFAULT false;

-- +goose Down
DROP COLUMN IF EXISTS priority FROM todos;
DROP COLUMN IF EXISTS is_daily FROM todos;
