---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
# External Integrations

**Analysis Date:** 2026-11-17

## APIs & External Services

**Remote code-generation plugins (gRPC):**
- EasyP remote generator service - executes `protoc` plugins server-side
  - SDK/Client: `google.golang.org/grpc` + generated stubs from `github.com/easyp-tech/service/api` (`easyp/generator/v1`)
  - Implementation: `internal/adapters/plugin/remote.go` — parses plugin URL into host/plugin/version, 30s timeout, TLS via system cert pool for `https://` or non-localhost hosts, insecure creds for `http://`/localhost/127.0.0.1; dials through the proxy configured by `EASYP_PROXY`
  - Requests/responses use `google.golang.org/protobuf/types/pluginpb` (`CodeGeneratorRequest`/`Response`), encoded deterministically

**Git hosting providers (dependency management):**
- Arbitrary Git remotes (GitHub, GitLab, etc.) — module dependency fetch, verify, vendor
  - Implementation: two adapters:
    - `internal/adapters/go_git/` — in-process `go-git` for revisions and snapshots (`revision.go`, `snapshot.go`)
    - `internal/adapters/gitcommand/` — `git` subprocess wrapper with `GIT_TERMINAL_PROMPT=0` and `EASYP_GIT_TIMEOUT` per-operation limit (`command.go`, `cancel_unix.go`)
  - Verified checkouts and the v1 module cache: `internal/adapters/gitmodules/`
  - Module/lock logic: `internal/modules/` (manifest/lock, import roots, vendoring)

**Buf Schema Registry (compatibility mapping):**
- Known `buf.build/*` modules are mapped to explicit Git snapshots for migration
  - Implementation: `internal/adapters/bsr/resolver.go` — hard-coded table, e.g. `buf.build/googleapis/googleapis` → `github.com/googleapis/googleapis` at pinned commit; used by `easyp migrate` flows in `internal/migration`

**Local/WASM plugins:**
- Built-in plugins run in-process via the wazero WASM runtime (embedded wasm in `internal/adapters/plugin/`)
  - Implementation: `internal/adapters/plugin/builtin.go`, `builtin_impl.go`, `builtin_run.go` (uses `tetratelabs/wazero` + `wasilibs/wazero-helpers`)
- Local protoc-style plugins run as subprocesses: `internal/adapters/plugin/local.go`, `command.go`

**MCP (Model Context Protocol):**
- `easypconfig` MCP server exposing policy/generation YAML schemas, `protobuf.lock` JSON Schema and `protobuf.mod` grammar to AI agents
  - Implementation: `cmd/easyp-mcp/main.go` (stdio transport via `modelcontextprotocol/go-sdk`), tool registration in `mcp/easypconfig/tool.go` / `RegisterTool`; schema source is `v1.SchemaJSON` from `internal/config/v1`

## Data Storage

**Databases:**
- None — no database usage

**File Storage:**
- Local filesystem only: project files, `protobuf.mod`, `protobuf.lock`, vendored modules under the v1 Git module cache (`internal/adapters/gitmodules`), generated code output, `schemas/` artifacts

**Caching:**
- Git module cache managed by `internal/adapters/gitmodules` / `internal/modules` (verified checkouts keyed by revision)

## Authentication & Identity

**Auth Provider:**
- None custom. Git access relies on ambient credentials (SSH agent/config, `https` helpers); the tool forces `GIT_TERMINAL_PROMPT=0` so no interactive credential prompts occur
- Remote plugins: TLS only, no client authentication detected

## Monitoring & Observability

**Error Tracking:**
- None (no Sentry/OTel)

**Logs:**
- Standard library `log/slog`, wrapped by `internal/logger/` (`logger.go` interface, `slog.go` adapter, `nop.go` no-op); structured `slog.Attr` logging threaded through context

## CI/CD & Deployment

**Hosting:**
- GitHub repository `github.com/easyp-tech/easyp`; distributed as Go binary and Docker image (multi-arch via QEMU/Buildx in `.github/workflows/release.yml`)
- Documentation site in separate repo `easyp-tech/docs`

**CI Pipeline:**
- GitHub Actions: `.github/workflows/tests.yml` (tests), `release.yml` (nightly `v1.0.0-nightly.*` tag releases, validates tag via `scripts/validate-release-tag.sh`, publishes packages), `relator.yml`
- Docker image publishing only from release workflow; `task docker:build` is local-only

## Environment Configuration

**Required env vars:**
- None strictly required. Optional: `EASYP_PROXY` (Git proxy host, `internal/config/proxy.go`), `EASYP_GIT_TIMEOUT` (Git op timeout, `internal/adapters/gitcommand/command.go`)

**Secrets location:**
- None in repo; Git credentials come from the user environment

## Webhooks & Callbacks

**Incoming:**
- None (CLI tool; the MCP server speaks stdio, not HTTP)

**Outgoing:**
- None

---

*Integration audit: 2026-11-17*
