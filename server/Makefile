.PHONY: help install-tools generate run test test-race build clean reset-db

help:
	@echo "Available commands:"
	@echo "  make install-tools    - Install required tools (sqlc, etc.)"
	@echo "  make generate         - Generate code from SQL (sqlc)"
	@echo "  make run              - Run the server"
	@echo "  make test             - Run tests"
	@echo "  make test-race        - Run tests with race detector"
	@echo "  make build            - Build binary for production"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make reset-db         - Reset database (delete and recreate)"

install-tools:
	@echo "Installing sqlc..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

generate:
	@echo "Generating code from SQL..."
	sqlc generate

run: generate
	@echo "Running server..."
	go run ./cmd/api

test:
	@echo "Running tests..."
	go test ./...

test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

build: generate
	@echo "Building binary..."
	go build -o yourownboss-api ./cmd/api
	@echo "Binary created: yourownboss-api"

clean:
	@echo "Cleaning build artifacts..."
	rm -f yourownboss-api
	rm -f yourownboss-api.exe
	go clean

reset-db:
	@echo "Resetting database..."
	rm -f yourownboss.db
	rm -f yourownboss.db-shm
	rm -f yourownboss.db-wal
	@echo "Database reset complete"
