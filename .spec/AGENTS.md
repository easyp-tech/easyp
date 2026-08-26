<!-- generated: 2026-05-14, template: agents-index.md -->
# AGENTS.md — AI Agent Guide for `.spec/` Documentation

## 1. What is `.spec/`?

`.spec/` is the project documentation directory optimized for AI agent (LLM) consumption. It contains structured, machine-friendly descriptions of the **EasyP** codebase:

- **Architecture & design** — layered architecture, component relationships, data flows
- **Package descriptions** — purpose and responsibilities of each Go package
- **Domain model** — core types, interfaces, and business rules for Protocol Buffers tooling
- **Code & testing conventions** — naming, error handling, test patterns
- **Infrastructure & tooling** — CI/CD, Docker, release process
- **Agent rules** (`agent-rules.md`) — mandatory constraints for AI-assisted development

**Purpose:** give the agent full project context without reading the entire source code. Start here instead of `grep`-ing through hundreds of files.

## 2. How to Use (for agents)

When working on the EasyP project, follow this reading order:

1. **Starting a task** → read `.spec/README.md` for the documentation map and quick facts.
2. **Before modifying code** → read the relevant document:
   - Changing API handlers → `ARCHITECTURE.md` + `CODE_STYLE.md`
   - Adding a lint rule → `DOMAIN.md` + `TESTING.md`
   - Modifying config parsing → `ARCHITECTURE.md` + `PACKAGES.md`
   - Working with plugins → `ARCHITECTURE.md` (Plugin Executors section)
   - Updating CI/CD → `TOOLS.md` + `DEPLOYMENT.md`
3. **Always follow** rules from `agent-rules.md` — these are project-specific constraints.
4. **If a document appears outdated** — suggest an update to the user before proceeding.

## 3. File Structure Convention

Files in `.spec/` follow these naming conventions:

| Pattern | Purpose | Examples |
|---------|---------|---------|
| `README.md` | Index of all documents, quick facts, project structure, ports, commands | — |
| `agent-rules.md` | Mandatory rules for AI agents (code style, naming, error handling) | — |
| `UPPER_CASE.md` | Topical documents | `ARCHITECTURE.md`, `TESTING.md`, `DOMAIN.md` |
| `features/` | Pipeline phase artifacts (per-feature) | `features/new-rule/explore.md` |

> **Directory separation:** Project documentation lives in `.spec/`. Feature pipeline artifacts live in `.spec/features/<feature>/`. Never mix them.

## 4. Document Categories

### Core
- `ARCHITECTURE.md` — Layered architecture (cmd → api → core ← adapters), component diagram
- `PACKAGES.md` — Go package inventory with responsibilities
- `DOMAIN.md` — Core types (`ProtoInfo`, `Rule`, `Issue`, `Plugin`), interfaces, value objects
- `CODE_STYLE.md` — Naming conventions, error wrapping, interface compliance patterns

### Development
- `TOOLS.md` — Task runner commands, linters, mockery, GoReleaser
- `TESTING.md` — testify patterns, testdata fixtures, mock generation, race detection
- `FILES.md` — Key file locations and their purposes

### Infrastructure
- `DEPLOYMENT.md` — Docker multi-stage build, GoReleaser, GitHub Actions, Homebrew tap
- `CLI.md` — Full CLI command reference with flags, exit codes, output formats

## 5. How to Maintain

Keep documentation current by following these rules:

- **Adding a new lint rule** → update `DOMAIN.md` (rule list) and `TESTING.md` (test pattern)
- **Adding a new CLI command** → update `CLI.md` and `ARCHITECTURE.md` (handler list)
- **Changing architecture** → update `ARCHITECTURE.md`
- **Adding a dependency** → update `PACKAGES.md`
- **Changing config schema** → update `DOMAIN.md` and run `task schema:generate`
- `README.md` must always reflect the current list of documents
- Documents must contain code examples from the actual project, not abstract ones

## 6. How to Add a New Document

1. Create a file `TOPIC_NAME.md` in `.spec/`
2. Add `<!-- generated: YYYY-MM-DD, template: <template>.md -->` as line 1
3. Use this structure:
   - Title → Overview → Architectural diagram (ASCII/Mermaid) → Details → Code examples → Configuration → Testing → Key files
4. Add a link in `.spec/README.md` under the appropriate category
5. Use real code examples from the EasyP source, not invented ones
