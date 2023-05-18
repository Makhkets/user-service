-- +goose Up
ALTER TABLE users ADD COLUMN is_admin BOOL NOT NULL DEFAULT false;

-- +goose Down
DROP TABLE users;