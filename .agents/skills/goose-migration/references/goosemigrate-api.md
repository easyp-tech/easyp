# goosemigrate API Reference

## Package

```
internal/database/goosemigrate
```

## Public API

### `Up(ctx context.Context, dsn string) error`

Applies all pending migrations to the database.

**Parameters:**
- `ctx` — context for cancellation and timeout
- `dsn` — PostgreSQL connection string (e.g., `postgres://user:pass@host:5432/db?sslmode=disable`)

**Behavior:**
1. Opens a new `*sql.DB` connection using `database/sql`
2. Pings the database to verify connectivity
3. Extracts embedded migrations via `fs.Sub(migrationsFS, "migrations")`
4. Creates a PostgreSQL session locker (`pg_advisory_lock`)
5. Creates a goose Provider with PostgreSQL dialect and session locker
6. Runs all pending Up migrations
7. Closes the database connection (with error logging)

**Error wrapping chain:**
```
sql.Open: <err>
db.PingContext: <err>
fs.Sub: <err>
lock.NewPostgresSessionLocker: <err>
goose.NewProvider: <err>
provider.Up: <err>
```

## Source

```go
// Package goosemigrate provides database migration using embedded SQL files and goose.
package goosemigrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Up applies all pending migrations to the database at the given DSN.
// It uses PostgreSQL advisory locking to prevent concurrent migration execution.
// Migrations are embedded via embed.FS and applied in a transaction per migration.
func Up(ctx context.Context, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("db.Close", "error", err)
		}
	}()

	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("db.PingContext: %w", err)
	}

	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("fs.Sub: %w", err)
	}

	sessionLocker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("lock.NewPostgresSessionLocker: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations,
		goose.WithSessionLocker(sessionLocker),
	)
	if err != nil {
		return fmt.Errorf("goose.NewProvider: %w", err)
	}

	_, err = provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("provider.Up: %w", err)
	}

	return nil
}
```

## Usage in cmd/main.go

```go
err := goosemigrate.Up(ctx, cfg.Database.DSN)
if err != nil {
    return fmt.Errorf("goosemigrate.Up: %w", err)
}
```

## Dependencies

| Import | Purpose |
|--------|---------|
| `database/sql` | Standard SQL database interface |
| `embed` | Compile-time file embedding |
| `io/fs` | Filesystem abstraction (`fs.Sub`) |
| `log/slog` | Structured logging for deferred close errors |
| `github.com/pressly/goose/v3` | Migration engine and Provider API |
| `github.com/pressly/goose/v3/lock` | PostgreSQL advisory session locker |
| `github.com/lib/pq` | PostgreSQL driver (blank import in `cmd/main.go`) |

## Migration Files

Located in `internal/database/goosemigrate/migrations/`.

Current migrations:

| File | Description |
|------|-------------|
| `00001_init.sql` | Initial schema: `plugins` table (with GIN index on tags) + `audit_log` table (with indexes on created_at and operation_type) |
