-- +goose Up
-- Weekly and monthly tasks. recurrence is none, daily, weekly or monthly;
-- repeat_on is the weekday (1 = Monday .. 7 = Sunday) of a weekly task or
-- the day of the month (1..31) of a monthly one, and NULL for the others.
-- is_daily stays and now means "recurs" (true for all three kinds), so
-- everything that tells a recurring task from a one-off one keeps working.
ALTER TABLE todos ADD COLUMN recurrence TEXT NOT NULL DEFAULT 'none';
ALTER TABLE todos ADD COLUMN repeat_on INTEGER;

UPDATE todos SET recurrence = 'daily' WHERE is_daily = true;

-- +goose Down
ALTER TABLE todos DROP COLUMN repeat_on;
ALTER TABLE todos DROP COLUMN recurrence;
