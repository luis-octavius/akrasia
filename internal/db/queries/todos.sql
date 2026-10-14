-- name: AddTodo :one 
INSERT INTO todos (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: ExtendHistorySince :exec
-- Moves a task's streak lower bound earlier when backfill records days
-- that predate it. Never moves it later (guarded by the WHERE), so it is
-- safe to call with any candidate date.
UPDATE todos
SET history_since = sqlc.arg(earliest)
WHERE id = sqlc.arg(id)
  AND (history_since IS NULL OR history_since > sqlc.arg(earliest));

-- name: GetTodos :many 
SELECT * FROM todos
ORDER BY expires_at; 

-- name: GetDailyTodos :many 
SELECT * FROM todos 
WHERE is_daily = true; 

-- name: GetTodoByName :one 
WITH ranked_todos AS (
  SELECT *,
    CASE
    -- same text (case-insensitive; the caller passes a lowercase term)
    WHEN LOWER(name) = ? THEN 1
    -- begins with the term 
    WHEN name LIKE ? || '%' THEN 2
    WHEN name LIKE '%' || ? || '%' THEN 3
    ELSE 4 
    END as relevance
  FROM todos
  WHERE name LIKE '%' || ? || '%'
) 
SELECT * FROM ranked_todos
ORDER BY relevance, name COLLATE NOCASE
LIMIT 1;

-- name: GetIDByName :one 
SELECT id FROM todos 
WHERE name = ?;

-- name: UpdateTodoStatusById :one 
UPDATE todos 
SET concluded = 1, updated_at = datetime('now', 'localtime')
WHERE id = ?
RETURNING *;

-- name: DeleteTodoByName :exec 
DELETE FROM todos 
WHERE name = ?;

-- name: CheckExpired :many 
SELECT * FROM todos 
WHERE expires_at < datetime('now') AND is_daily = false
ORDER BY expires_at DESC;

-- name: AutoCompleteDelete :one 
SELECT name FROM todos 
WHERE NAME LIKE ? || '%'
ORDER BY NAME LIMIT 5;
