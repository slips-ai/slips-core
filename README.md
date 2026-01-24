# slips-core

A Go-based task and tag management service built with clean architecture principles.

## Architecture

The project follows a three-layer clean architecture organized by feature:

- **Domain Layer**: Contains entities and repository interfaces
- **Application Layer**: Contains business logic and use cases
- **Infrastructure Layer**: Contains gRPC servers and database implementations

### Features

- Task management (CRUD operations)
- Tag management (CRUD operations)
- MCP Token authentication (UUID-based API tokens)

## Tech Stack

- **Language**: Go
- **gRPC**: Protocol Buffers with Buf
- **Database**: PostgreSQL with sqlc
- **Configuration**: Viper
- **Logging**: slog with tint
- **Tracing**: OpenTelemetry
- **Containerization**: Docker Compose

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Make

## Getting Started

### 1. Install development tools

```bash
make tools
```

### 2. Start infrastructure services

```bash
make docker-up
```

This starts:

- PostgreSQL on port 5432
- Jaeger (tracing UI) on port 16686

### 3. Run database migrations

```bash
make migrate-up
```

### 4. Generate code

```bash
make proto  # Generate gRPC code from proto files
make sqlc   # Generate Go code from SQL queries
```

### 5. Build and run the service

```bash
make build
make run
```

Or simply:

```bash
make all
```

## Development

### Project Structure

The project follows Go best practices with a clean architecture organized by feature:

```text
slips-core/
├── cmd/server/          # Main application entry point
├── internal/            # Private application code
│   ├── task/            # Task feature
│   │   ├── domain/      # Task entities and interfaces
│   │   ├── application/ # Task business logic
│   │   └── infra/       # Task infrastructure (gRPC, Postgres)
│   └── tag/             # Tag feature
│       ├── domain/      # Tag entities and interfaces
│       ├── application/ # Tag business logic
│       └── infra/       # Tag infrastructure (gRPC, Postgres)
├── pkg/                 # Shared packages (reusable libraries)
│   ├── config/          # Configuration loader
│   ├── logger/          # Logging setup
│   └── tracing/         # OpenTelemetry setup
├── api/                 # API definitions
│   └── proto/           # Protocol Buffer definitions
├── migrations/          # Database migrations
└── gen/                 # Generated code (gitignored)
```

### Available Make Commands

- `make tools` - Install development tools (buf, sqlc, migrate)
- `make proto` - Generate gRPC code from proto files
- `make sqlc` - Generate database code from SQL queries
- `make build` - Build the application
- `make run` - Run the application
- `make clean` - Clean build artifacts
- `make docker-up` - Start Docker services
- `make docker-down` - Stop Docker services
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback database migrations
- `make test` - Run tests
- `make fmt` - Format code
- `make tidy` - Tidy Go modules

## Configuration

Configuration can be provided via:

- `config.yaml` file
- Environment variables (prefix: `SLIPS_`)

Example configuration:

```yaml
server:
  grpc_port: 9090

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: slips
  sslmode: disable

tracing:
  enabled: true
  service_name: slips-core
  endpoint: localhost:4317
```

## Observability

### Tracing

The service includes OpenTelemetry tracing integration:

- Traces are exported to Jaeger via OTLP
- Access Jaeger UI at <http://localhost:16686>
- Tracing middleware is automatically applied to all gRPC calls

### Logging

Structured logging using Go's slog package:

- Development: Colorful console output with tint
- Production: JSON formatted logs

## API

The service exposes gRPC APIs for:

### MCP Token Service

- `CreateMCPToken` - Create a new MCP token for API access
- `GetMCPToken` - Get an MCP token by ID
- `ListMCPTokens` - List all MCP tokens for the authenticated user
- `RevokeMCPToken` - Revoke (deactivate) an MCP token
- `DeleteMCPToken` - Delete an MCP token

See [MCP Token Documentation](docs/MCP_TOKEN.md) for detailed usage.

### Task Service

- `CreateTask` - Create a new task
- `GetTask` - Get a task by ID
- `UpdateTask` - Update a task
- `DeleteTask` - Delete a task
- `ListTasks` - List tasks with pagination

### Tag Service

- `CreateTag` - Create a new tag
- `GetTag` - Get a tag by ID
- `UpdateTag` - Update a tag
- `DeleteTag` - Delete a tag
- `ListTags` - List tags with pagination

## License

See LICENSE file.
