-- +goose Up 
ALTER TABLE todos
ADD COLUMN expires_at TIMESTAMP NOT NULL; 

-- +goose Down
ALTER TABLE todos 
DROP COLUMN expires_at;
