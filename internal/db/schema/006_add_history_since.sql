-- +goose Up
ALTER TABLE todos ADD COLUMN history_since DATE;

-- Existing daily tasks: trust history from whichever came first —
-- the earliest row we already have, or the task's creation date.
-- This is what replaces the old 'backfilled' sentinel: instead of a
-- neutral marker sprinkled through the data, we simply never look at
-- dates before this boundary when computing a streak.
UPDATE todos
SET history_since = COALESCE(
  (SELECT MIN(date) FROM todos_history WHERE todos_history.todo_id = todos.id),
  date(created_at)
)
WHERE is_daily = true;

-- +goose Down
ALTER TABLE todos DROP COLUMN history_since;
