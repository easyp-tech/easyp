<!-- generated: 2026-05-14, template: cli.md -->
# EasyP CLI Reference

## 1. Overview

**EasyP** is a modern CLI toolkit for Protocol Buffers development: linting, breaking change detection, code generation, and dependency management.

**Installation:**
```bash
# Homebrew
brew install easyp-tech/tap/easyp

# Go install
go install github.com/easyp-tech/easyp/cmd/easyp@latest

# Docker
docker run --rm -v $(pwd):/work -w /work ghcr.io/easyp-tech/easyp lint
```

**Quick start:**
```bash
easyp init                   # Create easyp.yaml interactively
easyp lint                   # Lint proto files
easyp generate               # Generate code
easyp breaking --against main # Check for breaking changes
```

## 2. Command Tree

```
easyp
├── lint (l)                        # Lint proto files
│   ├── --path/-p <dir>             # Path to proto files (default: ".")
│   └── --root/-r <dir>            # Root directory for file search
├── generate (g)                    # Generate code from proto files
│   ├── --path/-p <dir>             # Path to proto files (default: ".")
│   ├── --root/-r <dir>            # Root directory for file search
│   ├── --descriptor_set_out <path> # Output for binary FileDescriptorSet
│   └── --include_imports           # Include transitive deps in descriptor
├── breaking                        # Breaking change detection
│   ├── --path/-p <dir>             # Path to proto files (default: ".")
│   ├── --against <ref>             # Git ref to compare (default: "master")
│   └── --root/-r <dir>            # Root directory for file search
├── init (i)                        # Initialize easyp.yaml
│   └── --dir/-d <dir>             # Target directory (default: ".")
├── mod (m)                         # Package manager
│   ├── download                    # Download deps to cache
│   ├── update                      # Update deps versions
│   └── vendor                      # Copy deps to vendor dir
├── validate-config (validate)      # Validate easyp.yaml
├── ls-files (ls)                   # List proto files
│   └── --include-imports/-I        # Include transitive imports (default: true)
├── schema-gen                      # Generate JSON Schema artifacts
│   ├── --out-versioned <path>      # Versioned schema path
│   └── --out-latest <path>         # Latest schema alias path
├── completion                      # Shell completion scripts
│   ├── bash                        # Bash completion
│   └── zsh                         # Zsh completion
├── --help/-h                       # Show help
└── --version/-v                    # Show version
```

## 3. Commands Reference

### `easyp lint [flags]`

Lint `.proto` files against configured rules.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `--path` / `-p` | string | no | `.` | Relative path to directory with proto files |
| `--root` / `-r` | string | no | config dir | Root directory for file search |

**Examples:**
```bash
easyp lint                           # Lint current directory
easyp lint --path api/proto          # Lint specific directory
easyp lint --root /projects/myproto  # Custom root
easyp lint --format json             # JSON output (NDJSON)
```

### `easyp generate [flags]`

Generate code from proto files using configured plugins.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `--path` / `-p` | string | no | `.` | Path to proto files |
| `--root` / `-r` | string | no | config dir | Root directory for file search |
| `--descriptor_set_out` | string | no | — | Output path for FileDescriptorSet |
| `--include_imports` | bool | no | `false` | Include transitive deps in descriptor |

**Examples:**
```bash
easyp generate                                     # Generate with defaults
easyp generate --path api/proto                    # Custom proto directory
easyp generate --descriptor_set_out desc.pb        # Emit descriptor set
easyp generate --descriptor_set_out desc.pb --include_imports
```

### `easyp breaking [flags]`

Check for breaking API changes against a Git reference.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `--path` / `-p` | string | no | `.` | Path to proto files |
| `--against` | string | no | `master` | Git ref to compare against |
| `--root` / `-r` | string | no | config dir | Root directory for file search |

**Examples:**
```bash
easyp breaking                       # Compare against master
easyp breaking --against main        # Compare against main
easyp breaking --against v1.0.0      # Compare against tag
easyp breaking --format json         # JSON output
```

### `easyp init [flags]`

Initialize a new project with interactive `easyp.yaml` creation.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `--dir` / `-d` | string | no | `.` | Directory to initialize |

**Examples:**
```bash
easyp init                   # Initialize in current directory
easyp init --dir ./myproject # Initialize in specific directory
```

### `easyp mod <subcommand>`

Package manager for proto dependencies.

| Subcommand | Description |
|------------|-------------|
| `download` | Download modules declared in `deps:` to local cache (`$EASYPPATH`) |
| `update` | Re-resolve and update module versions |
| `vendor` | Copy proto files from deps to `easyp_vendor/` |

**Examples:**
```bash
easyp mod download   # Download all deps
easyp mod update     # Update to latest versions
easyp mod vendor     # Vendor deps locally
```

### `easyp validate-config [flags]`

Validate `easyp.yaml` for syntax, types, and required fields.

Uses only the global `--cfg` flag. Default output format: **JSON**.

**Examples:**
```bash
easyp validate-config                      # Validate default config (JSON)
easyp validate-config --format text        # Human-readable output
easyp validate-config --cfg custom.yaml    # Validate custom config
```

