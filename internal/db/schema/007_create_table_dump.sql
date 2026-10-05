-- +goose Up
CREATE TABLE IF NOT EXISTS dump (
  id UUID PRIMARY KEY, 
  name TEXT NOT NULL, 
);

-- +goose Down 
DROP TABLE IF EXISTS dump;
