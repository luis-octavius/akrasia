-- name: AddTodo :one 
INSERT INTO todos (id, name, description, created_at, updated_at, concluded)
VALUES (
  ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetTodos :many 
SELECT * FROM todos; 




