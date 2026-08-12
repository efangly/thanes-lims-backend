include .env
export

MIGRATIONS_DIR=migrations

.PHONY: migrate-up migrate-down migrate-create migrate-force run-api run-seed test build

migrate-up:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) down 1

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-force:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) force $(version)

run-api:
	go run ./cmd/api

run-seed:
	go run ./cmd/seed

test:
	go test ./...

build:
	go build -o bin/api ./cmd/api
	go build -o bin/seed ./cmd/seed
