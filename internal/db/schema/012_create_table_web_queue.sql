-- +goose Up
-- ids of the offline-queue requests the web app already applied, so a
-- request replayed after a lost response is not applied twice
CREATE TABLE IF NOT EXISTS web_queue (
  id TEXT PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT (datetime('now', 'localtime'))
);

-- +goose Down
DROP TABLE IF EXISTS web_queue;
