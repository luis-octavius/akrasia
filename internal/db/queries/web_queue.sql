-- name: QueueSeen :one
SELECT EXISTS(SELECT 1 FROM web_queue WHERE id = ?) AS seen;

-- name: QueueRemember :exec
INSERT INTO web_queue (id) VALUES (?)
ON CONFLICT(id) DO NOTHING;

-- name: QueuePrune :exec
-- Replays come within days at most; older ids only take space.
DELETE FROM web_queue WHERE created_at < datetime('now', 'localtime', '-30 days');
