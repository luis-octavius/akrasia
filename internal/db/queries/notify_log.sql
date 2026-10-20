-- name: GetNotifiedSlots :many
SELECT slot FROM notify_log
WHERE day = ?;

-- name: AddNotifiedSlot :exec
INSERT OR IGNORE INTO notify_log (day, slot) VALUES (?, ?);
