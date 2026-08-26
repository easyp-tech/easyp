<!-- scope: * -->
# Bootstrap Documentation Template

## What This Generates

- `.spec/README.md` — project documentation index (Quick Facts, Project Structure, Running, Ports, Key Interfaces, Adding Features)
- `.spec/agent-rules.md` — mandatory rules for AI agents working on this project (Code Style, Naming, Error Handling, Testing, Dependencies, Formatting)

**Run this template first** — it creates the root index that all other templates reference and update.

## Instructions

You are a technical documentarian. Your task is to analyze the current project and create the foundational documentation in the `.spec/` directory for use by AI agents.

### Step 1: Project Analysis

Study the project structure. Pay attention to:
- Root files: `go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pom.xml`, etc. → determine language and framework
- Directory structure: `cmd/`, `src/`, `internal/`, `pkg/`, `app/`, `lib/`, etc.
- Configuration: `docker-compose.yml`, `Makefile`, `Taskfile.yml`, `.github/`, CI/CD
- API: proto files, OpenAPI/Swagger, GraphQL schemas
- Tests: `*_test.go`, `*.test.ts`, `test/`, `tests/`, `spec/`
- Infrastructure: `Dockerfile`, `kubernetes/`, `terraform/`, `configs/`
- Client applications: `clients/`, `frontend/`, `web/`, `mobile/`
- Dependencies: `go.sum`, `package-lock.json`, `poetry.lock`, etc.

### Step 2: Create .spec/README.md

Create `.spec/README.md` with the following structure:

#### Header
```markdown
# {Project Name} Documentation
This folder contains documentation to help LLMs and developers quickly understand the project context.
```

#### Documentation Index
Group documents by category. Use only categories relevant to the project:

- **Core** — Architecture, Packages, Domain, Code Style (always needed)
- **Development** — Tools, Testing, Files/Storage (always needed)
- **Auth & Security** — if the project has OAuth, JWT, API keys
- **Infrastructure** — if the project has Docker, monitoring, caching, queues
- **Clients** — if the project has frontend, mobile apps, bots

For each document, add a link: `- [DOC_NAME.md](./DOC_NAME.md) — brief description`

IMPORTANT: If documents have not been created yet, mark them as "TODO" or add a note that they will be created later.

#### Quick Facts
Table of key project technologies:

```markdown
| Aspect | Technology |
|--------|------------|
| **Language** | ... |
| **Architecture** | ... |
| **API** | ... |
| **Database** | ... |
| ...    | ...        |
```

Fill in only what actually exists in the project. Do not guess.

#### Project Structure
ASCII tree of main directories (1-2 levels deep) with comments:

```
project/
├── cmd/           # Entry points
├── internal/      # Private packages
├── ...
```

#### Running
Key commands for running, testing, and building. Source from Makefile/Taskfile/package.json.

#### Ports
Port table (if server components exist). Determine from docker-compose.yml, configs, code.

#### Key Interfaces / Entry Points
Main interfaces or application entry points.

#### Adding New Features
Brief guide: how to add a new endpoint / page / module.

### Step 3: Create .spec/agent-rules.md

Create `.spec/agent-rules.md` with rules for AI agents specific to this project.

Determine rules based on:
- Programming language (Go → Effective Go, Python → PEP8, TypeScript → project standards)
- Linter config (`.golangci.yml`, `.eslintrc`, `.prettierrc`, `pyproject.toml`)
- Existing code (naming patterns, error handling, file structure)

Required sections:
- **Code Style** — main style rules
- **Naming Conventions** — variable, function, and file naming
- **Error Handling** — project's error handling pattern
- **Testing** — testing conventions (framework, coverage, patterns)
- **Dependencies** — how to manage dependencies
- **Formatting** — formatting and linting

Each rule: 1-2 lines. No filler.

### Step 4: Output Recommendations

After creating files, output a list of recommended documents to generate:

```
Recommended documents for generation:
1. ARCHITECTURE.md — [reason: detected Clean Architecture / MVC / ... pattern]
2. PACKAGES.md — [reason: X packages in the project]
3. ...
```

Include only documents for which there is real content in the project.

## General Rules

- Language: English
- Use only facts from actual code — do not invent
- If something is unclear — mark as "TODO: requires clarification"
- Format: Markdown, ASCII diagrams, tables
- Style: concise, no filler
- After creating files, the README.md should serve as the map for all future documentation
