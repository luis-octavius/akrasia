-- name: AddDump :one 
INSERT INTO dump (id, name) VALUES (
  ?,
  ?
) 
RETURNING *;

-- name: GetAllDump :many 
SELECT * FROM dump 
ORDER BY created_at;

-- name: GetDumpById :one 
SELECT * FROM dump 
WHERE id = ?; 


