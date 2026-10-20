-- name: AddTodo :one 
INSERT INTO todos (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since, parent_id)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
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

-- name: GetTodoByName :many
-- Candidates for a fuzzy name search, best first. Subtasks come with their
-- parent's name; on equal relevance a top-level task wins.
WITH ranked_todos AS (
  SELECT t.*,
    COALESCE(p.name, '') AS parent_name,
    CASE
    -- same text (case-insensitive; the caller passes a lowercase term)
    WHEN LOWER(t.name) = sqlc.arg(term) THEN 1
    -- begins with the term
    WHEN t.name LIKE sqlc.arg(term) || '%' THEN 2
    ELSE 3
    END AS relevance
  FROM todos t
  LEFT JOIN todos p ON p.id = t.parent_id
  WHERE t.name LIKE '%' || sqlc.arg(term) || '%'
)
SELECT * FROM ranked_todos
ORDER BY relevance, parent_name != '', name COLLATE NOCASE
LIMIT 20;

-- name: GetSubtaskByPath :one
-- The subtask written as "Parent/Subtask" (exact names, case-insensitive).
SELECT c.*, p.name AS parent_name
FROM todos c
JOIN todos p ON p.id = c.parent_id
WHERE LOWER(p.name) = sqlc.arg(parent) AND p.parent_id IS NULL
  AND LOWER(c.name) = sqlc.arg(child);

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

-- name: DeleteTodoByID :exec
DELETE FROM todos
WHERE id = ?;

-- name: DeleteSubtasks :exec
DELETE FROM todos
WHERE parent_id = ?;

-- name: ReopenTodo :exec
-- Undoes a parent's completion when it gets a pending subtask: a one-off
-- task is no longer concluded, a daily one loses today's history row.
UPDATE todos
SET concluded = 0, updated_at = datetime('now', 'localtime')
WHERE id = ?;

-- name: SetSubtasksDaily :exec
-- Subtasks follow their parent's daily mode.
UPDATE todos
SET is_daily = sqlc.arg(is_daily),
    history_since = COALESCE(history_since, date('now', 'localtime'))
WHERE parent_id = sqlc.arg(parent_id);

-- name: ClampSubtaskDates :exec
-- Keeps subtasks due no later than their parent.
UPDATE todos
SET expires_at = sqlc.arg(expires_at)
WHERE parent_id = sqlc.arg(parent_id) AND expires_at > sqlc.arg(expires_at);

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
