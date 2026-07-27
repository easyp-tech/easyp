---
name: goose-migration
description: "PostgreSQL migration management with goose v3. Use when: adding new database tables or columns, modifying schema, creating migration files, debugging migration errors, understanding the migration system architecture."
argument-hint: "Describe the schema change you want to make (e.g. 'add users table', 'add column to plugins')"
---

# Goose Migration — EasyP Service

You manage PostgreSQL schema migrations for the EasyP Service using goose v3 with the Provider API and embedded SQL files.

## When to Use

- Adding new database tables
- Adding or modifying columns
- Creating or modifying indexes
- Writing `-- +goose Down` rollback logic
- Debugging migration errors
- Understanding the migration architecture

## Architecture

```
internal/database/goosemigrate/
├── goosemigrate.go              # Up(ctx, dsn) function — Provider API + advisory lock
└── migrations/
    └── 00001_init.sql           # Numbered SQL migration files
```

**Key design decisions:**
- **Provider API** (not Functional API) — encapsulates state, supports session locking
- **`pg_advisory_lock`** via `lock.NewPostgresSessionLocker()` — prevents concurrent migrations
- **`embed.FS`** — migrations compiled into the binary, no external files needed
- **`fs.Sub(migrationsFS, "migrations")`** — required to strip the embedded directory prefix
- **Called at startup** from `cmd/main.go` before servers start

For implementation details, see [goosemigrate-api.md](./references/goosemigrate-api.md).

## Creating a New Migration

### Step 1: Determine the next number

```bash
ls internal/database/goosemigrate/migrations/
# 00001_init.sql
# → next is 00002
```

Always use **5-digit zero-padded** numbers: `00001`, `00002`, ..., `99999`.

### Step 2: Create the file

```bash
touch internal/database/goosemigrate/migrations/00002_<descriptive_name>.sql
```

Name should describe the change: `00002_add_users_table.sql`, `00002_add_plugin_description.sql`.

### Step 3: Write the migration

```sql
-- +goose Up
CREATE TABLE users
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    email      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (email),
    PRIMARY KEY (id)
);

CREATE INDEX idx_users_email ON users (email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Step 4: Verify

No code changes needed — `embed.FS` with `//go:embed migrations/*.sql` automatically picks up new files.

```bash
go build ./...   # Ensure it compiles
```

## SQL Conventions

### Table creation

```sql
CREATE TABLE table_name
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    -- columns...
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (col1, col2),     -- constraints INLINE
    PRIMARY KEY (id)          -- constraints INLINE
);
```

- **PRIMARY KEY, UNIQUE, CHECK, FOREIGN KEY** — define inline in `CREATE TABLE`
- **Indexes** — separate `CREATE INDEX` statements (cannot be inline)

### Indexes

```sql
-- Regular index
CREATE INDEX idx_<table>_<column> ON <table> (<column>);

-- GIN index (for arrays, JSONB)
CREATE INDEX idx_<table>_<column> ON <table> USING gin (<column>);

-- Conditional index
CREATE INDEX IF NOT EXISTS idx_<table>_<column> ON <table> (<column>);
```

GIN indexes **cannot** be defined inline in `CREATE TABLE`.

### Down migration

- **Reverse order** of Up operations
- **Always use `IF EXISTS`** for safety
- **Drop indexes before tables** if they reference the table

```sql
-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## Common Pitfalls

| Problem | Cause | Solution |
|---------|-------|---------|
| `no migrations found` | `embed.FS` includes subdirectory path | Use `fs.Sub(migrationsFS, "migrations")` |
| Concurrent migration corruption | No locking | Use `goose.WithSessionLocker(lock.NewPostgresSessionLocker())` |
| Global state leaks between tests | Using `goose.SetBaseFS()` / `goose.SetDialect()` | Use Provider API — state is encapsulated |
| GIN index inside CREATE TABLE | PostgreSQL doesn't support this | Separate `CREATE INDEX ... USING gin (...)` statement |
| Migration number conflict | Two developers create same number | Always check latest file before creating |
| `embed.FS` doesn't find files | Wrong `//go:embed` directive | Must be `//go:embed migrations/*.sql` — glob required |

## Integration Points

### Startup (cmd/main.go)

```go
err := goosemigrate.Up(ctx, cfg.Database.DSN)
if err != nil {
    return fmt.Errorf("goosemigrate.Up: %w", err)
}
```

Migration runs **before** gRPC/HTTP servers start, **after** config parsing.

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/pressly/goose/v3` | Migration engine |
| `github.com/pressly/goose/v3/lock` | PostgreSQL advisory locking |
| `github.com/lib/pq` | PostgreSQL driver (blank import in `cmd/main.go`) |

## Decision Flow

```
User wants to change database schema
│
├─ ADD table → Create new migration file (00002_...)
│
├─ ADD column → Create new migration file with ALTER TABLE
│
├─ ADD index → Create new migration file with CREATE INDEX
│
├─ MODIFY column → New migration with ALTER TABLE ALTER COLUMN
│   ⚠ Never modify existing migration files!
│
├─ DEBUG migration error → Check:
│   1. SQL syntax (run against psql directly)
│   2. Migration numbering (no gaps, no duplicates)
│   3. embed.FS directive (//go:embed migrations/*.sql)
│   4. fs.Sub usage
│
└─ ROLLBACK → Manual: connect to DB, run goose down
    (no automated rollback in current setup)
```

## Important Constraints

- **Never modify an existing migration** that has been applied — always create a new one
- **Never reorder migrations** — numbering determines execution order
- **Always include `-- +goose Down`** — even if it's empty, the marker must exist
- **Test SQL locally** against PostgreSQL before committing
- **One logical change per migration** — don't mix unrelated schema changes
