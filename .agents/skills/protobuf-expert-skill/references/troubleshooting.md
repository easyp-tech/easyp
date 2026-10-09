# EasyP v1 Troubleshooting Guide

Start with:

```bash
easyp --version                         # v1.0.0-nightly.… vs v0.x — most confusion starts here
easyp validate-config --format text     # all four files, whole tree
easyp --debug <command>                 # verbose logs on stderr
```

## Version mismatch

| Symptom | Cause | Fix |
|---|---|---|
| `legacy EasyP configuration detected; run easyp migrate --module <identity> …` | v1 binary, v0 `easyp.yaml` (top-level `lint`/`deps`/`generate`) | Migrate: [migration-v0-to-v1.md](./migration-v0-to-v1.md). Or use a v0 binary for now. |
| Same error for a file you did not expect (`examples/…/easyp.yaml`, `.agents/…/easyp.yaml`) | `validate-config` scans every directory; `lint` reads descendant policies | Migrate that nested project, or rename templates that are not real configs |
| v0 binary: unknown field `linters` / `version: v1`, or it ignores `easyp.gen.yaml` | v0 binary, v1 config | Install a v1 nightly ([installation.md](./installation.md)); check that CI does not pin v0 |
| `brew upgrade` / `go install …@latest` still gives v0.17.0 | v1 ships only as nightly tags | Pin `@v1.0.0-nightly.YYYYMMDD.N` explicitly |

## Configuration errors

