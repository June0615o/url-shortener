.PHONY: run build test clean dev docker-up docker-down

# Build the binary
build:
	go build -o server ./cmd/server

# Run locally (requires PostgreSQL and Redis running)
run: build
	./server

# Run tests
test:
	go test ./... -v -cover

# Run tests with race detection
test-race:
	go test ./... -v -race

# Development with hot reload (requires air)
dev:
	air

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

docker-logs:
	docker compose logs -f

# Clean binary
clean:
	rm -f server
	go clean -cache

# Format code
fmt:
	go fmt ./...

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Database
db-migrate:
	psql "$(DB_DSN)" -f migrations/001_init.up.sql

db-reset:
	psql "$(DB_DSN)" -f migrations/001_init.down.sql
	psql "$(DB_DSN)" -f migrations/001_init.up.sql
