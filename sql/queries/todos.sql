-- name: AddTodo :one 
INSERT INTO todos (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ? 
)
RETURNING *;

-- name: GetTodos :many 
SELECT * FROM todos
ORDER BY expires_at; 

-- name: GetTodoByName :one 
SELECT * FROM todos 
WHERE name LIKE ?; 

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
SET expires_at = ?, updated_at = datetime('now'), concluded = false
WHERE is_daily = true AND date(expires_at) <= date('now') 
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
