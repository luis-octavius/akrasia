
-- +goose Up
CREATE TABLE IF NOT EXISTS dump (
  id UUID PRIMARY KEY, 
  name TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT (datetime('now', 'localtime'))
);

-- +goose Down 
DROP TABLE IF EXISTS dump;
