-- name: AddTodoHistory :one 
INSERT INTO todos_history (id, todo_id, date, completed, completed_at, notes)
VALUES (
  ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetCurrentStreak :one
WITH daily_completions AS (
    SELECT 
        date,
        completed,
        date as ref_date, 
        julianday(date) - julianday(LAG(date) OVER (ORDER BY date DESC)) as days_diff
    FROM todos_history
    WHERE todo_id = ?
    ORDER BY date DESC
    LIMIT 100
),
streak_calc AS (
    SELECT 
        date,
        completed,
        days_diff,
        CASE 
            WHEN completed = 0 THEN 0 
            WHEN days_diff > 1 THEN 0
            ELSE 1
        END as is_streak_day,
        SUM(CASE 
            WHEN completed = 0 OR days_diff > 1 THEN 1 
            ELSE 0 
        END) OVER (ORDER BY date DESC ROWS UNBOUNDED PRECEDING) as streak_group
    FROM daily_completions
    WHERE date <= date('now')
)
SELECT 
    COUNT(*) as current_streak
FROM streak_calc
WHERE 
    completed = 1 
    AND streak_group = (
        SELECT streak_group 
        FROM streak_calc 
        WHERE date = date('now')
        LIMIT 1
    )
AND date <= date('now');

-- name: GetStreakHistory :many
WITH streak_groups AS (
    SELECT 
        date,
        completed,
        SUM(CASE 
            WHEN completed = 0 OR 
                 julianday(date) - julianday(LAG(date) OVER (ORDER BY date)) > 1 
            THEN 1 ELSE 0 
        END) OVER (ORDER BY date) as streak_id
    FROM todos_history
    WHERE todo_id = ?
    ORDER BY date
),
streaks AS (
    SELECT 
        streak_id,
        MIN(date) as start_date,
        MAX(date) as end_date,
        COUNT(*) as streak_length,
        julianday(MAX(date)) - julianday(MIN(date)) + 1 as days_count
    FROM streak_groups
    WHERE completed = 1
    GROUP BY streak_id
    HAVING days_count = streak_length
)
SELECT 
    start_date,
    end_date,
    streak_length
FROM streaks
WHERE streak_length > 1
ORDER BY streak_length DESC, start_date DESC;
