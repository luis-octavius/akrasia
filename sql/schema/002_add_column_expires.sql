-- +goose Up 
ALTER TABLE todos
ADD COLUMN expires_at TIMESTAMP DEFAULT NULL; 

-- +goose Down
ALTER TABLE todos 
DROP COLUMN expires_at;