| Message (substring) | Fix |
|---|---|
| `easyp.yaml version must be v1` | `version: v1` (or omit it) |
| `unknown v1 linter preset` | `linters.default`: `MINIMAL`, `BASIC`, `STANDARD` or `COMMENTS`. Put `DEFAULT` / `UNARY_RPC` in `enable`. |
| `invalid rule: PACKAGE_NO_IMPORT_CYCLE is not implemented` | Rule removed in v1; delete it |
| `unsupported linters-settings rule` | Only `ENUM_ZERO_VALUE_SUFFIX.suffix` and `SERVICE_SUFFIX.suffix` exist |
| `breaking.baseline must be empty or git:<ref>` | `baseline: git:main`; the CLI flag `--against main` takes a raw ref |
| `breaking.categories: unknown profile` | `FILE`, `PACKAGE`, `WIRE_JSON`, `WIRE` |
| `exactly one of name, path, command or remote is required` | One plugin source per entry |
| `specify the remote plugin version only in version: split remote into …` | `remote: host/org/plugin` + `version: vX.Y.Z` |
| `remote plugin requires a pinned semantic version` | Add `version: vX.Y.Z`; `latest` is rejected |
| `local and bundled plugin versions are selected by the executable` | Remove `version` from `name`/`path`/`command` plugins |
| `field … not found in type` (YAML decode) | Unknown key; v1 files are strict. Compare with [config-reference.md](./config-reference.md). |
| `malformed lint directive "easyp:off"` | v1 supports only `// easyp:disable RULE` / `// easyp:enable RULE` (plus `buf:lint:ignore`, `nolint:`) — see [lint-rules.md](./lint-rules.md#inline-comment-directives) |
| `unknown or unsupported lint rule "X" in comment directive` | Typo in a rule name inside an inline directive |
| `suppression X must annotate a declaration or have a matching easyp:enable` | Put the directive directly above a declaration, or close it with `easyp:enable X` |

## Dependencies

| Message (substring) | Cause | Fix |
|---|---|---|
| `cannot resolve imports [...]` | An import is not provided by any root or `require` | Add the module to `protobuf.mod` (`easyp get repo`), or fix `roots`. Several missing at once: write them all and run `easyp mod tidy`. |
| lock missing during `mod download` | No `protobuf.lock` yet | `easyp mod tidy`, then commit both files |
| `--frozen` rejects a missing/stale lock or `replace` | Lock out of date, or a local overlay is present | Outside frozen mode: remove `replace`, `easyp mod tidy`, commit `protobuf.mod` + `protobuf.lock` |
| `require module@v1.2.3` rejected | The version is a separate token | `require module v1.2.3` |
| Major-version / `+incompatible` errors | v2+ identities need `/vN`; pre-native v2+ repos need `+incompatible` | `easyp get repo@v2.3.0` adds the right form automatically |
| `locked version changed` (`ErrLockedVersionChanged`) | A semver tag now points to a different commit | Someone moved a tag. Investigate before re-pinning. |
| `using BSR compatibility snapshot; BSR revision equivalence and digest are not verified` (warning) | A dependency's `buf.yaml` declares BSR deps, mapped to fixed Git snapshots (googleapis, grpc-gateway, protovalidate, protoc-gen-validate, grpc-federation) | Informational. An unknown BSR module fails with `unsupported BSR module`. |
| `incompatible requirement X@A: historical commit is B` | Two pins for the same module disagree (often the BSR snapshot vs your pin) | Align the versions. During migration, see the [re-resolution fallback](./migration-v0-to-v1.md#fallback-re-resolve-dependencies). |
| Private Git dependency fails to fetch | EasyP uses system `git` | Configure git credentials or `url.<base>.insteadOf` as for `git clone` |

## Generation

| Symptom | Fix |
|---|---|
| `no selected easyp.gen.yaml` | Run inside a project directory, or `easyp generate --project DIR` / `--all` |
| `generate -p/-r` flags rejected | Removed in v1. Targets come from `easyp.gen.yaml` (`generate.paths`, `modules`, `packages`). |
| Nothing generated for a dependency | Dependencies are import-only; list them in `generate.modules` to generate them |
| Files outside a directory are no longer generated after migration | `generate.paths` limits targets to the old input directory (by design) |
| `protoc-gen-X: executable not found` | `name: X` runs `protoc-gen-X` from `PATH`; install it or use `path:` / `remote:` |
| Only the `// protoc v…-easyp` header line changed after migration | Expected: the header carries the EasyP version |
| Conflicting descriptor in a combined export | Use `--descriptor_set_out_dir` (one set per project/module) |

## Lint and breaking

| Symptom | Fix |
|---|---|
| `multiple independent policies found` | Choose one with `easyp --cfg path/easyp.yaml lint` |
| More lint findings after switching to `default: STANDARD` | v0 `use: [DEFAULT]` meant only 8 rules; STANDARD is 32. Disable rules or keep the migrated lists. |
| `symbol "…" already defined at …` in breaking | Module root `.` contains duplicate example/test protos. Scope with `--path proto`, or narrow `roots`. |
| Breaking: ref not found (exit 2) | Fetch history: `git fetch origin main`, `fetch-depth: 0` in CI. Check the branch name: `master` vs `main`. |
| `compile baseline descriptors: … Resolve x/y.proto: … no such file` | The baseline revision has v0 config and dependencies, and v1 cannot resolve them. Expected until the migration is merged; run v0 `easyp breaking` for that PR. |
| `module X has conflicting requirements A and B` | Two requirements pin X differently, often your pin vs a dependency's fixed BSR snapshot. Pin X to the commit the dependency requires, or drop the version. |
| Many new FILE findings after migration | The v1 descriptor checker is stricter (file moves, `go_package`); choose a looser profile if appropriate |

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Lint/breaking findings; config, dependency, usage or migration errors; `validate-config` invalid; `ls-files` collected errors |
| 2 | lint/breaking: missing import file; breaking: ref or repository not found |

## Debugging techniques

```bash
easyp --debug lint                               # resolution and policy details
easyp --format json lint | jq -s 'group_by(.rule) | map({rule: .[0].rule, n: length})'   # JSONL → summary
easyp ls-files --format text                     # what EasyP sees: import_path, source, abs path
easyp ls-files --include-imports=false           # local files only
easyp validate-config --config easyp.gen.yaml    # one file
easyp schema-gen --out-dir /tmp/easyp-schemas    # current JSON Schemas
```

## Environment variables

| Variable | Meaning |
|---|---|
| `EASYP_CFG` | Same as `--cfg` |
| `EASYP_DEBUG` | Same as `--debug` |
| `EASYP_FORMAT` | Same as `--format` |
| `EASYP_INIT_DIR` | Default for `init --dir` |
| `EASYPPATH` | Cache root (default `~/.easyp`; v1 data under `~/.easyp/v1`) |

`EASYP_ROOT_GENERATE_PATH` was removed in v1.
