-- name: AddTodoHistory :one 
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
  ?, ?, ?, ?, ?, ?
) RETURNING *;
