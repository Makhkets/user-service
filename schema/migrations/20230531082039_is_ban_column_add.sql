-- +goose Up
ALTER TABLE users ADD COLUMN is_banned BOOL NOT NULL DEFAULT false;


-- +goose Down

