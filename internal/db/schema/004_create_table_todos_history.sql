-- +goose Up 
CREATE TABLE IF NOT EXISTS todos_history (
  id UUID PRIMARY KEY, 
  todo_id UUID REFERENCES todos(id),
  date DATE, 
  completed BOOLEAN DEFAULT false, 
  completed_at TIMESTAMP, 
  notes TEXT, 
  UNIQUE(todo_id, date)
);

-- +goose Down 
DROP TABLE IF EXISTS todos_history;
