-- +goose Up
ALTER TABLE users
    ALTER COLUMN password SET DATA TYPE VARCHAR(250),
ALTER COLUMN password SET NOT NULL;

-- +goose Down
