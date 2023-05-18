-- +goose Up
CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(20) NOT NULL,
        password VARCHAR(50) NOT NULL,
        is_admin BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now()
);


-- +goose Down
DROP TABLE users;