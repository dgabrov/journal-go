# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**git** you may not run any git command, no git add no git commit, no nothing

**journal-go** is a Go-based journal management service that allows users to create, read, and manage journal entries. It integrates with an external authentication service (https://cbox.info/abax/loginNoToken) and uses MySQL for data persistence.

**Key characteristics:**
- HTTP server-based API (runs on 0.0.0.0:3001 by default with context path `/servr`)
- User authentication and session management
- Multi-user journal access with relation-based permissions (owner/reader)
- Go 1.25 module

## Architecture

The codebase follows a layered architecture pattern:

- **`main.go`**: Application entry point that calls `service.Start()`
- **`internal/service/`**: Service layer (`start.go`) — orchestrates startup, initializes connections and handlers
- **`internal/controller/`**: Controller layer (`proc.go`) — HTTP request handlers and business logic (e.g., authentication)
- **`internal/data/`**: Data layer (`conf.go`) — database models and query builders

**Key entities in the database:**
- `user` — user identity from auth provider
- `session` — active user sessions with tokens
- `journal` — journal collections owned by users
- `journal_item` — individual entries within a journal
- `user_journal` — relation table defining user access to journals (owns or reads)
- `relation` — lookup table for relationship types

See `docs/db/ddl.sql` for the complete schema.

## Development Setup

### Prerequisites
- Go 1.25+
- MySQL 8.0+ running locally on `localhost` with database `journal` (user: `jou`, password: `jou`)

### Building and Running

```bash
# Build the application
go build -o journal-go .

# Run the application
./journal-go

# Run with specific output
go run main.go
```

### Testing (when tests are added)

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./internal/service

# Run a single test
go test -run TestFunctionName ./internal/service

# Verbose test output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code (standard Go formatting)
gofmt -s -w .

# Lint the codebase (if golangci-lint is installed)
golangci-lint run ./...

# Run go vet to check for suspicious constructs
go vet ./...
```

## Database Setup

The MySQL database schema is defined in `docs/db/ddl.sql`. To initialize the database:

```bash
# Create the journal database and tables
mysql -u jou -p < docs/db/ddl.sql
# Password: jou
```

Database credentials are configured in `docs/config/config.json`.

## Configuration

The application reads its configuration from `docs/config/config.json`:

- `auth` — authentication service URL
- `tokenTimeToLive` — session token TTL in seconds
- `access` — authorization scope
- `db` — database connection parameters
- `serverAddress` — HTTP server bind address
- `context` — URL path prefix for all routes

## Key Files Reference

| Path | Purpose |
|------|---------|
| `main.go` | Entry point |
| `internal/service/start.go` | Application initialization and service startup |
| `internal/controller/proc.go` | Request processing and `CheckCredentials()` function |
| `internal/data/conf.go` | Data models (currently `ConfigData` stub) |
| `docs/db/ddl.sql` | Database schema definition |
| `docs/config/config.json` | Runtime configuration |

## Development Workflow

1. **Add new features:** Implement in the appropriate layer (`controller` for handlers, `service` for logic, `data` for DB queries)
2. **Extend schema:** Update `docs/db/ddl.sql` first, then migrate the local database
3. **Database migration:** Use raw SQL updates for now; consider migration tooling if the schema evolves frequently
4. **Authentication flow:** Passes through `controller.CheckCredentials()` which currently returns stubs
5. **Error handling:** Use `log/slog` (already imported in `main.go`) for structured logging

## Notes for Future Development

- **Early stage:** The project is still in its infancy; most functionality is stubbed out (service.Start() returns nil, CheckCredentials() returns nil values)
- **SQL migrations:** No migration framework is currently in place; manually manage schema changes against the local MySQL instance
- **Testing:** No test files exist yet — add `*_test.go` files following Go conventions
- **Dependencies:** The project has minimal dependencies as of now; check `go.mod` before adding external packages
- **errors** - please prefer the construct errors.Is rather than comparing references between errors

you have the tendency to use this construct:
```go
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
```

if the function only returns err, I believe the above can be replaced successfully with 
```go
    return tx.Commit()
```
