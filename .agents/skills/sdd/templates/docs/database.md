<!-- scope: **/migrations/*, **/schema*, **/*repo*, **/*store* -->
# Database Documentation Template

## What This Generates

- `.spec/DATABASE.md` — database schema, migrations, connection management, and query patterns

## Instructions

You are a technical documentarian. Create database documentation for the project in the `.spec/` directory.
Analyze: migration files, ORM config, repository/DAO layer, database drivers, connection setup, seed scripts.

### Step 1: Identify Database Usage

Determine which databases the project uses:
- **SQL**: PostgreSQL, MySQL, SQLite, CockroachDB
- **NoSQL**: MongoDB, DynamoDB, Cassandra, Redis (if used as primary store)
- **Search**: Elasticsearch, Meilisearch, Typesense

For each database, identify:
- Driver / ORM / query builder (e.g., `pgx`, `sqlx`, `GORM`, `Prisma`, `TypeORM`, `sqlc`)
- Migration tool (e.g., `goose`, `migrate`, `knex`, `Flyway`, `Alembic`)
- Connection configuration (env vars, config files)

### Step 2: Create .spec/DATABASE.md

#### Structure:

##### 1. Overview
- One sentence: which database(s) and why
- Driver / ORM used
- Connection string format (with placeholders, no real credentials)

##### 2. Schema Overview

**Entity-Relationship Diagram:**
ASCII or Mermaid ER diagram showing tables and relationships:
```
┌──────────┐     ┌──────────────┐     ┌──────────┐
│  users   │────<│ user_tokens  │     │  posts   │
│──────────│     │──────────────│     │──────────│
│ id (PK)  │     │ id (PK)      │     │ id (PK)  │
│ email    │     │ user_id (FK) │  ┌──│ user_id  │
│ name     │     │ token        │  │  │ title    │
└──────────┘     └──────────────┘  │  └──────────┘
      │                            │
      └────────────────────────────┘
```

**Table Reference:**
| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `users` | User accounts | `id`, `email`, `name`, `created_at` |

For each table: name, purpose, key columns. Group by domain area if many tables.

##### 3. Migration Strategy
- Migration tool and version
- Migration directory path
- Naming convention (e.g., `YYYYMMDDHHMMSS_description.sql`, sequential numbers)
- Up / down convention: are down migrations required?
- How to create a new migration (command)
- How to run migrations (command)
- How to rollback (command)
- CI integration: are migrations run automatically in CI/CD?

##### 4. Connection Management
- Connection pool configuration (min/max connections, idle timeout)
- Read replicas (if applicable): how reads are routed
- Connection lifecycle: where the connection is created, how it's passed (context, DI, global)
- Health check / ping mechanism
- Reconnection strategy

Code reference: file path where connection is initialized.

##### 5. Query Patterns
- Repository / DAO pattern: interface definition and implementation structure
- Query builder vs raw SQL vs ORM: which approach is used and when
- Transaction management: how transactions are started, committed, rolled back
- Prepared statements / named queries (if applicable)
- Pagination pattern (offset vs cursor)

Code example of a typical repository method from the project.

##### 6. Seed / Fixtures
- Dev seed data: how to populate the database for local development (command)
- Test fixtures: how test data is set up (helpers, factories, SQL files)
- Reset command: how to wipe and re-seed the database
- Sample data location (if seed files exist)

##### 7. Indexes & Performance (if discoverable)
Notable indexes beyond primary keys:
| Table | Index | Columns | Type |
|-------|-------|---------|------|

Only include if indexes are defined in migration files or ORM annotations.

##### 8. Backup & Recovery (if applicable)
- Backup strategy (pg_dump, cloud snapshots, scheduled jobs)
- Backup commands or scripts
- Point-in-time recovery (WAL archiving, binlog)
- Data retention policy

Only include if backup tooling or scripts exist in the project.

## General Rules

- Language: English
- All SQL examples must come from actual migration files or repository code
- Do not invent tables or columns — only document what exists
- If the project uses multiple databases, create sections for each
- Connection strings and credentials must use placeholders, never real values
- If the project has no database, do not generate this file — skip entirely
- After creating, update `.spec/README.md`: add a link under the appropriate section
