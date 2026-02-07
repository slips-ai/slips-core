.PHONY: all proto sqlc build run clean docker-up docker-down db-create migrate-up migrate-down migrate-hash migrate-validate migrate-status migrate-new migrate-diff test grpcurl-smoke

# Tools
BUF_VERSION := 1.28.1
SQLC_VERSION := 1.25.0
ATLAS_VERSION := 1.1.0

ATLAS := $(HOME)/go/bin/atlas

all: proto sqlc build

# Install tools
tools:
	@echo "Installing tools..."
	@go install github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION)
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@v$(SQLC_VERSION)
	@curl -sSf https://atlasgo.sh | ATLAS_VERSION=v$(ATLAS_VERSION) sh -s -- \
		--community --no-install -y -o "$(ATLAS)"
	@chmod +x "$(ATLAS)"

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

# grpcurl smoke tests (requires slips-core running)
grpcurl-smoke:
	@./scripts/grpcurl-smoke.sh

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

# Create databases inside the local Postgres container.
# This is intentionally manual (explicit command) so you can manage multiple DBs (e.g. slips + identra).
DBS ?= slips identra
db-create:
	@echo "Creating databases (if missing): $(DBS)"
	@for db in $(DBS); do \
		docker exec -i postgres createdb -U postgres "$$db" 2>/dev/null || true; \
	done

# Database migrations
# Database URL can be set via DB_URL environment variable
# Default (overridable via DB_USER / DB_PASSWORD): postgres://postgres:postgres@localhost:5432/slips?sslmode=disable
DB_URL ?= postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@localhost:5432/slips?sslmode=disable

# Atlas uses a dev database URL for planning some operations (e.g. migrate diff/down).
# This defaults to a disposable docker-based Postgres instance. Override if needed.
DEV_DB_URL ?= docker://postgres/18/dev?search_path=public

# Keep migrations in the existing golang-migrate file naming format.
MIGRATIONS_DIR ?= file://migrations?format=golang-migrate

# Optional baseline version for first-time switch from golang-migrate -> Atlas.
# Example: make migrate-up DB_BASELINE=004
DB_BASELINE ?=

migrate-up:
	@echo "Running migrations (Atlas)..."
	@$(ATLAS) migrate apply \
		--config "file://atlas.hcl" \
		--env local \
		--var url="$(DB_URL)" \
		--var dev_url="$(DEV_DB_URL)" \
		$(if $(DB_BASELINE),--baseline "$(DB_BASELINE)",)

migrate-down:
	@echo "Reverting migrations (Atlas)..."
	@$(ATLAS) migrate down \
		--config "file://atlas.hcl" \
		--env local \
		--var url="$(DB_URL)" \
		--var dev_url="$(DEV_DB_URL)"

migrate-status:
	@$(ATLAS) migrate status \
		--config "file://atlas.hcl" \
		--env local \
		--var url="$(DB_URL)" \
		--var dev_url="$(DEV_DB_URL)"

migrate-hash:
	@$(ATLAS) migrate hash --dir "$(MIGRATIONS_DIR)"

migrate-validate:
	@$(ATLAS) migrate validate --dir "$(MIGRATIONS_DIR)"

migrate-new:
	@$(ATLAS) migrate new --dir "$(MIGRATIONS_DIR)" --edit

# Generate a new migration automatically from a desired schema.
# Usage example:
#   make migrate-diff MIGRATE_NAME=add_table MIGRATE_TO=file://schema.hcl
MIGRATE_NAME ?=
MIGRATE_TO ?=
migrate-diff:
	@test -n "$(MIGRATE_NAME)" || (echo "MIGRATE_NAME is required" && exit 1)
	@test -n "$(MIGRATE_TO)" || (echo "MIGRATE_TO is required (e.g. file://schema.hcl)" && exit 1)
	@$(ATLAS) migrate diff "$(MIGRATE_NAME)" \
		--dir "$(MIGRATIONS_DIR)" \
		--to "$(MIGRATE_TO)" \
		--dev-url "$(DEV_DB_URL)"

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
