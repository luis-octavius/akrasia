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

-- name: UpdateTodo :one
-- Rewrites the editable fields of a task. history_since only gets a value
-- when it has none, so a one-off task turned daily starts its streak
-- window today and an existing window is never moved.
UPDATE todos
SET name = sqlc.arg(name),
    description = sqlc.arg(description),
    priority = sqlc.arg(priority),
    is_daily = sqlc.arg(is_daily),
    expires_at = sqlc.arg(expires_at),
    history_since = COALESCE(history_since, date('now', 'localtime')),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
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

-- name: ListTodosWithTodayStatus :many 
SELECT t.*, CAST(COALESCE(h.completed, 0) AS BOOLEAN) AS done_today
FROM todos t 
LEFT JOIN todos_history h 
ON h.todo_id = t.id AND h.date = date('now', 'localtime')
ORDER BY t.expires_at;
