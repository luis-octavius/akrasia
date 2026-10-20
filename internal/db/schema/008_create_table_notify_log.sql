-- +goose Up
-- one row per reminder slot already sent (or skipped with nothing to say)
CREATE TABLE IF NOT EXISTS notify_log (
  day TEXT NOT NULL,
  slot TEXT NOT NULL,
  sent_at TIMESTAMP DEFAULT (datetime('now', 'localtime')),
  PRIMARY KEY (day, slot)
);

-- +goose Down
DROP TABLE IF EXISTS notify_log;
