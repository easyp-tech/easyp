<!-- generated: 2026-09-30, template: development.md -->
# Development Tools

## Dev Environment Setup

EasyP is a Go Protocol Buffers CLI toolkit. The module declares Go 1.26.6;
the test and release workflows use <code>go-version-file: go.mod</code>. CI installs Task 3.x.

| Tool | Version / source | Purpose |
|---|---|---|
| Go | 1.26.6 (<code>go.mod</code>) | Build and test the CLI |
| Task | 3.x in CI | Run the repository task targets |
| Docker | Required by lint and image tasks | Run Hadolint and build images |
| Git | Required by EasyP module operations and releases | Work with Git-backed proto dependencies |

<code>Taskfile.yml</code> installs project-local helpers into <code>bin/</code>:
<code>golangci-lint</code> 2.1.6, <code>gotestsum</code> 1.13.0, and <code>mockery</code> 2.53.7.
Hadolint v2.12.1-beta is pulled as a Docker image instead of installed
locally.

### First Run

1. Clone the repository and enter its root.
2. Install Go 1.26.6 and Task.
3. Ensure Docker is available if you will run <code>task lint</code>, Docker image
   tasks, or release snapshots.
4. Run <code>task init</code>. It installs the local Go tools into <code>bin/</code> and runs
   <code>go get -v ./...</code>.
5. Run <code>task test</code> to verify the checkout.
6. Run <code>task build</code> to create the root-level <code>easyp</code> executable.

The repository has no <code>docker-compose</code> file or service startup target. EasyP
is a CLI, so development verification is normally a build or test command.

## Overview

Prefer <code>task</code> targets from <code>Taskfile.yml</code>; they encode the local tool paths,
coverage settings, race detection, and Docker-based lint checks used by the
project.

## Quick Reference

| Action | Command |
|---|---|
| Install development helpers | <code>task init</code> |
| Build the CLI | <code>task build</code> |
| Run the full test suite | <code>task test</code> |
| Run lint checks | <code>task lint</code> |
| Run tests and lint | <code>task quality</code> |
| Open HTML coverage | <code>task coverage</code> |
| Install the CLI into <code>GOBIN</code> | <code>task install</code> |
| Remove local build/tool artifacts | <code>task clean</code> |
| Regenerate config schemas | <code>task schema:generate</code> |
| Check config schema drift | <code>task schema:check</code> |
| Regenerate configured mocks (currently partially stale) | <code>task mocks</code> |
| Validate/generate the native v1 example | <code>task proto:check</code> |

## Build and Install

~~~bash
task build
task install
~~~

<code>task build</code> runs <code>go build -o easyp ./cmd/easyp</code>, placing a local executable
at <code>./easyp</code>. <code>task install</code> instead runs <code>go install ./cmd/easyp</code> and uses
the Go installation destination. Neither task needs generated schemas or
mocks as a prerequisite.

The direct build equivalent is:

~~~bash
go build -o easyp ./cmd/easyp
~~~

The root <code>Dockerfile</code> is a separate production build path. It compiles with
<code>CGO_ENABLED=0</code> using <code>golang:1.26-alpine</code> and packages the binary in
<code>alpine:3.22</code> with CA certificates, timezone data, Git, and Bash.

## Testing and Coverage

~~~bash
task test
task coverage
~~~

<code>task test</code> invokes the project-local <code>bin/gotestsum</code> with:

~~~bash
bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
~~~

This is the repository-wide test command: it records coverage in
<code>coverage.out</code>, enables the race detector, and disables Go test caching.
<code>task coverage</code> opens that existing profile with <code>go tool cover -html</code>; run
<code>task test</code> first. See [TESTING.md](TESTING.md) for conventions and narrower
Go commands.

## Linting and Quality

~~~bash
task lint
task quality
~~~

<code>task lint</code> first depends on <code>task build</code>. Its Hadolint steps still reference
missing <code>Docker/base/Dockerfile</code> and <code>Docker/lint/Dockerfile</code>; this unresolved
Taskfile issue prevents the target reaching <code>bin/golangci-lint run ./...</code>.
The existing production Docker build is root <code>Dockerfile</code>. The checked-in
<code>.golangci.yml</code> enables <code>staticcheck</code>; the installed linter version is pinned
by the Taskfile. When the local tool is available, its Go-only invocation is
<code>bin/golangci-lint run ./...</code>.

<code>task quality</code> depends on both <code>task test</code> and <code>task lint</code>. It is the closest
local approximation of the test and lint checks. CI's <code>tests.yml</code> workflow
performs <code>task init</code>, <code>task test</code> and <code>task proto:check</code>; it does not invoke
the lint target in that workflow.

