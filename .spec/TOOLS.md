<!-- generated: 2026-07-27, template: development.md -->
# Development Tools

## Dev Environment Setup

EasyP is a Go Protocol Buffers CLI toolkit. The module declares Go 1.24.0,
and the test workflow in GitHub Actions installs Go 1.24 and Task 3.x.

| Tool | Version / source | Purpose |
|---|---|---|
| Go | 1.24.0 (`go.mod`) | Build and test the CLI |
| Task | 3.x in CI | Run the repository task targets |
| Docker | Required by lint and image tasks | Run Hadolint and build images |
| Git | Required by EasyP module operations and releases | Work with Git-backed proto dependencies |

`Taskfile.yml` installs project-local helpers into `bin/`:
`golangci-lint` 2.1.6, `gotestsum` 1.11.0, and `mockery` 2.41.0.
Hadolint v2.12.1-beta is pulled as a Docker image instead of installed
locally.

### First Run

1. Clone the repository and enter its root.
2. Install Go 1.24 and Task.
3. Ensure Docker is available if you will run `task lint`, Docker image
   tasks, or release snapshots.
4. Run `task init`. It installs the local Go tools into `bin/` and runs
   `go get -v ./...`.
5. Run `task test` to verify the checkout.
6. Run `task build` to create the root-level `easyp` executable.

The repository has no `docker-compose` file or service startup target. EasyP
is a CLI, so development verification is normally a build or test command.

## Overview

Prefer `task` targets from `Taskfile.yml`; they encode the local tool paths,
coverage settings, race detection, and Docker-based lint checks used by the
project.

## Quick Reference

| Action | Command |
|---|---|
| Install development helpers | `task init` |
| Build the CLI | `task build` |
| Run the full test suite | `task test` |
| Run lint checks | `task lint` |
| Run tests and lint | `task quality` |
| Open HTML coverage | `task coverage` |
| Install the CLI into `GOBIN` | `task install` |
| Remove local build/tool artifacts | `task clean` |
| Regenerate config schemas | `task schema:generate` |
| Check config schema drift | `task schema:check` |
| Regenerate mocks | `task mocks` |

## Build and Install

```bash
task build
task install
```

`task build` runs `go build -o easyp ./cmd/easyp`, placing a local executable
at `./easyp`. `task install` instead runs `go install ./cmd/easyp` and uses
the Go installation destination. Neither task needs generated schemas or
mocks as a prerequisite.

The direct build equivalent is:

```bash
go build -o easyp ./cmd/easyp
```

The root `Dockerfile` is a separate production build path. It compiles with
`CGO_ENABLED=0` using `golang:1.25-alpine` and packages the binary in
`alpine:3.22` with CA certificates, timezone data, Git, and Bash.

## Testing and Coverage

```bash
task test
task coverage
```

`task test` invokes the project-local `bin/gotestsum` with:

```bash
bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
```

This is the repository-wide test command: it records coverage in
`coverage.out`, enables the race detector, and disables Go test caching.
`task coverage` opens that existing profile with `go tool cover -html`; run
`task test` first. See [TESTING.md](TESTING.md) for conventions and narrower
Go commands.

## Linting and Quality

```bash
task lint
task quality
```

`task lint` first depends on `task build`. It runs Hadolint in Docker against
`Docker/base/Dockerfile` and `Docker/lint/Dockerfile`, then runs
`bin/golangci-lint run ./...`. The checked-in `.golangci.yml` enables
`staticcheck`; the installed linter version is pinned by the Taskfile.

`task quality` depends on both `task test` and `task lint`. It is the closest
local approximation of the test and lint checks. CI's `tests.yml` workflow
currently performs `task init` followed by `task test`; it does not invoke
the lint target in that workflow.

## Generated Artifacts

### Configuration JSON Schema

```bash
task schema:generate
task schema:check
```

`task schema:generate` runs `go run ./cmd/easyp schema-gen`, producing the
versioned and latest schema artifacts under `schemas/`. The `go:generate`
directive in `mcp/easypconfig/generate.go` can also run schema generation with
explicit output paths.

Use `task schema:check` after changing config schema metadata or validation.
It regenerates the artifacts and fails if
`schemas/easyp-config-v1.schema.json` or `schemas/easyp-config.schema.json`
would differ. Do not hand-edit those generated JSON files.

### Test Mocks

```bash
task mocks
go generate ./internal/core ./mcp/easypconfig
```

`task mocks` uses local Mockery to regenerate mocks for `LockFile` in storage
and several `internal/core` interfaces, including `Rule`, `Console`,
`CurrentProjectGitWalker`, `Storage`, and `ModuleConfig`. It writes mocks to
the package `mocks/` directories.

`internal/core/mod.go` also declares test-only `go:generate mockery`
directives that generate `storage_mock_test.go` and `lockfile_mock_test.go`
in the `core` package. Regenerate mocks after changing an interface they
implement.

## Docker and Release Checks

```bash
task docker_base GIT_TAG=v0.0.0
task docker_lint GIT_TAG=v0.0.0
task goreleaser:check
task goreleaser:test
task goreleaser:test-docker
```

`docker_base` and `docker_lint` build and tag the base and lint images.
`task docker` obtains the current Git tag, builds both images, and pushes
them, so it is a publishing command rather than a local smoke test.

`task goreleaser:check` validates GoReleaser configuration. The snapshot
tasks create or select a `multiarch` Docker Buildx builder and run
`goreleaser release --snapshot --clean`; they do not publish a release.
The release workflow runs for stable `vMAJOR.MINOR.PATCH` tags and authenticates
to GitHub Container Registry and Docker Hub.

## Dependency and Vendor Tools

EasyP module dependencies are Git repositories, cached under `EASYPPATH`
(default `$HOME/.easyp`) and locked in `protobuf.lock`. The vendoring command is:

```bash
./easyp mod vendor
```

It copies installed proto dependencies to `easyp_vendor`, not Go's usual
`vendor/` directory. This is EasyP dependency cache and package-management
behavior, not application file storage. See
[config/dependency.md](config/dependency.md) for the authoritative dependency
flow and available `easyp mod` subcommands.

## Tool Installation

| Tool | Installation used by this repository |
|---|---|
| Go | Install Go 1.24, then verify with `go version` |
| Task | Install Task 3.x; CI uses `arduino/setup-task@v1` |
| GolangCI-Lint | `task init` installs it into `./bin` |
| Gotestsum | `task init` installs it into `./bin` |
| Mockery | `task init` installs it into `./bin` |
| Hadolint | `task lint` pulls `ghcr.io/hadolint/hadolint:v2.12.1-beta` |
| GoReleaser | Provide `goreleaser` on `PATH` for its Task targets |

`task clean` removes `bin/` and `coverage.out`, then clears the Go build
cache. It does not remove module caches, EasyP's dependency cache, schemas,
or the root-level `easyp` binary.
