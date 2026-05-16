-- +goose Up
ALTER TABLE todos_history RENAME TO todos_history_legacy_backup_005;

CREATE TABLE IF NOT EXISTS todos_history (
  id UUID PRIMARY KEY,
  todo_id UUID REFERENCES todos(id),
  date TEXT NOT NULL,
  completed BOOLEAN DEFAULT false,
  completed_at TIMESTAMP,
  notes TEXT,
  UNIQUE(todo_id, date)
);

WITH normalized AS (
  SELECT
    id,
    todo_id,
    date(date) AS normalized_date,
    completed,
    completed_at,
    notes,
    ROW_NUMBER() OVER (
      PARTITION BY todo_id, date(date)
      ORDER BY
        COALESCE(completed, 0) DESC,
        CASE WHEN COALESCE(notes, '') = 'backfilled' THEN 1 ELSE 0 END ASC,
        COALESCE(completed_at, '') DESC,
        id ASC
    ) AS row_rank
  FROM todos_history_legacy_backup_005
  WHERE date IS NOT NULL
)
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
SELECT
  id,
  todo_id,
  normalized_date,
  completed,
  completed_at,
  notes
FROM normalized
WHERE row_rank = 1
  AND normalized_date IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS todos_history;
ALTER TABLE todos_history_legacy_backup_005 RENAME TO todos_history;
