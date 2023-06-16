-- +goose Up
-- Удаляем поле is_admin
ALTER TABLE users DROP COLUMN IF EXISTS is_admin;

-- Добавляем поле status
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'user';

-- Ограничение на значения status
ALTER TABLE users ADD CONSTRAINT check_status CHECK (status IN ('user', 'moderator', 'admin'));

-- Обновляем created_at и updated_at
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE users ALTER COLUMN updated_at SET DEFAULT now();


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