### `easyp ls-files [flags]`

List `.proto` files considering inputs and imports.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `--include-imports` / `-I` | bool | no | `true` | Include transitive imports |

Default output format: **JSON**.

**Examples:**
```bash
easyp ls-files                          # List all files (JSON)
easyp ls-files --format text            # Human-readable
easyp ls-files --include-imports=false  # Skip imports
```

### `easyp schema-gen [flags]`

Generate JSON Schema artifacts for `easyp.yaml` configuration.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `--out-versioned` | string | no | `schemas/easyp-config-v1.schema.json` | Versioned schema path |
| `--out-latest` | string | no | `schemas/easyp-config.schema.json` | Latest schema alias |

### `easyp completion <shell>`

Generate shell completion scripts.

```bash
# Bash
source <(easyp completion bash)

# Zsh
source <(easyp completion zsh)
```

## 4. Configuration

Precedence: **CLI flags → Environment variables → Config file → Defaults**

Config file format: YAML (`easyp.yaml`) with `envsubst` expansion.

| Setting | Flag | Env Var | Config Key | Default |
|---------|------|---------|------------|---------|
| Config file | `--cfg` / `--config` | `EASYP_CFG` | — | `easyp.yaml` |
| Debug mode | `--debug` / `-d` | `EASYP_DEBUG` | — | `false` |
| Output format | `--format` / `-f` | `EASYP_FORMAT` | — | `text` |
| Generate path | `--path` | `EASYP_ROOT_GENERATE_PATH` | — | `.` |
| Init directory | `--dir` | `EASYP_INIT_DIR` | — | `.` |
| Cache directory | — | `EASYPPATH` | — | `~/.easyp` |

## 5. Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success / no issues |
| `1` | Issues found (lint errors, breaking changes, validation errors, version not found) |
| `2` | Infrastructure error (missing import, Git ref not found, repository missing) |

## 6. I/O Contracts

- **stdin:** Not used by any command (except `init` which uses interactive TUI prompts).
- **stdout:** Command output. Format depends on `--format` flag:
  - `lint`, `breaking` → NDJSON (`text` default, `json` per-line objects)
  - `validate-config`, `ls-files` → single JSON document (`json` default)
  - `completion` → shell script
  - `schema-gen` → writes to files, not stdout
- **stderr:** Debug logs (when `--debug` is enabled), error messages.
- **Files created/modified:**
  - `init` → creates `easyp.yaml`
  - `mod download` → caches repos in `$EASYPPATH` (`~/.easyp`)
  - `mod update` → updates `easyp.lock`
  - `mod vendor` → copies protos to `easyp_vendor/`
  - `generate` → writes generated code to plugin `out` directories
  - `schema-gen` → writes JSON Schema files

## 7. Shell Completion

Supported shells: **Bash**, **Zsh**.

```bash
# Bash — add to ~/.bashrc
source <(easyp completion bash)

# Zsh — add to ~/.zshrc
source <(easyp completion zsh)
```

Completions are generated dynamically via the `completion` subcommand.

## 8. Global Flags

| Flag | Type | Default | Env Var | Description |
|------|------|---------|---------|-------------|
| `--cfg` / `--config` | string | `easyp.yaml` | `EASYP_CFG` | Path to config file |
| `--debug` / `-d` | bool | `false` | `EASYP_DEBUG` | Enable debug logging to stderr |
| `--format` / `-f` | enum | `text` | `EASYP_FORMAT` | Output format: `text` or `json` |
| `--help` / `-h` | — | — | — | Show help |
| `--version` / `-v` | — | — | — | Show version |

## 9. Error Messages & Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `Cannot import file: <name>` | Missing proto import | Add the dependency to `deps:` in `easyp.yaml` and run `easyp mod download` |
| `Cannot find git ref: <ref>` | Invalid `--against` reference | Verify the branch/tag exists: `git branch -a` or `git tag -l` |
| `Repository does not exist` | `breaking` run outside a git repo | Run from within a git repository |
| `config not found` | Missing `easyp.yaml` | Run `easyp init` or specify `--cfg <path>` |
| `version not found` | Dependency version doesn't exist | Check the version tag in the dependency repo |

## 10. Development

**Run locally:**
```bash
task build && ./easyp lint
# or
go run ./cmd/easyp lint
```

**Build release binary:**
```bash
task build
# Output: ./easyp
```

**Add a new command:**
1. Create `internal/api/<command>.go` — struct implementing `Handler`
2. Add `var _ Handler = (*MyCommand)(nil)` for interface compliance
3. Implement `Command() *cli.Command` with flags, description, action
4. Register in `cmd/easyp/main.go` via `buildCommand(...)` call

**Source files:**
- Entry point: [`cmd/easyp/main.go`](../cmd/easyp/main.go)
- Handlers: [`internal/api/`](../internal/api/)
- Global flags: [`internal/flags/flags.go`](../internal/flags/flags.go)
- CLI framework: `github.com/urfave/cli/v2`
