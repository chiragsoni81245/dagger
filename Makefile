# Paths and configurations
MIGRATE_BIN := $(shell which migrate) # Path to the migrate binary
DB_URL := "postgres://test:12345@localhost:5432/dagger?sslmode=disable"
MIGRATIONS_DIR := ./migrations


build:
	@go build -o bin/dagger cmd/dagger/main.go

run: build
	@./bin/dagger

test:
	@go test ./... -v --race

# Database migrations
migrate:
	@$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database $(DB_URL) $(CMD)

migrate-up:
	@make migrate CMD=up

migrate-down:
	@make migrate CMD=down
