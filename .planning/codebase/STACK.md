---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
# Technology Stack

**Analysis Date:** 2026-11-17

## Languages

**Primary:**
- Go 1.26.6 (`go.mod` declares `go 1.26.6`) - entire CLI toolkit, all `cmd/` and `internal/` packages

**Secondary:**
- Protocol Buffers - embedded proto files and generated plugin descriptors (e.g. `easyp/generator/v1` API types from `github.com/easyp-tech/service/api`)
- YAML - user-facing config: `easyp.yaml` (lint/breaking policy), `easyp.gen.yaml` (generation), `protobuf.mod` / `protobuf.lock`
- JSON Schema - generated config schemas in `schemas/` (do not hand-edit)
- Dockerfile - root `Dockerfile` for the CLI image
- Shell - `Taskfile.yml` tasks and `scripts/` helpers (e.g. `scripts/check-proto-example.sh`, `scripts/validate-release-tag.sh`)

## Runtime

**Environment:**
- Go toolchain 1.26.6 (CI uses `go-version-file: go.mod`)
- Docker required for `task lint:docker` (Hadolint) and `task docker:build`
- Git required for module operations (EasyP's dependency model is Git-based)

**Package Manager:**
- Go modules (`go.mod` / `go.sum`); no lockfile tooling beyond go.sum

## Frameworks

**Core:**
- CLI framework: `github.com/urfave/cli/v2` v2.27.7 - command wiring in `cmd/easyp` / `internal/api`
- gRPC: `google.golang.org/grpc` v1.83.2 - remote plugin execution (`internal/adapters/plugin/remote.go`)
- Protobuf runtime: `google.golang.org/protobuf` v1.36.11
- Proto compilation: `github.com/bufbuild/protocompile` v0.14.1 (`internal/core`)
- MCP server SDK: `github.com/modelcontextprotocol/go-sdk` v1.6.1 (`cmd/easyp-mcp`, `mcp/easypconfig`)

**Testing:**
- `github.com/stretchr/testify` v1.11.1 - all `*_test.go` files
- `gotestsum` v1.13.0 (pinned in `bin/`) - test runner with `-race -count=1` and coverage
- `mockery` v2.53.7 (pinned, optional) - mocks for `core.Rule`, `core.CurrentProjectGitWalker`, `console.Console`

**Build/Dev:**
- Task (Taskfile.yml, v3 in CI) - all dev workflows; prefer `task` targets over raw commands
- `golangci-lint` v2.14.0 (pinned in `bin/`, config in `.golangci.yml`)
- Hadolint v2.12.1-beta (Docker image, not local install)
- Docker Buildx + QEMU for multi-arch release builds (`.github/workflows/release.yml`)

## Key Dependencies

**Critical:**
- `github.com/bufbuild/protocompile` v0.14.1 - proto parsing/compilation, the heart of `internal/core`
- `github.com/yoheimuta/go-protoparser/v4` v4.14.2 - proto source parsing for lint rules and comment suppressions
- `github.com/go-git/go-git/v5` v5.16.3 - in-process Git operations for module cache/snapshots (`internal/adapters/go_git`, `internal/adapters/gitmodules`)
- `github.com/tetratelabs/wazero` v1.9.0 (+ `github.com/wasilibs/wazero-helpers`) - WASM runtime for builtin plugins (`internal/adapters/plugin/builtin_impl.go`, `builtin_run.go`)
- `github.com/easyp-tech/service/api` v1.0.0 - remote generator gRPC API (`easyp/generator/v1`)
- `github.com/urfave/cli/v2` v2.27.7 - CLI framework

**Infrastructure:**
- `github.com/a8m/envsubst` v1.4.3 - env substitution in config files
- `github.com/Yakwilik/go-yamlvalidator` v0.3.1 + `gopkg.in/yaml.v3` - YAML config parsing/validation
- `github.com/samber/lo` v1.52.0 - generic utilities
- `github.com/charmbracelet/x/term` v0.2.1 - terminal interaction (interactive migrate wizard, prompts)
- `golang.org/x/mod` v0.38.0 - semver handling
- `github.com/santhosh-tekuri/jsonschema/v6` (indirect, via schema tooling) and `github.com/google/jsonschema-go` v0.4.3 - JSON Schema generation in `internal/schemagen`
- Embedded plugin protos (indirect): `github.com/easyp-tech/protoc-gen-easydoc` v0.4.0, `github.com/easyp-tech/protoc-gen-mcp` v0.5.0

## Configuration

**Environment:**
- `EASYP_PROXY` - proxy host for Git operations (validated host-only, see `internal/config/proxy.go`)
- `EASYP_GIT_TIMEOUT` - per-operation Git timeout duration (see `internal/adapters/gitcommand/command.go`)
- `GIT_TERMINAL_PROMPT=0` is forced by the tool itself for non-interactive Git
- No `.env` file mechanism; configuration comes from CLI flags, `easyp.yaml`, `easyp.gen.yaml`, `protobuf.mod` and `protobuf.lock`

**Build:**
- `Taskfile.yml` - all task definitions, pinned tool versions
- `.golangci.yml` - golangci-lint v2 config (staticcheck enabled; ST1005 exemption for identifier-style error wrappers per `.spec/agent-rules.md`)
- `Dockerfile` - root CLI image
- `schemas/*.json` - generated via `go run ./cmd/easyp schema-gen --out-dir schemas`; never hand-edit

## Platform Requirements

**Development:**
- Go 1.26.6, Task 3.x, Docker (for lint/docker tasks), Git
- `task init` installs pinned helpers into repo-local `bin/`

**Production:**
- Distributed as a Go binary (`easyp`) plus Docker image; released via GitHub Actions on `v1.0.0-nightly.*` tags (multi-arch), separate MCP binary `cmd/easyp-mcp`
- Docs site lives in the separate `easyp-tech/docs` repository

---

*Stack analysis: 2026-11-17*
