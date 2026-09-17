<!-- scope: cmd/*, internal/*, pkg/*, src/*, app/*, domain/* -->
# Core Documentation Template

## What This Generates

- `.spec/ARCHITECTURE.md` — project architecture overview
- `.spec/PACKAGES.md` — reference of all packages/modules
- `.spec/DOMAIN.md` — domain model description
- `.spec/CODE_STYLE.md` — project-specific code conventions

## Instructions

You are a technical documentarian. Create the core project documentation in the `.spec/` directory.
Analyze the source code, directory structure, dependencies, and configuration files.

Create the following files:

---

### File 1: .spec/ARCHITECTURE.md

A document describing the project's architecture.

#### Structure:

##### 1. Overview
- One sentence: application type + architectural pattern
- ASCII diagram of application layers (top-down: Transport → Application → Adapters → Infrastructure)
- Show connections between layers with arrows

##### 2. Component Deep Dive
For each architectural layer:
- Name and package path
- Key files with descriptions (table: File | Description)
- Role in the system
- Which interfaces it implements / uses

Analyze the actual code:
- Entry point (`main.go` / `main.ts` / `app.py` etc.)
- Business logic (`app/`, `domain/`, `services/`, `use_cases/`)
- Adapters (`adapters/`, `repositories/`, `infrastructure/`)
- API layer (`api/`, `handlers/`, `controllers/`, `routes/`)

##### 3. Directory Structure
ASCII tree with comments (2-3 levels deep):
```
project/
├── cmd/                    # Entry points
│   └── app/
│       └── internal/
│           ├── app/        # Business logic
│           ├── adapters/   # Infrastructure adapters
│           └── api/        # API handlers
├── internal/               # Shared packages
└── configs/                # Configuration files
```

##### 4. Key Design Decisions
Numbered list of key architectural decisions (3-7 items):
- Architectural pattern and why
- API strategy (REST/gRPC/GraphQL)
- Dependency injection approach
- Testing strategy
- Observability approach (if present)
- Context propagation (if present)

For each decision: title, 2-3 bullet points, code example if appropriate.

##### 5. Data Flow
Trace the lifecycle of a typical request from entry point to response:
- ASCII sequence diagram showing how a request passes through layers
- Name the concrete types/functions involved at each step
- Show where validation, authorization, business logic, and persistence happen

Example format:
```
HTTP Request
  → Router (api/router.go)
    → Middleware (auth, logging, tracing)
      → Handler (api/handler.go)
        → App method (app/service.go)
          → Adapter (adapters/repo.go)
            → Database
          ← Return domain object
        ← Convert to response
      ← Send HTTP response
```

If the project has multiple entry points (HTTP + gRPC, CLI + API), show the primary one and note differences for others.

---

### File 2: .spec/PACKAGES.md

Reference of all project packages/modules.

#### Structure:

For each package/module:

```markdown
### `path/to/package`
**Brief description** — one sentence.

| File | Description |
|------|-------------|
| `file1.go` | Description |
| `file2.go` | Description |

Key details (if applicable):
- Which interfaces it implements
- Which libraries it uses
- Specifics (tracing, caching, etc.)
```

Group packages by layer:
1. **Application Layer** — business logic
2. **Adapters Layer** — infrastructure adapters
3. **API Layer** — transport
4. **Shared/Internal** — common utilities
5. **Generated** — generated code (proto, OpenAPI, etc.)

---

### File 3: .spec/DOMAIN.md

Description of the project's domain model.

#### Structure:

##### 1. Core Entities
For each entity:
- Name
- Structure definition (in the project's language)
- Field comments (if non-obvious)

##### 2. Enums / Value Objects
All enumerations and value objects with their values.

##### 3. Business Errors
> Owned by `errors.md` → `ERRORS.md`. Do **not** duplicate the catalog here.

One-line summary: "For the full business error catalog (codes, HTTP/gRPC mapping, retry policy) see `ERRORS.md`."
If `ERRORS.md` does not exist yet, list errors inline and note that they should be migrated when `errors.md` is run.

##### 4. Search/Filter Parameters
Search/filter structures (if present).

Important:
- Take definitions from actual code (`domain.go`, `models.py`, `types.ts`, etc.)
- Do not invent fields — only what exists in the code
- For large structures, show all fields

---

### File 4: .spec/CODE_STYLE.md

Project-specific code conventions.

#### Structure:

##### 1. Layer Structure
ASCII diagram of layers with rules for each:
- Which structures are used (public/private, with/without tags)
- How conversion between layers works

##### 2. Struct Tags by Layer (if applicable)
For each layer:
- Rule
- Code example (Correct)
- Explanation (Why)

##### 3. Conversion Pattern
How data is converted between layers:
- Adapters ↔ App (convert methods)
- API ↔ App (toX functions)

##### 4. Naming Conventions
Tables:
- Files: file naming patterns by layer
- Structs: visibility (public/private) by layer
- Enums: definition pattern

##### 5. Interface Conventions
Rules for interfaces:
- Argument order (context first, etc.)
- Error documentation
- Interface composition

##### 6. Import Ordering
Import order with example.

##### 7. Test File Organization
Table of test file patterns.

##### 8. Quick Reference
Summary table: Aspect | App Layer | Adapters Layer | API Layer

##### 9. Error Propagation
Rules for error handling across layers:
- How errors are wrapped when crossing layer boundaries
- Sentinel errors vs typed errors: when to use each
- Error types by layer (domain errors, adapter errors, API errors)
- Code example showing error wrapping chain from adapter → app → API

##### 10. Logging Conventions
Structured logging rules:
- Logger type and initialization pattern
- What to log at each layer (table: Layer | Log Level | What to Log)
- Sensitive data handling (what must never be logged)
- Correlation IDs / request IDs (how passed through context)
- Code example of a properly logged operation

##### 11. Concurrency Patterns (if applicable)
Concurrency rules used in the project:
- Goroutine / async task management (how spawned, how tracked)
- Context cancellation propagation
- Channel / queue usage patterns
- Mutex / synchronization patterns
- Graceful shutdown sequence

If the project is single-threaded or has no concurrency patterns, skip this section.

## General Rules

- Language: English
- Format: Markdown with ASCII diagrams, tables, fenced code blocks with language tags
- Code examples: ONLY from the actual project, not abstract
- If something is not found in the project — do not create the corresponding section
- Each document must be self-contained (can be read independently)
- Do not duplicate information between documents — cross-reference instead
- After creating files, update `.spec/README.md`: add links to created documents in the Documentation Index section
