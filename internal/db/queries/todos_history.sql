-- name: AddDailyTaskHistory :one
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date('now', 'localtime'), ?, ?, ?
) 
ON CONFLICT(todo_id, date) DO NOTHING
RETURNING *;

-- name: AddDailyTaskHistoryForDate :one
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date(?), ?, ?, ?
) 
ON CONFLICT(todo_id, date) DO NOTHING
RETURNING *;

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
-- Fix: uses ASC ordering for LAG so days_diff is always positive for consecutive days,
-- and anchors the current group on the latest entry instead of requiring today's date
-- to exist in the table (so the streak is not broken mid-day before the cron runs).
WITH ordered AS (
    SELECT
        date,
        completed,
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
        SUM(
            CASE
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
WHERE completed = 1
  AND streak_group = (SELECT streak_group FROM current_group);

-- name: GetStreakHistory :many
-- Returns all completed streak intervals ordered by length descending.
-- Fix: removed HAVING days_count = streak_length which silently dropped any streak
-- with an internal gap, and removed streak_length > 1 so single-day streaks appear.
WITH ordered AS (
    SELECT
        date,
        completed,
        julianday(date) - julianday(LAG(date) OVER (ORDER BY date ASC)) AS days_diff
    FROM todos_history
    WHERE todo_id = ?
    ORDER BY date ASC
),
grouped AS (
    SELECT
        date,
        completed,
        SUM(
            CASE
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
    WHERE completed = 1
    GROUP BY streak_id
)
SELECT
    start_date,
    end_date,
    streak_length
FROM streaks
ORDER BY streak_length DESC, start_date DESC;
