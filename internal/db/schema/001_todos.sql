-- +goose Up 
CREATE TABLE IF NOT EXISTS todos (
  id UUID PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT, 
  created_at DATE NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  concluded BOOLEAN NOT NULL 
);

-- +goose Down 
DROP TABLE IF EXISTS todos;

