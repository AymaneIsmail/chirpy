-- +goose Up
ALTER TABLE users ADD COLUMN is_chirpy_red BOOLEAN DEFAULT FALSE;
UPDATE users SET is_chirpy_red = FALSE WHERE hashed_password IS NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN is_chirpy_red;