## Generated Artifacts

### Configuration JSON Schema

~~~bash
task schema:generate
task schema:check
~~~

<code>task schema:generate</code> runs <code>go run ./cmd/easyp schema-gen --out-dir schemas</code>.
The generator derives separate versioned and latest JSON Schemas from the v1
models for <code>easyp.yaml</code>, <code>easyp.gen.yaml</code>, and <code>protobuf.lock</code>. The text
<code>protobuf.mod</code> manifest is validated by <code>easyp validate-config</code>.

Use <code>task schema:check</code> after changing v1 configuration models or validation.
It generates all six artifacts in a temporary directory and diffs that directory
against <code>schemas/</code>, without rewriting committed files:

- <code>schemas/easyp-v1.schema.json</code> and <code>schemas/easyp.schema.json</code>
- <code>schemas/easyp.gen-v1.schema.json</code> and <code>schemas/easyp.gen.schema.json</code>
- <code>schemas/protobuf.lock-v1.schema.json</code> and <code>schemas/protobuf.lock.schema.json</code>

Do not hand-edit those generated JSON files.

### Test Mocks

~~~bash
task mocks
~~~

<code>task mocks</code> configures Mockery for <code>Rule</code>, <code>Console</code> and
<code>CurrentProjectGitWalker</code>, all with <code>--dir ./internal/core</code> and output to
that package's generated mocks directory. Only <code>Rule</code> and
<code>CurrentProjectGitWalker</code> still belong there: <code>Console</code> is defined in
<code>internal/adapters/console/new.go</code>. This stale Taskfile target is unresolved;
do not assume the aggregate command succeeds. There are no checked-in
Mockery outputs or <code>go:generate</code> directives in the current source. Module
ports use hand-written test doubles; see [TESTING.md](TESTING.md).

## Docker and Release Checks

~~~bash
task docker_base GIT_TAG=v0.0.0
task docker_lint GIT_TAG=v0.0.0
task goreleaser:check
task goreleaser:test
task goreleaser:test-docker
~~~

<code>docker_base</code> and <code>docker_lint</code> are configured to build and tag the base and
lint images, but their missing Dockerfiles make their preconditions fail.
<code>task docker</code> obtains the current Git tag, builds both images, and pushes
them, so it is a publishing command rather than a local smoke test.

<code>task goreleaser:check</code> validates GoReleaser configuration. The snapshot
tasks create or select a <code>multiarch</code> Docker Buildx builder and run
<code>goreleaser release --snapshot --clean</code>; they do not publish a release.
The release workflow runs for stable <code>vMAJOR.MINOR.PATCH</code> tags and authenticates
to GitHub Container Registry and Docker Hub.

## Dependency and Vendor Tools

EasyP module dependencies are Git repositories, cached under <code>EASYPPATH</code>
(default <code>$HOME/.easyp</code>) and locked in <code>protobuf.lock</code>. The vendoring command is:

~~~bash
./easyp mod vendor
~~~

It verifies the v1 lock, then copies <code>.proto</code> files from locked dependency
roots to <code>easyp_vendor</code> by import path. It rejects local replacements and
duplicate import paths. This is a materialized dependency tree, separate from
the module cache and Go's usual <code>vendor/</code> directory. See
[config/dependency.md](config/dependency.md) for the authoritative dependency
flow and available <code>easyp mod</code> subcommands.

## Native v1 Example

<code>task proto:check</code> invokes <code>scripts/check-proto-example.sh</code>. The script builds
a temporary CLI, copies the four repository v1 files and <code>examples/proto</code>,
validates configuration, tidies and compares the lock, lints, then generates
twice and compares descriptors and generated output. Its work and cache are
isolated in a temporary directory. It runs in the test workflow.

## Tool Installation

| Tool | Installation used by this repository |
|---|---|
| Go | Install Go 1.26.6, then verify with <code>go version</code> |
| Task | Install Task 3.x; CI uses <code>arduino/setup-task@v1</code> |
| GolangCI-Lint | <code>task init</code> installs it into <code>./bin</code> |
| Gotestsum | <code>task init</code> installs it into <code>./bin</code> |
| Mockery | <code>task init</code> installs it into <code>./bin</code> |
| Hadolint | <code>task lint</code> pulls <code>ghcr.io/hadolint/hadolint:v2.12.1-beta</code> |
| GoReleaser | Provide <code>goreleaser</code> on <code>PATH</code> for its Task targets |

<code>task clean</code> removes <code>bin/</code> and <code>coverage.out</code>, then clears the Go build
cache. It does not remove module caches, EasyP's dependency cache, schemas,
or the root-level <code>easyp</code> binary.
