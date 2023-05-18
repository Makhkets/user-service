
MIGRATION_NAME = "none"
DATABASE_URL = "postgres://postgres:1324@localhost/postgres"

create:
	cd schema/migrations/ && goose create $(MIGRATION_NAME) sql

migrate:
	cd schema/migrations/ && goose postgres $(DATABASE_URL) up

down:
	cd schema/migrations/ && goose postgres $(DATABASE_URL) down

run:
	go run cmd/api/main.go