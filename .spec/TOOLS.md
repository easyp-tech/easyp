<!-- generated: 2026-05-14, template: development.md -->
# EasyP Tools

## 0. Dev Environment Setup

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.24+ | `brew install go` |
| Task | v3 | `brew install go-task` |
| golangci-lint | v2.1.6 | Installed via `task init` |
| gotestsum | latest | Installed via `task init` |
| mockery | v2.41.0 | Installed via `task init` |
| hadolint | latest | Installed via `task init` |

### First Run

```bash
git clone https://github.com/easyp-tech/easyp.git
cd easyp
task init       # Install all dev tools
task build      # Build binary
task test       # Verify everything works
```

## 1. Overview

All commands are managed via [Task](https://taskfile.dev) v3 (`Taskfile.yml`). Run commands with `task <target>`.

## 2. Quick Reference

| Action | Command |
|--------|---------|
| Build | `task build` |
| Test | `task test` |
| Lint | `task lint` |
| Full quality check | `task quality` |
| Coverage report | `task coverage` |
| Regenerate mocks | `task mocks` |
| Regenerate schemas | `task schema:generate` |
| Check schema freshness | `task schema:check` |
| Install binary | `task install` |
| Init dev tools | `task init` |

## 3. Detailed Command Groups

### Testing

```bash
task test           # Run all tests with gotestsum (-race, -count=1)
task coverage       # Generate and open coverage report in browser
```

- Uses `gotestsum` with `--format pkgname`
- Always runs with `-race -count=1`
- Outputs to stdout in human-readable format

### Building

```bash
task build          # Build easyp binary (output: ./easyp)
task install        # Install to $GOPATH/bin
```

### Linting

```bash
task lint           # Run golangci-lint + hadolint
```

- Go linter: `golangci-lint` v2.1.6 with staticcheck
- Docker linter: `hadolint` for Dockerfile

### Code Generation

```bash
task mocks              # Regenerate all mockery mocks
task schema:generate    # Regenerate JSON Schema from Go types
task schema:check       # Ensure schemas are up to date (used in CI)
```

**Mocks:**
- Tool: `mockery` v2.41.0
- Config: `with-expecter: true`, `inpackage: false`, `disable-version-string: true`
- Output: `mocks/` subdirectory next to the interface

**Schemas:**
- Source-of-truth: `mcp/easypconfig` package
- Output: `schemas/easyp-config-v1.schema.json` + `schemas/easyp-config.schema.json`

## 4. CI/CD Cheatsheet

Simulate the full CI pipeline locally:

```bash
task init       # Install tools (same as CI)
task quality    # Runs both test and lint (same as CI)
task schema:check  # Ensure schemas are fresh
```

## 5. Tool Installation

| Tool | Install Command |
|------|----------------|
| Task | `brew install go-task` or `go install github.com/go-task/task/v3/cmd/task@latest` |
| golangci-lint | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6` |
| gotestsum | `go install gotest.tools/gotestsum@latest` |
| mockery | `go install github.com/vektra/mockery/v2@v2.41.0` |
| hadolint | `brew install hadolint` |
