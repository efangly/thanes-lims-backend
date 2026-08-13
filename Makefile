include .env
export

MIGRATIONS_DIR=migrations

.PHONY: migrate-up migrate-down migrate-create migrate-force run-api run-seed test test-integration build swagger

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

# Requires a running Docker daemon - spins up disposable Postgres containers
# via testcontainers-go for each repository test.
test-integration:
	go test -tags=integration ./...

build:
	go build -o bin/api ./cmd/api
	go build -o bin/seed ./cmd/seed

# Regenerates docs/ from the @-annotations on handlers (internal/adapters/http/**/handler.go)
# and cmd/api/main.go. Commit the regenerated docs/ - it's imported by cmd/api and must be
# present for `go build` to succeed without running swag first.
swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
