# AGENTS.md

Instructions for AI coding agents working in this repository.

## Project overview

EasyP is a Protocol Buffers CLI toolkit (<code>github.com/easyp-tech/easyp</code>, Go 1.26.6) that provides:

- linting (buf-compatible rules)
- breaking change detection
- code generation (local and remote plugins)
- Git-based dependency / package management (<code>easyp mod</code>)

Human-facing docs: [README.md](README.md) and https://easyp.tech.

## Layout

| Path | Role |
|------|------|
| <code>cmd/easyp</code> | CLI entrypoint |
| <code>internal/api</code> | CLI command wiring |
| <code>internal/core</code> | Proto parsing, lint/breaking engines and low-level plugin execution |
| <code>internal/modules</code> | Dependency resolution, manifest/lock operations, import roots and vendoring |
| <code>internal/generation</code> | Consumer-project discovery, module selection, inherited generation options and descriptor export |
| <code>internal/migration</code> | Legacy-to-v1 migration plans, validation and guarded apply |
| <code>internal/rules</code> | Lint rules (each rule + colocated <code>&lt;file&gt;_test.go</code>) |
| <code>internal/config</code> | Shared engine configuration and validation issue types |
| <code>internal/adapters</code> | Git module cache, dependency metadata, legacy modfile reader, plugins, console and prompts |
| <code>internal/adapters/gitmodules</code> | Verified Git checkouts and installed modules under the v1 cache |
| <code>mcp/easypconfig</code> | MCP descriptions of v1 <code>easyp.yaml</code> and <code>easyp.gen.yaml</code> |
| <code>internal/config/v1</code> | v1 parsing, validation, and JSON Schema source |
| <code>internal/schemagen</code> | Writes the six versioned/latest v1 JSON Schema artifacts |
| <code>schemas/</code> | Generated policy, generation and lock JSON Schemas |
| <code>easyp-tech/docs</code> (separate repository) | Documentation site and its publishing workflow |

## What is .spec/
.spec/ is a project documentation directory optimized for AI agent (LLM) consumption. It contains:

* Structured descriptions of architecture, packages, and domain model
* Code conventions and testing patterns
* Infrastructure and tooling descriptions
* [Agent rules](.spec/agent-rules.md) with mandatory coding standards
* Purpose: give the agent project context without reading the entire source code. Start with [.spec/README.md](.spec/README.md), then verify relevant details against source before changing behavior.


## Build, test, lint

Prefer Task targets from [<code>Taskfile.yml</code>](Taskfile.yml):

~~~sh
task init              # install local tools (golangci-lint, gotestsum, mockery)
task build             # go build -o easyp ./cmd/easyp
task test              # gotestsum with -race and coverage
task lint              # Go lint + hadolint on root Dockerfile
task lint:go           # Go lint without Docker
task lint:docker       # hadolint on root Dockerfile
task quality           # test + lint
task schema:generate   # regenerate schemas/*.json
task schema:check      # generate and fail if schemas drift
task mocks             # optional core and console mocks
task dev-tools:check   # offline Taskfile dispatch smoke test
task docker:build      # local root Dockerfile build; no publish
~~~

Equivalents without Task:

~~~sh
go build -o easyp ./cmd/easyp
go test -race -count=1 ./...
go run ./cmd/easyp schema-gen
~~~

After behavior changes, run the relevant tests (at least the packages you touched). For config schema changes, run <code>task schema:check</code> before finishing.

<code>task init</code> installs pinned helpers into the repository's <code>bin/</code>, including when invoked from a nested directory, and fetches dependencies with <code>go mod download</code>. <code>task lint</code> attempts both checks and fails if either fails; Docker must be available for Hadolint. <code>task docker</code> aliases the local <code>docker:build</code> target, with an overridable <code>DOCKER_IMAGE</code> tag (default <code>easyp:local</code>). Release publishing stays in the release workflow. See [.spec/TOOLS.md](.spec/TOOLS.md).

## Conventions

- **New lint rules**: add <code>internal/rules/&lt;rule&gt;.go</code> + <code>&lt;file&gt;_test.go</code>, register via the existing rule builder; mirror patterns in neighboring rules.
- **v1 YAML schemas**: change types and schema rules in <code>internal/config/v1</code>, then regenerate <code>schemas/</code> with <code>task schema:generate</code>. The MCP tool reads the same schema through <code>v1.SchemaJSON</code>; update its field descriptions when behavior changes. Do not hand-edit generated schema JSON.
- **Tests**: use <code>testify</code>; update test doubles when interfaces change. Optional <code>task mocks</code> targets <code>core.Rule</code>, <code>core.CurrentProjectGitWalker</code> and <code>console.Console</code> in <code>internal/adapters/console</code>. Existing tests use handwritten doubles; mock generation is not a build/test prerequisite. Compile generated outputs before using them and do not commit unused mocks.
- **v1 files**: policy is <code>easyp.yaml</code>; generation is <code>easyp.gen.yaml</code>; native <code>protobuf.mod</code> declares dependencies; <code>protobuf.lock</code> records verified revisions. Use <code>easyp migrate</code> for a legacy migration preview, <code>easyp migrate --interactive</code> for the wizard, and <code>--write</code> for explicit apply; see [.spec/CLI.md](.spec/CLI.md).
- **Implementation boundaries**: do not treat reserved <code>generate.packages</code>, or non-<code>FILE</code> breaking categories as implemented. Local replacements use a non-persisted overlay; explicit frozen mode verifies only the published graph and unknown-import-to-module discovery is intentionally excluded; [.spec/ERRORS.md](.spec/ERRORS.md) separates those gaps from actual CLI exit behavior.
- Prefer pointing agents/humans at README and docs over copying long usage text into this file.

## Do not

- Commit secrets, credentials, or unrelated local artifacts.
- Hand-edit generated schema JSON: <code>schemas/easyp-v1.schema.json</code>, <code>schemas/easyp.schema.json</code>, <code>schemas/easyp.gen-v1.schema.json</code>, <code>schemas/easyp.gen.schema.json</code>, <code>schemas/protobuf.lock-v1.schema.json</code>, <code>schemas/protobuf.lock.schema.json</code>.
- Commit or “fix” generated dependency/build noise, <code>coverage.out</code>, or the local <code>easyp</code> binary in the repo root. The documentation site lives in the separate docs repository.
- Add parallel agent instruction files (<code>.cursor/rules/</code>, <code>CLAUDE.md</code>) unless explicitly requested; keep this file the source of truth.

Shared producer policies use section-scoped <code>extends</code> from bounded local files or already declared/locked modules. See [.spec/config/policy-extends.md](.spec/config/policy-extends.md). No standalone policy versions or unpinned reference downloads are allowed.
