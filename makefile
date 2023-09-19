
MIGRATION_NAME = "none"
DATABASE_URL = "postgres://postgres:1324@postgres/postgres"

create:
	cd schema/migrations/ && goose create $(MIGRATION_NAME) sql

migrate:
	cd schema/migrations/ && goose postgres $(DATABASE_URL) up

down:
	cd schema/migrations/ && goose postgres $(DATABASE_URL) down

run:
	go run cmd/api/main.go

build:
	docker-compose build
	docker-compose up

rebuild:
	docker-compose down
	docker-compose up --build

restart:
	docker-compose restart app

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	del coverage.out

swaginit:
	swag init --parseDependency --parseInternal --parseDepth 2 -g cmd\api\server.go
