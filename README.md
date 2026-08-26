# journal-go

A Go-based journal management service that allows users to create, read, and manage journal entries with user authentication and multi-user access control.

## Features

- HTTP REST API server (runs on `0.0.0.0:3001` with context path `/servr`)
- User authentication via external auth provider (https://cbox.info/abax/loginNoToken)
- Session management with configurable token TTL
- Multi-user journal access with relation-based permissions (owner/reader)
- MySQL database persistence
- Structured logging with `log/slog`

## Prerequisites

- Go 1.25 or later
- MySQL 8.0 or later
- MySQL server running locally on `localhost` with database `journal`
  - Default credentials: user `jou`, password `jou`

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd journal-go
```

2. Install Go dependencies:
```bash
go mod download
```

3. Initialize the database:
```bash
mysql -u jou -p < docs/db/ddl.sql
# Password: jou
```

## Running the Application

### Build and Run

```bash
# Build the application
go build -o journal-go .

# Run the application
./journal-go
```

Or run directly:
```bash
go run main.go
```

The server will start on `0.0.0.0:3001` with all API endpoints available under the `/servr` context path.

## Configuration

Configuration is read from `docs/config/config.json`:

```json
{
  "auth": "https://cbox.info/abax/loginNoToken",
  "tokenTimeToLive": 3600,
  "access": "authorization-scope",
  "db": {
    "host": "localhost",
    "port": 3306,
    "user": "jou",
    "password": "jou",
    "database": "journal"
  },
  "serverAddress": "0.0.0.0:3001",
  "context": "/servr"
}
```

## Database Schema

The database schema includes the following tables:

- `user` — User identity from auth provider
- `session` — Active user sessions with tokens
- `journal` — Journal collections owned by users
- `journal_item` — Individual entries within a journal
- `user_journal` — Relation table defining user access to journals
- `relation` — Lookup table for relationship types (owner/reader)

See `docs/db/ddl.sql` for the complete schema definition.

## Development

### Code Quality

Format code using standard Go formatting:
```bash
gofmt -s -w .
```

Run linting (if golangci-lint is installed):
```bash
golangci-lint run ./...
```

Check for suspicious constructs:
```bash
go vet ./...
```

### Testing

Run all tests:
```bash
go test ./...
```

Run tests in a specific package:
```bash
go test ./internal/service
```

Run a single test:
```bash
go test -run TestFunctionName ./internal/service
```

Verbose test output:
```bash
go test -v ./...
```

With coverage:
```bash
go test -cover ./...
```

## Architecture

The codebase follows a layered architecture pattern:

- **`main.go`** — Application entry point that initializes and starts the service
- **`internal/service/`** — Service layer (`start.go`) — orchestrates startup, initializes connections and handlers
- **`internal/controller/`** — Controller layer (`proc.go`) — HTTP request handlers and business logic (e.g., authentication)
- **`internal/data/`** — Data layer (`conf.go`) — database models and query builders

## Key Files

| Path | Purpose |
|------|---------|
| `main.go` | Entry point |
| `internal/service/start.go` | Application initialization and service startup |
| `internal/controller/proc.go` | Request processing and authentication |
| `internal/data/conf.go` | Data models and database operations |
| `docs/db/ddl.sql` | Database schema definition |
| `docs/config/config.json` | Runtime configuration |

## License

This project is licensed under the GNU General Public License v2.0. See the `LICENSE` file for details.
