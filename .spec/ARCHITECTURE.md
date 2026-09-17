<!-- generated: 2026-05-14, template: core.md -->
# EasyP Architecture

## 1. Overview

EasyP is a CLI application following a **clean/hexagonal architecture** pattern with three well-defined layers.

```
┌──────────────────────────────────────────────────────┐
│                    cmd/easyp/                        │
│               CLI entrypoint (main.go)               │
│          Registers handlers, initializes logger      │
└────────────────────────┬─────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────┐
│                   internal/api/                      │
│             CLI Command Handlers                     │
│    Parses flags, reads config, builds Core,          │
│    formats output                                    │
└────────────────────────┬─────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────┐
│                  internal/core/                      │
│              Business Logic (Core)                   │
│   Lint, Generate, Download, BreakingCheck,           │
│   Initialize, ListFiles, Vendor, Update              │
│                                                      │
│   Depends on INTERFACES (ports), not implementations │
└────────┬───────────────┬──────────────┬──────────────┘
         │               │              │
         ▼               ▼              ▼
┌────────────┐  ┌──────────────┐  ┌────────────────┐
│  adapters/ │  │   adapters/  │  │   adapters/    │
│  storage   │  │   go_git     │  │   plugin/      │
│  lock_file │  │   repository │  │  (4 executors) │
│  console   │  │              │  │                │
│  prompter  │  │              │  │                │
└────────────┘  └──────────────┘  └────────────────┘
```

## 2. Component Deep Dive

### cmd/easyp (Entrypoint)

| File | Description |
|------|-------------|
| `main.go` | Creates `cli.App`, registers all handlers, initializes `slog` logger |

Registers handlers via `buildCommand(...)`: Lint, Mod, Completion, Init, Generate, SchemaGen, LsFiles, Validate, BreakingCheck.

### internal/api (Command Handlers)

Each handler implements `Handler` interface: `Command() *cli.Command`.

| File | Description |
|------|-------------|
| `lint.go` | Lint handler — walks protos, applies rules, prints issues |
| `generate.go` | Generate handler — invokes plugin executors |
| `breaking_check.go` | Breaking change handler — compares against git ref |
| `init.go` | Init handler — interactive config creation |
| `mod.go` | Mod handler — download/update/vendor subcommands |
| `validate.go` | Validate handler — config validation |
| `ls_files.go` | LsFiles handler — lists proto files |
| `schema_gen.go` | SchemaGen handler — generates JSON Schema |
| `completion.go` | Completion handler — bash/zsh scripts |
| `temporaly_helper.go` | Shared helpers: `buildCore()`, `getLogger()`, `getEasypPath()` |

### internal/core (Business Logic)

Central orchestrator. The `Core` struct holds all dependencies and exposes methods for each operation.

| File | Description |
|------|-------------|
| `core.go` | `Core` struct, constructor `New(...)`, dependency injection |
| `dom.go` | Domain types: `Rule`, `Issue`, `ProtoInfo`, `Plugin`, `Repo` |
| `lint.go` | `Core.Lint()` — walks files, applies rules |
| `generate.go` | `Core.Generate()` — compiles protos, invokes plugins |
| `breaking.go` | `Core.BreakingCheck()` — compares proto versions |
| `download.go` | `Core.Download()` — fetches dependencies |
| `initialize.go` | `Core.Initialize()` — creates easyp.yaml |
| `ls_files.go` | `Core.ListFiles()` — lists proto files |
| `vendor.go` | `Core.Vendor()` — copies deps to vendor dir |
| `nolint.go` | `nolint:` / `buf:lint:ignore` comment parsing |

### internal/adapters (Infrastructure)

| Adapter | Package | Purpose |
|---------|---------|---------|
| Storage | `adapters/storage` | Dependency cache management (`~/.easyp`) |
| Git Walker | `adapters/go_git` | Git-based dir walker for breaking checks |
| Lock File | `adapters/lock_file` | `easyp.lock` read/write |
| Module Config | `adapters/module_config` | Module configuration reader |
| Console | `adapters/console` | stdin/stdout interaction |
| Prompter | `adapters/prompter` | Interactive TUI prompts (bubbletea) |
| Repository | `adapters/repository` | Git repository operations |
| Plugin | `adapters/plugin/` | 4 executor types: local, remote, builtin (WASM), command |

## 3. Key Design Decisions

1. **Clean Architecture** — Core depends only on interfaces (ports), never on concrete implementations. This enables easy testing via mocks and swapping adapters.

2. **Interface Compliance Guards** — Every concrete type verifies its interface at compile time:
   ```go
   var _ Handler = (*Lint)(nil)
   var _ core.Rule = (*FieldLowerSnakeCase)(nil)
   ```

3. **Plugin Executor Strategy** — Code generation supports 4 execution strategies (local, remote, WASM, command) behind a single `Executor` interface. The executor is chosen based on the `PluginSource` fields.

4. **Rule Auto-Discovery** — Lint rule names are auto-derived from struct names via `PascalCase → UPPER_SNAKE_CASE` reflection. Groups expand to individual rules at build time.

5. **Config-Driven** — All behavior is driven by `easyp.yaml`. No hardcoded defaults in business logic — everything comes from config or constructor injection.

## 4. Data Flow

```
CLI invocation (easyp lint --path api/proto)
  → cmd/easyp/main.go: cli.App.Run()
    → internal/api/lint.go: Lint.Action()
      → resolveRoots(): compute configPath, projectRoot, lintRoot
      → config.New(): parse easyp.yaml
      → buildCore(): construct Core with all adapters
        → rules.New(): build lint rules from config
        → storage.New(): init cache storage
        → go_git.New(): init git walker
      → core.Lint(ctx, dirWalker)
        → dirWalker.Walk(): iterate *.proto files
        → parse each file (go-protoparser)
        → for each rule: rule.Validate(protoInfo)
          → AppendIssue(): check nolint comments, collect issues
      ← return []IssueInfo
    → printIssues(format, stdout, issues)
      → textPrinter() or jsonPrinter()
    ← os.Exit(0 or 1)
```
