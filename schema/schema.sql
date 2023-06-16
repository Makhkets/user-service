-- +goose Up
CREATE TABLE users (
           id SERIAL PRIMARY KEY,
           username VARCHAR(20) UNIQUE NOT NULL,
           password VARCHAR(200) NOT NULL,
           is_banned BOOL NOT NULL DEFAULT false,
           status VARCHAR(20) NOT NULL DEFAULT 'user',
           created_at TIMESTAMP NOT NULL DEFAULT now(),
           updated_at TIMESTAMP NOT NULL DEFAULT now(),
           CONSTRAINT check_status CHECK (status IN ('user', 'moderator', 'admin'))
);


-- +goose Down
DROP TABLE users;