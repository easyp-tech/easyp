# EasyP v1 CLI Commands Reference

Verified against `easyp --help` of `v1.0.0-nightly.20261008.1`. Running a v0 binary? Its flags differ: see [v0 → v1 changes](#v0--v1-command-changes).

```text
easyp
├── lint (l)                    Lint proto files with the effective easyp.yaml policy
├── generate (g)                Run easyp.gen.yaml generators
├── breaking                    Compare API compatibility against a Git ref
├── get                         Add/promote a requirement in protobuf.mod and lock it
├── mod (m)
│   ├── download                Install exactly what protobuf.lock pins
│   ├── update                  Upgrade within current majors, rewrite mod + lock
│   ├── tidy                    Resolve requirements, add indirects, write the lock
│   └── vendor                  Copy locked dependency protos into easyp_vendor/
├── init (i)                    Create protobuf.mod, easyp.yaml, easyp.gen.yaml
├── migrate                     Preview or apply v0 → v1 conversion
├── ls-files (ls)               List module sources and reachable imports
├── validate-config (validate)  Validate all four v1 files, recursively
├── schema-gen                  Write JSON Schemas for the v1 YAML files
└── completion bash|zsh
```

## Global flags

Put them before the command: `easyp --format json lint`.

| Flag | Short | Default | Env | Meaning |
|---|---|---|---|---|
| `--cfg`, `--config` | — | discovery | `EASYP_CFG` | Explicit `easyp.yaml` for lint/breaking; file or directory for `validate-config` |
| `--debug` | `-d` | `false` | `EASYP_DEBUG` | Debug logs on stderr |
| `--format` | `-f` | `text` (`json` for ls-files, validate-config) | `EASYP_FORMAT` | `text` or `json` |
| `--frozen` | — | `false` | — | Use the existing `protobuf.mod` + `protobuf.lock` exactly. No resolution, no file writes. **Never** inferred from `CI=true`. |

Other environment variables:
- `EASYPPATH`: cache root, default `$HOME/.easyp`; v1 uses `$EASYPPATH/v1/…`.
- `EASYP_INIT_DIR`: default for `init --dir`.

## lint

```bash
easyp lint                          # nearest easyp.yaml, its whole subtree
easyp lint --path proto             # only files under proto/
easyp --format json lint > issues.jsonl
```

| Flag | Default | Meaning |
|---|---|---|
| `--path`, `-p` | `.` | Source path relative to the operation root |
| `--root`, `-r` | policy directory | Override the operation root |
| `--frozen` | `false` | See global flag |

- **Policy discovery:** it uses the nearest ancestor `easyp.yaml`. Descendant `easyp.yaml` files govern their own subtrees, so a nested **v0** file fails the whole run.
- **Output:** text is `path:line:col:source message (RULE)`. JSON is one object per line (JSONL), not an array.
- **Exit codes:**
  - 0: clean;
  - 1: findings or any error;
  - 2: an import cannot be opened.

## generate

```bash
easyp generate                                        # nearest ancestor easyp.gen.yaml
easyp generate --project services/api --project services/web
easyp generate --all                                  # every easyp.gen.yaml below cwd
easyp --frozen generate                               # CI: no dependency resolution
easyp generate --descriptor_set_out api.binpb --include_imports
easyp generate --all --descriptor_set_out_dir descriptors
```

| Flag | Meaning |
|---|---|
| `--project DIR` | Select a generator directory (repeatable). Exclusive with `--all`. |
| `--all` | Recursively run every generator below the working directory. Skips hidden dirs, `easyp_vendor`, `node_modules`, nested Git repos. |
| `--workspace DIR` | Root for repository-relative module selectors (default: enclosing Git repo) |
| `--descriptor_set_out FILE` | One combined binary `FileDescriptorSet` |
| `--descriptor_set_out_dir DIR` | One descriptor set per project/module; exclusive with the above |
| `--include_imports` | Include transitive dependencies in descriptor sets |

- **Removed in v1:** `--path/-p`, `--root/-r` and `EASYP_ROOT_GENERATE_PATH`. Targets now come from `easyp.gen.yaml` (`generate.modules`, `paths`, `packages`).
- Plugin `out` paths are relative to the generator file. Descriptor paths are relative to the working directory.
- Error `no selected easyp.gen.yaml`: run from inside a project, or pass `--project` / `--all`.

## breaking

```bash
easyp breaking                          # against breaking.baseline from easyp.yaml, else master
easyp breaking --against main
easyp breaking --path proto --against origin/main
easyp --format json breaking --against HEAD~1
```

| Flag | Default | Meaning |
|---|---|---|
| `--against` | `breaking.baseline`, then `master` | Raw Git ref (no `git:` prefix on the CLI) |
| `--path`, `-p` / `--root`, `-r` | `.` / policy dir | Same as lint |

- **How revisions are compared:** each revision is compiled with its *own* manifest and lock. A baseline that still has v0 config works only for projects without dependencies; otherwise the run fails with `compile baseline descriptors: … Resolve …`.
- **Profiles:** set by `breaking.categories` (default FILE); see [breaking-checks.md](./breaking-checks.md).
- **Exit codes:**
  - 0: compatible;
  - 1: breaking changes or errors;
  - 2: ref or repository not found, or an import cannot be opened.
- **CI:** needs full history (`fetch-depth: 0`).

## get

```bash
easyp get github.com/googleapis/googleapis                     # HEAD, pinned to a commit in the lock
easyp get github.com/acme/contracts@v1.2.3                     # semver tag
easyp get github.com/acme/contracts@3f9c…(full commit)
easyp get github.com/googleapis/googleapis@common-protos-1_3_1 # named tag → full commit in protobuf.mod
easyp get --import-root proto gitlab.com/acme/svc              # dependency's import root hint (repeatable)
```

- **What it does:**
  - adds a new direct requirement, or promotes an `// indirect` one;
  - resolves the transitive graph;
  - checks every import of the module;
  - writes `protobuf.mod` + `protobuf.lock`.
- **Versions:**
  - branch names are not accepted;
  - v2+ needs a `/vN` identity, or gets `+incompatible` automatically for pre-native repositories.
- **Pitfall:** adding several dependencies one at a time fails while the others are still missing (`cannot resolve imports`). Put them all in `protobuf.mod` and run `easyp mod tidy`.

## mod

All subcommands use the nearest ancestor `protobuf.mod` and accept `--frozen`.

| Command | Behavior |
|---|---|
| `easyp mod download` | Validate the lock against the manifest, then install the exact pins (hash-verified). Missing lock → run `mod tidy`. |
| `easyp mod tidy` | Resolve `require`s (versionless → HEAD on first resolve, then kept), add/mark `// indirect`, write the lock. Also repairs consumer `import` paths after a dependency moved files, but only when the mapping is verified. Never upgrades. Refuses `--frozen`. |
| `easyp mod update` | Move semver requirements to the newest tag in the same major, refresh versionless ones to HEAD, keep commit pins; rewrite mod + lock. Refuses `--frozen`. |
| `easyp mod vendor` | Copy locked dependency protos into `easyp_vendor/`. |

With `replace` directives present, operations do not write the lock (it stays byte-identical). Remove replacements and run `mod tidy` before committing.

## init

```bash
easyp init --module github.com/acme/contracts           # in the repo root
easyp init --dir proto/billing --module github.com/acme/contracts/proto/billing
```

- Creates `protobuf.mod` (`module …`), `easyp.yaml` (`linters.default: STANDARD`, `breaking.baseline: git:main`) and `easyp.gen.yaml` (`plugins: []`).
- No lock: run `easyp mod tidy` (or `get`) to create one.
- **Identity order:**
  1. `--module`;
  2. an existing manifest;
  3. the `origin` remote (only at the repo root);
  4. a prompt (terminal only).
- Asks before overwriting differing files. Non-interactive runs fail instead.
- Rejects directories with `buf.yaml`; see [migration-from-buf.md](./migration-from-buf.md).

## migrate

```bash
easyp migrate --module github.com/acme/contracts                       # preview only
easyp migrate --module github.com/acme/contracts --resolve-lock        # preview + verify dependency pins
easyp migrate --module github.com/acme/contracts --resolve-lock --write
easyp migrate --dir examples/demo --module github.com/acme/contracts/examples/demo
easyp migrate                                                          # interactive wizard (terminal only)
```

| Flag | Meaning |
|---|---|
| `--dir` | Directory with the legacy `easyp.yaml` (default `.`) |
| `--module` | Canonical identity for `protobuf.mod`; required without the wizard |
| `--resolve-lock` | Allow network/cache access to verify `easyp.lock` pins and write `protobuf.lock` |
| `--write` | Apply the candidates. Backs up to `*.v0.bak`, keeps `easyp.lock`, never overwrites differing native files. |
| `--interactive` | Force (`true`) or disable (`false`) the wizard |

- Never runs plugins. Rejects `--frozen`.
- Full procedure, blockers and fixes: [migration-v0-to-v1.md](./migration-v0-to-v1.md).

## ls-files

```bash
easyp ls-files                                  # JSON: files, roots, errors
easyp ls-files --include-imports=false --format text
```

- **Reads:** the manifest in the **current directory**; it does not search upward.
- **Fields:**
  - each file has `import_path`, `source` (`workspace` / `dependency` / `wellknown`) and `abs_path`;
  - well-known types appear under `/wellknownimports`.
- **Exit code:** 1 if parse/import errors were collected.

## validate-config

```bash
easyp validate-config                                   # whole tree below cwd, JSON
easyp validate-config --format text                     # VALID: true|false + table
easyp validate-config --config ./services/api           # one file or directory
```

- **Scope:** validates `easyp.yaml`, `easyp.gen.yaml`, `protobuf.mod` and `protobuf.lock` under the selected path, **hidden directories included**. A template named `easyp.yaml` under `.agents/` or `docs/` counts as config.
- **Checks:** structure and semantics only. It does not fetch dependencies or run plugins.
- **Exit code:** 1 when invalid. Legacy files report `legacy EasyP configuration detected`.

## schema-gen

```bash
easyp schema-gen --out-dir schemas
```

Writes `easyp{,-v1}.schema.json`, `easyp.gen{,-v1}.schema.json` and `protobuf.lock{,-v1}.schema.json`. `protobuf.mod` has no JSON Schema.

Wire one into an editor with a YAML comment:

```yaml
# yaml-language-server: $schema=./schemas/easyp.schema.json
```

## completion

```bash
easyp completion bash > "$(brew --prefix)/etc/bash_completion.d/easyp"
easyp completion zsh  > "${fpath[1]}/_easyp"
```

## Exit codes (summary)

| Code | Meaning |
|---|---|
| 0 | Success (also: wizard confirmation declined) |
| 1 | Lint/breaking findings; any returned error (invalid config, module/cache, usage, migration); ls-files collected errors; validate-config invalid |
| 2 | lint/breaking only: missing import file; breaking: Git ref or repository not found |

## v0 → v1 command changes

| v0 | v1 |
|---|---|
| `easyp generate -p proto -r .` | `easyp generate` / `--project DIR` / `--all` |
| `EASYP_ROOT_GENERATE_PATH` | removed |
| `easyp mod download` reads `deps` from `easyp.yaml` | reads `protobuf.mod` + `protobuf.lock` |
| `easyp mod update` | same name, writes `protobuf.mod` + `protobuf.lock` |
| — | `easyp get`, `easyp mod tidy`, `easyp migrate`, `--frozen` |
| `easyp init` interactive rule-group wizard, buf import | writes 3 files with defaults; rejects `buf.yaml` |
| `easyp schema-gen --out-versioned/--out-latest` | `easyp schema-gen --out-dir` |
| `easyp validate-config` checks one `easyp.yaml` | recursive over all v1 files |
| `breaking --against` defaults to `master` | `breaking.baseline`, then `master` |
| `lint/breaking -p/-r` | unchanged |
