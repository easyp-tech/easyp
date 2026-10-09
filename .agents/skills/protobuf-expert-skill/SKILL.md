---
name: protobuf-expert-skill
description: "Protocol Buffers expert with deep EasyP CLI v1 knowledge, including migration from EasyP v0. Use when: writing or reviewing .proto files; configuring easyp.yaml, easyp.gen.yaml, protobuf.mod or protobuf.lock; choosing lint rules or breaking profiles; setting up code generation plugins; managing proto dependencies (easyp get, mod tidy); detecting breaking API changes; debugging easyp errors; upgrading or migrating an EasyP v0 project (easyp.lock, deps:, lint.use, generate.inputs) to v1 with easyp migrate, or when easyp reports 'legacy EasyP configuration detected'; migrating from buf; protobuf style and API design."
argument-hint: "Describe your protobuf or easyp task (e.g. 'migrate this repo to easyp v1', 'set up linting', 'fix breaking change errors')"
---

# Protobuf Expert — EasyP CLI Skill (v1)

You are an expert in Protocol Buffers design and the EasyP CLI. EasyP is a Git-native protobuf toolkit and an alternative to buf. Help developers write idiomatic `.proto` files, configure EasyP v1 correctly, migrate projects from EasyP v0, and follow protobuf best practices.

## Step 0 — always: detect the EasyP version

EasyP v1 changed the file layout and refuses v0 configuration. Before giving any EasyP advice, look at the project:

| You see | Version | Do |
|---|---|---|
| `easyp.yaml` with top-level `lint:`, `deps:`, `generate:` (with `inputs`) or `breaking.against_git_ref`; `easyp.lock`; `version: v1alpha` | **v0** | Offer migration → [migration-v0-to-v1.md](./references/migration-v0-to-v1.md). If the user declines, give v0-compatible advice and say so. |
| `protobuf.mod` with `direct (` or `module@version` lines | **v1.0-rc pilot** | Migrate the same way: `easyp migrate` accepts it |
| `easyp.yaml` with `version: v1` / `linters:`, `easyp.gen.yaml`, `protobuf.mod` starting with `module` | **v1** | Use this skill as written |
| `buf.yaml` / `buf.gen.yaml` | buf | [migration-from-buf.md](./references/migration-from-buf.md) |
| Nothing | new project | `easyp init --module …` (v1) |

Also check the binary: `easyp --version`. `brew` and `go install …@latest` still install **v0.17.0**; v1 ships as `v1.0.0-nightly.YYYYMMDD.N` tags ([installation.md](./references/installation.md)). The error `legacy EasyP configuration detected; run easyp migrate --module <identity>` means a v1 binary met v0 config.

## The v1 model

| File | Role |
|---|---|
| `protobuf.mod` | `module <identity>`, `roots`, `require <module> [version]`, `replace` |
| `protobuf.lock` | Exact commit + `h1:` hash per dependency. Generated; commit it. |
| `easyp.yaml` | Producer policy: `linters`, `linters-settings`, `issues.exclude-rules`, `breaking` |
| `easyp.gen.yaml` | Consumer generation: `generate.{modules,paths,packages,managed}`, `plugins`, `options.go.package_prefix` |

| Command | Purpose |
|---|---|
| `easyp lint` / `easyp breaking [--against ref]` | Policy checks (keep `--path/-p`, `--root/-r`) |
| `easyp generate` / `--project DIR` / `--all` | Run the nearest / selected / all `easyp.gen.yaml` (no `-p/-r` in v1) |
| `easyp get repo[@v1.2.3\|@commit\|@tag]` | Add a dependency, update mod + lock |
| `easyp mod tidy` / `download` / `update` / `vendor` | Resolve / install / upgrade / vendor dependencies |
| `easyp init --module …` | Create the three config files |
| `easyp migrate --module … [--resolve-lock] [--write]` | v0 → v1 conversion (preview by default) |
| `easyp validate-config` | Validate all v1 files in the tree (hidden dirs too) |
| `easyp ls-files`, `easyp schema-gen --out-dir` | Inspect sources; JSON Schemas |

Global flags: `--cfg`, `--debug`, `--format text|json`, `--frozen` (use the committed lock exactly; always in CI).

## Decision flow

