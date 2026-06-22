-- name: AddDailyTaskHistory :one
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date('now', 'localtime'), ?, ?, ?
) 
ON CONFLICT(todo_id, date) DO NOTHING
RETURNING *;

-- name: AddDailyTaskHistoryForDate :execresult
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date(?), ?, ?, ?
) 
ON CONFLICT(todo_id, date) DO NOTHING;

-- name: AddTodoHistory :one
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date('now', 'localtime'), ?, ?, ?
) 
ON CONFLICT(todo_id, date) DO UPDATE SET
    completed = excluded.completed,
    completed_at = excluded.completed_at,
    notes = excluded.notes
RETURNING *;

-- name: GetCurrentStreak :one
-- Returns the number of consecutive completed days ending on the most recent entry.
-- Backfilled rows (notes = 'backfilled') are neutral — they don't break the streak
-- but also don't count as completed days.
WITH ordered AS (
    SELECT
        date,
        completed,
        notes,
        julianday(date) - julianday(LAG(date) OVER (ORDER BY date ASC)) AS days_diff
    FROM todos_history
    WHERE todo_id = ?
      AND date <= date('now', 'localtime')
    ORDER BY date ASC
),
grouped AS (
    SELECT
        date,
        completed,
        notes,
        SUM(
            CASE
                WHEN notes = 'backfilled' THEN 0
                WHEN completed = 0 THEN 1
                WHEN days_diff > 1 THEN 1
                ELSE 0
            END
        ) OVER (ORDER BY date ASC ROWS UNBOUNDED PRECEDING) AS streak_group
    FROM ordered
),
current_group AS (
    SELECT streak_group
    FROM grouped
    ORDER BY date DESC
    LIMIT 1
)
SELECT COUNT(*) AS current_streak
FROM grouped
WHERE (completed = 1 OR notes = 'backfilled')
  AND streak_group = (SELECT streak_group FROM current_group);

-- name: GetStreakHistory :many
-- Returns all completed streak intervals ordered by length descending.
-- Backfilled rows (notes = 'backfilled') are neutral — they don't break the streak
-- but also don't count as completed days.
WITH ordered AS (
    SELECT
        date,
        completed,
        notes,
        julianday(date) - julianday(LAG(date) OVER (ORDER BY date ASC)) AS days_diff
    FROM todos_history
    WHERE todo_id = ?
    ORDER BY date ASC
),
grouped AS (
    SELECT
        date,
        completed,
        notes,
        SUM(
            CASE
                WHEN notes = 'backfilled' THEN 0
                WHEN completed = 0 THEN 1
                WHEN days_diff > 1 THEN 1
                ELSE 0
            END
        ) OVER (ORDER BY date ASC ROWS UNBOUNDED PRECEDING) AS streak_id
    FROM ordered
),
streaks AS (
    SELECT
        streak_id,
        MIN(date) AS start_date,
        MAX(date) AS end_date,
        COUNT(*)  AS streak_length
    FROM grouped
    WHERE (completed = 1 OR notes = 'backfilled')
    GROUP BY streak_id
)
SELECT
    start_date,
    end_date,
    streak_length
FROM streaks
ORDER BY streak_length DESC, start_date DESC;
