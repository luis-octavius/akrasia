-- +goose Up
-- Subtasks: a task can have a parent (one level only, enforced in code).
-- Names become unique per parent instead of across every task, so two
-- parents can each have a "Buy ingredients". SQLite can't drop the old
-- UNIQUE on name, so the table is rebuilt; ids are kept, so history rows
-- stay linked.
CREATE TABLE todos_new (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_at DATE NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  concluded BOOLEAN NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  priority TEXT NOT NULL DEFAULT 'medium',
  is_daily BOOL NOT NULL DEFAULT false,
  history_since DATE,
  parent_id UUID REFERENCES todos(id)
);

INSERT INTO todos_new (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since)
SELECT id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since
FROM todos;

DROP TABLE todos;

ALTER TABLE todos_new RENAME TO todos;

-- NULLs never collide in a UNIQUE index, so top-level names need their own
CREATE UNIQUE INDEX todos_top_level_name ON todos (name) WHERE parent_id IS NULL;
CREATE UNIQUE INDEX todos_subtask_name ON todos (parent_id, name) WHERE parent_id IS NOT NULL;

-- +goose Down
-- Subtasks are lost: names must be unique again across every task.
CREATE TABLE todos_old (
  id UUID PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  created_at DATE NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  concluded BOOLEAN NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  priority TEXT NOT NULL DEFAULT 'medium',
  is_daily BOOL NOT NULL DEFAULT false,
  history_since DATE
);

INSERT INTO todos_old (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since)
SELECT id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since
FROM todos
WHERE parent_id IS NULL;

DROP TABLE todos;

ALTER TABLE todos_old RENAME TO todos;
