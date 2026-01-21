.PHONY: all proto sqlc build run clean docker-up docker-down migrate-up migrate-down test

# Tools
BUF_VERSION := 1.28.1
SQLC_VERSION := 1.25.0
MIGRATE_VERSION := 4.17.0

all: proto sqlc build

# Install tools
tools:
	@echo "Installing tools..."
	@go install github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION)
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@v$(SQLC_VERSION)
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v$(MIGRATE_VERSION)

# Generate proto code
proto:
	@echo "Generating proto code..."
	@$(HOME)/go/bin/buf generate

# Generate sqlc code
sqlc:
	@echo "Generating sqlc code..."
	@$(HOME)/go/bin/sqlc generate

# Build the application
build: proto sqlc
	@echo "Building application..."
	@go build -o bin/slips-core cmd/server/main.go

# Run the application
run: build
	@echo "Running application..."
	@./bin/slips-core

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ gen/

# Docker commands
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

# Database migrations
# Database URL can be set via DB_URL environment variable
# Default: postgres://postgres:postgres@localhost:5432/slips?sslmode=disable
DB_URL ?= postgres://postgres:postgres@localhost:5432/slips?sslmode=disable

migrate-up:
	@echo "Running migrations..."
	@migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path migrations -database "$(DB_URL)" down

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
