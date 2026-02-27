-- name: AddTodo :one 
INSERT INTO todos (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ? 
)
RETURNING *;

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
    -- same text 
    WHEN LOWER(title) = LOWER(?) THEN 1
    -- begins with the term 
    WHEN LOWER(title) LIKE LOWER(?) || '%' THEN 2
    WHEN LOWER(title) LIKE '%' || LOWER(?) || '%' THEN 3
    ELSE 4 
    END as relevance
  FROM todos
  WHERE LOWER(title) LIKE '%' || LOWER(?) || '%'
) 
SELECT * FROM ranked_todos
ORDER BY relevance, title COLLATE NOCASE
LIMIT 1;

-- name: GetIDByName :one 
SELECT id FROM todos 
WHERE name = ?;

-- name: UpdateTodoStatusByName :one 
UPDATE todos 
SET concluded = true, updated_at = datetime('now')
WHERE name LIKE ?
RETURNING *;

-- name: UpdateDailyTodo :many
UPDATE todos 
SET 
  expires_at = ?, 
  updated_at = CURRENT_TIMESTAMP, 
  concluded = false 
WHERE 
  is_daily = true 
  AND expires_at < datetime('now', 'start of day', 'localtime')
RETURNING *;

-- name: DeleteTodoByName :exec 
DELETE FROM todos 
WHERE name LIKE ?;

-- name: DeleteConcluded :exec 
DELETE FROM todos 
WHERE concluded = true; 

-- name: CheckExpired :many 
SELECT * FROM todos 
WHERE expires_at < datetime('now') AND is_daily = false
ORDER BY expires_at DESC;

-- name: AutoCompleteDelete :one 
SELECT name FROM todos 
WHERE NAME LIKE ? || '%'
ORDER BY NAME LIMIT 5;
