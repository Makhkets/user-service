-- +goose Up
CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(20) UNIQUE NOT NULL,
        password VARCHAR(200) NOT NULL,
        is_admin BOOLEAN NOT NULL DEFAULT false,
        is_banned BOOL NOT NULL DEFAULT false;
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now()
);


-- +goose Down
DROP TABLE users;