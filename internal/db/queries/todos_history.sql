-- name: AddTodoHistory :one
-- Logs a completion for "today" (local time). Used by `done` for both
-- daily and one-off tasks. ON CONFLICT DO UPDATE means calling this
-- more than once on the same day always reflects the latest state —
-- there is no separate "reset" step that can race against it.
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date('now', 'localtime'), ?, ?, ?
) 
ON CONFLICT(todo_id, date) DO UPDATE SET
    completed = excluded.completed,
    completed_at = excluded.completed_at,
    notes = excluded.notes
RETURNING *;

-- name: IsDoneToday :one
SELECT EXISTS(
    SELECT 1 FROM todos_history
    WHERE todo_id = ? AND date = date('now', 'localtime') AND completed = 1
) AS done_today;

-- name: GetDoneTodayIDs :many
-- IDs of every task (daily or not) already completed today. Used to
-- render "today" views without relying on todos.concluded, which no
-- longer gets reset for daily tasks.
SELECT todo_id FROM todos_history
WHERE date = date('now', 'localtime') AND completed = 1;

-- name: BackfillDailyDone :one
-- Records a real completion for a past date the user forgot to log.
-- Unlike the old backfill, this always writes completed = true — there
-- is no "neutral" state. Dates before a task's history_since are simply
-- never considered by the streak queries below, so there is nothing
-- left to accidentally overwrite or misclassify.
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
    ?, ?, date(?), true, NULL, 'backfilled'
)
ON CONFLICT(todo_id, date) DO UPDATE SET
    completed = true,
    notes = 'backfilled'
RETURNING *;

-- name: GetCurrentStreak :one
-- Returns the number of consecutive completed days ending on the most
-- recent entry. Only dates from todos.history_since onward are
-- considered, so days before history tracking existed for this task
-- can never break (or pad) the streak.
WITH ordered AS (
    SELECT
        date,
        completed,
        julianday(date) - julianday(LAG(date) OVER (ORDER BY date ASC)) AS days_diff
    FROM todos_history
    WHERE todo_id = ?
      AND date >= (SELECT history_since FROM todos WHERE id = ?)
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
WITH ordered AS (
    SELECT
        date,
        completed,
        julianday(date) - julianday(LAG(date) OVER (ORDER BY date ASC)) AS days_diff
    FROM todos_history
    WHERE todo_id = ?
      AND date >= (SELECT history_since FROM todos WHERE id = ?)
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
