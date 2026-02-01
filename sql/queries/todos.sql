-- name: AddTodo :one 
INSERT INTO todos (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ? 
)
RETURNING *;

-- name: GetTodos :many 
SELECT * FROM todos; 

-- name: GetTodoByName :one 
SELECT * FROM todos 
WHERE name LIKE ?; 

-- name: UpdateTodoStatusByName :one 
UPDATE todos 
SET concluded = true, updated_at = datetime('now')
WHERE name LIKE ?
RETURNING *;

-- name: DeleteTodoByName :exec 
DELETE FROM todos 
WHERE name LIKE ?;

-- name: DeleteConcluded :exec 
DELETE FROM todos 
WHERE concluded = true; 

-- name: CheckExpired :many 
SELECT * FROM todos 
WHERE expires_at < datetime('now')
ORDER BY expires_at DESC;