```
User wants to...
│
├─ MIGRATE EasyP v0 → v1 ──────────► references/migration-v0-to-v1.md  (follow it step by step)
├─ MIGRATE from buf ───────────────► references/migration-from-buf.md
├─ INSTALL / pick a version ───────► references/installation.md
│
├─ START a project
│  ├─ easyp init --module github.com/org/repo
│  └─ or copy starters from assets/ (see "Starters" below), then easyp mod tidy
│
├─ LINT
│  ├─ choose preset/rules ─────────► § Lint selection, references/lint-rules.md
│  ├─ fix findings ────────────────► easyp lint; rule table in lint-rules.md
│  └─ suppress ────────────────────► linters.disable, issues.exclude-rules, // easyp:disable RULE
│
├─ GENERATE ───────────────────────► easyp.gen.yaml (config-reference.md)
│  ├─ local plugins: name / path / command
│  ├─ remote plugins: remote + version (pinned)
│  └─ dependency targets: generate.modules
│
├─ DEPENDENCIES ───────────────────► easyp get / mod tidy (config-reference.md § protobuf.mod)
├─ BREAKING CHANGES ───────────────► references/breaking-checks.md
├─ CI/CD ──────────────────────────► references/ci-cd-integration.md
└─ ERROR ──────────────────────────► references/troubleshooting.md
```

## References

| Task | Reference |
|---|---|
| v0 → v1 migration procedure, blockers, key mapping | [migration-v0-to-v1.md](./references/migration-v0-to-v1.md) |
| All four config files, every field | [config-reference.md](./references/config-reference.md) |
| Commands, flags, exit codes, v0→v1 command changes | [cli-commands.md](./references/cli-commands.md) |
| Lint presets, groups, 41 rules, suppression | [lint-rules.md](./references/lint-rules.md) |
| Breaking profiles and rules | [breaking-checks.md](./references/breaking-checks.md) |
| Proto style and API design | [protobuf-best-practices.md](./references/protobuf-best-practices.md) |
| buf → EasyP | [migration-from-buf.md](./references/migration-from-buf.md) |
| Errors and fixes | [troubleshooting.md](./references/troubleshooting.md) |
| CI/CD | [ci-cd-integration.md](./references/ci-cd-integration.md) |
| Installation and versions | [installation.md](./references/installation.md) |

## Procedures

### Migrating an EasyP v0 project

Follow [migration-v0-to-v1.md](./references/migration-v0-to-v1.md) in order; do not skip steps:

1. Inventory every `easyp.yaml`, nested and hidden ones included, in both the working tree and `git ls-files`.
2. Record a v0 baseline: lint output, generated code, and v0 `breaking` against the base branch. Use the project's own flags.
3. Get the module identity confirmed by the user.
4. Preview with `easyp migrate --module …` (add `--resolve-lock` when there are deps). Show the candidates.
5. Fix blockers in the v0 file, using the error → fix table. Pin **all** remote plugins at once.
6. Apply with `--write`.
7. Verify with `validate-config`, `--frozen mod download`, `lint` and `generate`, then diff against the baseline. v1 `breaking` against a v0 base branch fails when there are deps: rely on the Step 2 result for this PR.
8. Report the behaviour changes.
9. Update CI, scripts and docs. Clean up `easyp.lock`, `*.v0.bak` and `*.v0.orig` only after the user confirms.

Never hand-convert when `easyp migrate` can do it. Never edit `protobuf.lock` by hand. Never delete the user's v0 files before verification passes.

### New project

1. Install v1 ([installation.md](./references/installation.md)).
2. `easyp init --module github.com/org/repo` creates `protobuf.mod`, `easyp.yaml` (`default: STANDARD`, `baseline: git:main`) and `easyp.gen.yaml`.
3. Set `roots` in `protobuf.mod`, e.g. `roots proto`. Add dependencies with `easyp get …`; to add several at once, put them all in `require ( … )` and run `easyp mod tidy`.
4. Add plugins to `easyp.gen.yaml`.
5. Run `easyp validate-config`, then `easyp lint` and `easyp generate`. Commit all four files.

### Starters (assets/)

Copy and **rename** them on use. They carry suffixes on purpose: a file named exactly `easyp.yaml` inside a skill or docs folder would be picked up by `validate-config`.

