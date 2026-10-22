-- +goose Up
-- Optional time of day (HH:MM, local) for a task. expires_at can't carry it:
-- it holds local midnight for a dated task and the creation time for an
-- undated one, so its time part means nothing.
ALTER TABLE todos ADD COLUMN due_time TEXT;

-- +goose Down
ALTER TABLE todos DROP COLUMN due_time;