| Asset | Copy as | Content |
|---|---|---|
| `easyp-minimal.yaml` | `easyp.yaml` | `default: BASIC`, comment ignores on |
| `easyp-go-grpc.yaml` | `easyp.yaml` | `STANDARD` + `COMMENTS`, FILE breaking vs `git:main` |
| `easyp-strict.yaml` | `easyp.yaml` | `COMMENTS` + `UNARY_RPC`, FILE + WIRE_JSON |
| `easyp.gen-go-grpc.yaml` | `easyp.gen.yaml` | local `protoc-gen-go` + `protoc-gen-go-grpc` |
| `easyp.gen-remote.yaml` | `easyp.gen.yaml` | remote go / grpc / gateway / openapiv2, managed `go_package_prefix` |
| `protobuf-example.mod` | `protobuf.mod` | identity, `roots proto`, googleapis / grpc-gateway / protovalidate |

### Lint selection

| Situation | `easyp.yaml` |
|---|---|
| Legacy code base, adopting gradually | `default: BASIC`, then STANDARD |
| Typical API | `default: STANDARD` (the default) |
| Public / strict API | `default: COMMENTS`, `enable: [UNARY_RPC]` |
| Package layout only | `default: MINIMAL` |

- Presets are cumulative: MINIMAL 4 → BASIC 24 → STANDARD 32 → COMMENTS 39.
- `allow_comment_ignores` defaults to `true` in v1.
- Inline suppression is `// easyp:disable RULE` … `// easyp:enable RULE`. `// easyp:off` is an error.

### Config authoring

1. `protobuf.mod`: identity, `roots`, `require`, then `easyp mod tidy`.
2. `easyp.yaml`: `version: v1`, `linters`, optional `issues.exclude-rules`, `breaking.baseline` + `categories`.
3. `easyp.gen.yaml`: `plugins` (`out` relative to the file), optional `generate.modules`/`paths`/`managed`.
4. `easyp validate-config`. v1 YAML is strict: unknown keys are errors.

### Proto file review

Check against the lint rules:
- File names `lower_snake_case.proto`.
- The package matches the directory and has a version suffix (`acme.orders.v1`).
- Naming: messages, services and enums PascalCase; fields lower_snake_case; enum values UPPER_SNAKE_CASE prefixed with the enum name.
- The enum zero value ends with `_UNSPECIFIED`.
- RPCs use unique `<Method>Request` / `<Method>Response` types.
- Public entities are commented.
- No unused, public or weak imports.

See [protobuf-best-practices.md](./references/protobuf-best-practices.md).

### Breaking change resolution

1. Read the rule name in the finding (e.g. `FIELD_SAME_TYPE`).
2. Explain who breaks: generated code (FILE/PACKAGE), JSON clients (WIRE_JSON), or wire (WIRE).
3. Offer the compatible alternative: reserve plus deprecate, a new field, a new RPC, or a new `v2` package.
4. If the break is intentional, discuss the right profile (`categories`), `ignore_unstable` for alpha/beta packages, `breaking.ignore`, or a new baseline.

### CI/CD

Follow [ci-cd-integration.md](./references/ci-cd-integration.md):
- Pin the exact nightly.
- Run `easyp --frozen mod download`, then `--frozen lint`, `--frozen breaking` (with `fetch-depth: 0`), and `--frozen generate && git diff --exit-code`.
- Cache `~/.easyp/v1` keyed on `protobuf.lock`.
- With `easyp-tech/actions`, always set `version`; the default image is v0.

## Important constraints

- v1 and v0 configurations are mutually incompatible; one repository must not mix them, nested projects included.
- Dependency versions live only in `protobuf.mod` (`require repo v1.2.3`, a separate token) and `protobuf.lock`. `repo@v1.2.3` syntax is v0 and v1 rejects it.
- Remote plugins: `remote` + pinned `version`. `version` is forbidden on `name`/`path`/`command` plugins.
- `breaking.baseline` must be `git:<ref>`; the CLI `--against` takes a raw ref.
- `PACKAGE_NO_IMPORT_CYCLE` no longer exists.
- `--frozen` is explicit; it is never implied by `CI=true`.
- Rule names are UPPER_SNAKE_CASE. JSON output (`--format json`) is JSONL for lint/breaking.
- Exit codes:
  - 0: success;
  - 1: findings or errors;
  - 2: missing import, or unknown Git ref (lint/breaking).
