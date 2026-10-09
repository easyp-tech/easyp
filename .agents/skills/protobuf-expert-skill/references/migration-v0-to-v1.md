# Migrating EasyP v0 → v1

EasyP v1 replaces the single `easyp.yaml` with four files and refuses to run on v0 configuration. The supported path is the built-in `easyp migrate` command. This guide wraps it in a procedure an agent can follow end to end: inventory → baseline → preview → fix blockers → apply → verify → update the surroundings.

Verified against `v1.0.0-nightly.20261008.1` (v0 baseline: `v0.17.0`).

## Contents

- [What changes](#what-changes)
- [Step 1 — Inventory](#step-1--inventory)
- [Step 2 — Baseline with v0](#step-2--baseline-with-v0)
- [Step 3 — Module identity](#step-3--module-identity)
- [Step 4 — Preview](#step-4--preview)
- [Step 5 — Fix blockers](#step-5--fix-blockers)
- [Step 6 — Apply](#step-6--apply)
- [Step 7 — Verify on v1](#step-7--verify-on-v1)
- [Step 8 — Behaviour changes to report](#step-8--behaviour-changes-to-report)
- [Step 9 — CI, scripts and cleanup](#step-9--ci-scripts-and-cleanup)
- [Rollback](#rollback)
- [Key mapping v0 → v1](#key-mapping-v0--v1)
- [Worked example](#worked-example)

## What changes

| v0 | v1 |
|---|---|
| One `easyp.yaml` (lint + deps + generate + breaking) | `easyp.yaml` (lint/breaking policy), `easyp.gen.yaml` (generation), `protobuf.mod` (identity, roots, dependencies), `protobuf.lock` (resolved pins) |
| `easyp.lock`, plain text `module version h1:hash` | `protobuf.lock`, YAML: `source`, `version`, `commit`, `hash` |
| Cache `~/.easyp/{cache,mod}` | Cache `~/.easyp/v1/git/...` (`$EASYPPATH/v1`) |
| `easyp generate -p proto -r .` | `easyp generate` (nearest `easyp.gen.yaml`), `--project DIR`, `--all` |
| `remote: host/org/plugin:v1.2.3` | `remote: host/org/plugin` + `version: v1.2.3` |
| `deps: [repo@ref]` | `require repo v1.2.3` in `protobuf.mod` |

A v1 binary run against v0 configuration stops with:

```
legacy EasyP configuration detected; run easyp migrate --module <identity> to preview conversion to v1
```

## Step 1 — Inventory

1. **Find every EasyP config in the repository, not just the root one.**
   - `validate-config` scans every directory, hidden ones included.
   - `lint` reads descendant policies in non-hidden directories.
   - So a nested v0 `easyp.yaml` breaks the v1 run: an `examples/` project fails `lint` and `validate-config`, and a skill asset under `.agents/` fails `validate-config`.

   Search both the working tree **and** the tracked files. A dirty tree can hide files that a clean CI checkout still has.

   ```bash
   find . \( -name easyp.yaml -o -name easyp.yml -o -name protobuf.mod -o -name easyp.lock \) \
     -not -path '*/node_modules/*' -not -path '*/easyp_vendor/*' -not -path './.git/*'
   git ls-files | grep -E '(^|/)(easyp\.ya?ml|protobuf\.mod|easyp\.lock)$'
   ```

2. **Classify each directory:**

   | Signal | Version |
   |---|---|
   | `easyp.yaml` has top-level `lint:`, `deps:`, `generate:` or `breaking_check:`, and no `version: v1` | **v0** (`version: v1alpha` is still v0) |
   | `protobuf.mod` contains a `direct (` block, or `name@version` entries | **v1.0-rc pilot** (migrate accepts it as legacy input) |
   | `easyp.yaml` has `version: v1` / `linters:`; `easyp.gen.yaml` exists; `protobuf.mod` starts with `module` | **v1**, nothing to migrate |
   | `easyp.yaml` that is only an example or a template (docs, skill assets) | Rename it (`easyp-example.yaml`) or migrate it; do not leave a v0 file named `easyp.yaml` |

3. **Check the binary:** `easyp --version`. v1 is distributed only as nightly tags (`v1.0.0-nightly.YYYYMMDD.N`). `brew` and `go install …@latest` still install v0.17.0. Install v1 side by side; see [installation.md](./installation.md). Keep v0 available for Step 2.

## Step 2 — Baseline with v0

Before touching configs, record what v0 produces so that the v1 result can be compared with it.

Work on a dedicated branch. If `git status` shows unrelated changes, ask the user whether to commit them first, or do the migration in a separate `git worktree`. Do not stash blindly: stashing can bring back deleted files, nested v0 configs among them.

Run lint, generate and breaking **exactly the way the project already does**: copy the flags from its Taskfile, Makefile or CI (`-p api`, `-p proto -r .` or no flags at all). Write logs inside the repo's ignored scratch area, or in the agent's scratch directory, not in `/tmp`.

```bash
EASYP_V0="go run github.com/easyp-tech/easyp/cmd/easyp@v0.17.0"   # or an installed v0 binary

$EASYP_V0 --format json lint > .easyp-v0-lint.json; echo "lint exit=$?"     # + the project's -p/-r flags
$EASYP_V0 generate                                                           # + the project's flags
$EASYP_V0 breaking --against <base-branch>; echo "breaking exit=$?"          # v1 cannot do this yet (Step 7)
git status --short
```

- If v0 generation already shows a diff, commit it as a separate baseline commit, or at least note it. Typical causes: an unpinned remote plugin that now resolves to a newer `latest`, or newer local plugins. Otherwise the diff gets blamed on the migration.
- The `breaking` result of this step is the last compatibility check of this branch against the v0 base branch (Step 7).

## Step 3 — Module identity

`easyp migrate` needs `--module <identity>`: the canonical name of the local protobuf module, written to `module …` in `protobuf.mod`.

- Usually the repository path: `github.com/acme/contracts`. For a module nested in a subdirectory, append the directory: `github.com/acme/contracts/proto/billing`.
- It is **not** the Go `go_package` and **not** the protobuf package.
- Derive a candidate from `git remote get-url origin`, then **confirm it with the user**. Other repositories will `require` this identity, so it is hard to change later.
- For v2+ modules the identity must end with `/vN` (Go semantic import versioning).

## Step 4 — Preview

`migrate` writes nothing without `--write`. Run the preview for every directory found in Step 1:

```bash
easyp migrate --dir . --module github.com/acme/contracts
```

- If `easyp.yaml` has `deps`, the preview reports `protobuf.lock cannot be prepared without --resolve-lock`. Run the preview again with `--resolve-lock`. It fetches the dependencies, verifies the historical `easyp.lock` pins and hashes, and fills the cache (network required):

  ```bash
  easyp migrate --dir . --module github.com/acme/contracts --resolve-lock
  ```

- Run on a terminal without `--module`, migrate starts an interactive wizard. Agents should always pass `--module`, which keeps it non-interactive, or pass `--interactive=false`.
- `--frozen` is rejected by migrate.

**Read the preview.** It prints every candidate file and its warnings. Usual warnings:

| Warning | Meaning |
|---|---|
| `legacy version "v1alpha" … is omitted` | Harmless. |
| `Legacy directory selection is preserved through literal paths selectors` | `generate.inputs[].directory` became `generate.paths`. Files outside these paths are not generation targets even when they share a package. |
| `easyp.lock is retained byte-for-byte` | The old lock stays; delete it later (Step 9). |
| `Relative paths and variable placeholders are preserved; no plugin is executed` | `${VAR}` in `out`/`opts` is kept; make sure CI still sets the variable. |
| `using BSR compatibility snapshot` | A dependency's `buf.yaml` declares BSR deps; EasyP maps them to fixed Git snapshots. See [troubleshooting.md](./troubleshooting.md). |

Show the user the candidate `easyp.yaml`, `easyp.gen.yaml` and `protobuf.mod`, and confirm before applying.

## Step 5 — Fix blockers

migrate refuses to convert what it cannot prove equivalent. The error is printed after `Build: …`. Fix it in the **v0** file, then preview again.

| Error (substring) | Cause | Fix |
|---|---|---|
| `PACKAGE_NO_IMPORT_CYCLE is not implemented` | The rule was removed in v1 | Delete it from `lint.use` / `lint.except`. Tell the user import-cycle checking is gone. |
| `unsupported legacy breaking.use "WIRE_JSON"` (or `WIRE`, `PACKAGE`) | Only FILE converts automatically | Remove `breaking.use`, migrate, then set `breaking.categories: [WIRE_JSON]` in the new `easyp.yaml`. |
| `generate.plugins[N].remote needs a pinned semantic version` | Remote plugin without `:vX.Y.Z`. migrate reports only the **first** one, so check every `remote:` entry. | Pin **all** of them in the v0 file (`remote: plugins.beta.easyp.tech/grpc/go:v1.6.2`), then rerun v0 `generate`. The pins are right when the Step 2 baseline is reproduced with no diff. See [choosing remote plugin pins](#choosing-remote-plugin-pins). |
| `unsupported ref "<name>": use a semantic version or full Git commit` | `deps` entry pinned to a branch, short hash or non-semver tag (`@common-protos-1_3_1`) | Replace it with the full commit: `git ls-remote https://<repo> <name>` (for an annotated tag, use the `^{}` line). Or remove the pin and add it after migration with `easyp get repo@<name>`, which resolves named tags. |
| `legacy easyp.lock is missing required dependency X; recover its historical pin` | `deps` lists X but `easyp.lock` has no entry for it | Restore the entry from git history (`git log -p -- easyp.lock`), or run v0 `easyp mod download` to record it. migrate never silently picks HEAD for a dependency it expected to be locked. |
| `unsupported legacy config field "excludes"` (inside `buf.yaml`) | A dependency's `buf.yaml` uses a field the legacy-hash verifier does not read (seen with `bufbuild/protovalidate`) | Use the [re-resolution fallback](#fallback-re-resolve-dependencies) below. |
| `incompatible requirement X@A: historical commit is B` | Another dependency (through its BSR snapshot) needs X at a commit different from your `easyp.lock` pin | Use the [re-resolution fallback](#fallback-re-resolve-dependencies). |
| *(fallback)* `module X has conflicting requirements A and B` from `mod tidy` | Your pin for X disagrees with the commit another dependency requires (typically a fixed BSR snapshot such as googleapis `03a91044…` required through grpc-gateway) | Pin X to the commit the other dependency requires: the snapshot is fixed and cannot move. Or drop the version from X's `require` line. Then compare the generated code (Step 7) and tell the user that X moved. |
| `legacy EasyP configuration detected` from a **nested** path | Another v0 project inside the tree | Migrate that directory too (`--dir path --module <its identity>`), or rename it if it is only an example. |
| `Resolve <dir>: no such file or directory` | `generate.inputs[].directory.path` is resolved relative to its `root` | Fix the v0 input. It is already broken in v0. |
| *(after apply)* `malformed lint directive "easyp:off"` from `easyp lint` | v1 accepts only `easyp:disable RULE` / `easyp:enable RULE`; any other `easyp:` comment fails the run | `grep -rn 'easyp:' --include='*.proto' .` and rewrite each one as a `// easyp:disable RULE[, RULE]` / `// easyp:enable RULE` pair that names the rules. `buf:lint:ignore` and `nolint:` still work. |

### Fallback: re-resolve dependencies

When the historical lock cannot be verified, migrate the configuration without dependencies and let v1 resolve them again. **This can change dependency versions.** Say so to the user, and rely on the Step 7 comparison.

```bash
cp easyp.yaml easyp.yaml.v0.orig           # the untouched original: migrate's .v0.bak will hold the edited copy
git rm --cached -q easyp.lock && mv easyp.lock easyp.lock.v0   # old pins, for reference below
# in easyp.yaml: set `deps: []` (the original list is in easyp.yaml.v0.orig)
easyp migrate --dir . --module github.com/acme/contracts --write
```

In this path, roll back from `easyp.yaml.v0.orig` and `easyp.lock.v0`, not from `easyp.yaml.v0.bak`.

Then add **all** requirements at once to `protobuf.mod` and run `mod tidy`. Adding them one by one with `easyp get` fails: each call checks every project import, and the other dependencies are still missing.

```text
module github.com/acme/contracts

roots (
	.
)

require (
	github.com/googleapis/googleapis
	github.com/grpc-ecosystem/grpc-gateway
	github.com/bufbuild/protovalidate
)
```

```bash
easyp mod tidy          # resolves and writes protobuf.lock; dependencies you don't import directly get `// indirect`
```

To keep old pins (recommended, as it keeps the generated code stable), write `require <repo> <full-commit>`, or a semver tag. The commit is the last segment of the pseudo-version in `easyp.lock.v0`, e.g. `v0.0.0-20260713124408-3807e3d1c38b…` → `3807e3d1c38b…`. If `mod tidy` then reports `conflicting requirements`, follow the row in the table above.

### Choosing remote plugin pins

There is no CLI command that lists registry versions. The registry speaks gRPC only, and the newest upstream tag is not always published there.

1. **Find candidates:**
   - generated-file headers (`// - protoc-gen-go-grpc v1.6.2`, `// 	protoc-gen-go v1.36.11`). Gateway, openapiv2 and validate plugins write no version;
   - `$EASYP_V0 --debug generate`, which logs the resolved `version=` (`latest` means unpinned);
   - upstream tags: `git ls-remote --tags https://github.com/<org>/<plugin>`.
2. **Decide with the user** between the version the *committed* code was generated with, and the version v0 resolves today (`latest`). Prefer the one that reproduces the committed code. If that is impossible, use the Step 2 baseline and note the bump.
3. **Verify** each pin with v0 `generate`. `NotFound` means the registry lacks that version; try the previous tag.

## Step 6 — Apply

```bash
easyp migrate --dir . --module github.com/acme/contracts --resolve-lock --write
```

- Omit `--resolve-lock` when there are no dependencies.
- Written: `easyp.yaml` (v1 policy), `easyp.gen.yaml`, `protobuf.mod`, `protobuf.lock`. The latter is `modules: []` when there are no dependencies, and it is still required.
- Backups: `easyp.yaml.v0.bak`, plus `protobuf.mod.v0.bak` for rc input. `easyp.lock` stays unchanged.
- Existing native v1 files are never overwritten with different content.
- Repeat for every directory from Step 1.

## Step 7 — Verify on v1

```bash
easyp validate-config --format text      # VALID: true, across the whole tree
easyp --frozen mod download              # the lock is complete and verifiable
easyp --format json lint > /tmp/easyp-v1-lint.json; echo "lint exit=$?"
easyp generate                           # or --project DIR / --all for multi-project repos
git status --short && git diff --stat    # compare generated code with the Step 2 baseline
easyp breaking --against HEAD            # v1 vs v1 sanity check once the migration is committed
```

Expected results:

- **Lint:** same issue set as the baseline. migrate expands v0 groups into exact rule lists, so the result is identical unless you change `linters.default` later.
- **Generated code:** identical, **except** the descriptor header line `// 	protoc v0.14.1-v0.17.0-…-easyp` → `…-v1.0.0-…`. Any other diff needs an explanation: plugin version, dependency version from the fallback, or `generate.paths` selection.
- **Breaking against the v0 base branch works only for projects without dependencies.** Each revision is compiled with its own files. A v0 revision has no `protobuf.mod`/`protobuf.lock`, so v1 cannot resolve its dependency imports, and the run fails with `compile baseline descriptors: … Resolve validate/validate.proto: … no such file`.
  - Use the v0 `breaking` result from Step 2 as the compatibility check of this branch.
  - Tell the user that CI's breaking job will fail on the migration PR. Make it non-blocking for that PR, or keep running v0 for it.
  - Once the migration is merged, base and head are both v1 and the check works normally.
- **Base branch name:** use the repository's real default branch (`git symbolic-ref refs/remotes/origin/HEAD`), not `main` by assumption.
- **`symbol … already defined`** across unrelated directories means the module root `.` includes example or test projects. Scope the check with `--path proto` (as the old CI did), or narrow `roots` in `protobuf.mod`.
- **`validate-config`** must be clean for the **whole tree**. Remaining v0 files are the nested projects or templates from Step 1.

## Step 8 — Behaviour changes to report

Tell the user explicitly:

- **`allow_comment_ignores` now defaults to `true`.** migrate writes the old value explicitly (`false` when v0 did not set it), so behaviour is preserved. New v1 configs get `true`.
- **Presets differ from v0 groups.** v1 `linters.default` is cumulative: `MINIMAL` ⊂ `BASIC` ⊂ `STANDARD` ⊂ `COMMENTS`; `STANDARD` = MINIMAL+BASIC+DEFAULT. A v0 `use: [DEFAULT]` alone did *not* include MINIMAL/BASIC rules, so migrate emits `default: MINIMAL` plus explicit `enable`/`disable` lists. Offer to simplify to `default: STANDARD`; it may surface new findings.
- **`PACKAGE_NO_IMPORT_CYCLE` is gone.**
- **Breaking:** omitted or empty `breaking.categories` means `FILE`. The default baseline comes from `breaking.baseline: git:<ref>`; the CLI fallback is still `master`. migrate writes `breaking: {}` when v0 had no breaking section, so offer to set `baseline: git:<default-branch>`.
- **Generation targets:** `generate.paths` limits targets to the old input directories. Dependencies are import-only unless listed in `generate.modules`.
- **Remote plugins** need `version:`; `:latest` is rejected.

## Step 9 — CI, scripts and cleanup

Search for every caller, and for docs that quote the v0 config:

```bash
grep -rnE 'easyp( |$)|easyp\.lock|EASYP_' .github .gitlab-ci.yml Makefile Taskfile*.yml scripts docs 2>/dev/null
grep -rnE '^\s*(deps|lint|generate|inputs|use|except|ignore_only|against_git_ref|enum_zero_value_suffix|service_suffix):' \
  --include='*.md' --include='*.mdx' . | grep -v node_modules
```

Agent instructions count too: `AGENTS.md`, `.github/agents`, `.github/instructions` and similar files.

| v0 usage | v1 replacement |
|---|---|
| `easyp generate -p proto -r .` / `EASYP_ROOT_GENERATE_PATH` | `easyp generate` (from the directory with `easyp.gen.yaml`), `--project DIR`, or `--all` |
| `easyp lint -p proto -r .` | Unchanged: `lint` and `breaking` keep `--path/-p` and `--root/-r` |
| `easyp mod download` in CI | `easyp --frozen mod download` |
| `easyp mod update` | `easyp mod update` (rewrites `protobuf.mod` + `protobuf.lock`) |
| cache key on `easyp.lock`, path `~/.easyp` | key on `protobuf.lock`, path `~/.easyp/v1` |
| `go install …/cmd/easyp@latest`, `brew install` | pin a nightly tag explicitly (see [ci-cd-integration.md](./ci-cd-integration.md)) |
| `easyp-tech/actions/*@v1` without `version` | pass `version: v1.0.0-nightly.YYYYMMDD.N` (the action runs the Docker image of that tag) |
| `.gitignore` ignoring `easyp.lock` patterns | commit `protobuf.mod` and `protobuf.lock` |

Cleanup, **only after the user confirms** the verification results: delete `easyp.lock` (or `easyp.lock.v0`), `*.v0.bak` and `*.v0.orig` (they stay in git history), and update README/docs that show v0 config or commands.

## Rollback

```bash
mv easyp.yaml.v0.bak easyp.yaml          # fallback path: use easyp.yaml.v0.orig and easyp.lock.v0 instead
rm -f easyp.gen.yaml protobuf.lock
[ -f protobuf.mod.v0.bak ] && mv protobuf.mod.v0.bak protobuf.mod || rm -f protobuf.mod
# or simply: git checkout -- . && git clean -fd -- easyp.gen.yaml protobuf.mod protobuf.lock '*.v0.bak'
```

## Key mapping v0 → v1

Use this table only when migrate cannot run, or to explain its output.

| v0 (`easyp.yaml`) | v1 | File |
|---|---|---|
| — | `version: v1` | `easyp.yaml`, `easyp.gen.yaml` |
| `version: v1alpha` | dropped | — |
| `lint.use: [groups/rules]` | `linters.default` (MINIMAL/BASIC/STANDARD/COMMENTS) + `linters.enable` | `easyp.yaml` |
| `lint.except` | `linters.disable` | `easyp.yaml` |
| `lint.allow_comment_ignores` | `linters.allow_comment_ignores` (default **true**) | `easyp.yaml` |
| `lint.enum_zero_value_suffix: X` | `linters-settings.ENUM_ZERO_VALUE_SUFFIX.suffix: X` | `easyp.yaml` |
| `lint.service_suffix: X` | `linters-settings.SERVICE_SUFFIX.suffix: X` | `easyp.yaml` |
| `lint.ignore: [dir]` | `issues.exclude-rules: [{path: dir}]` | `easyp.yaml` |
| `lint.ignore_only: {RULE: [dir]}` | `issues.exclude-rules: [{path: dir, linters: [RULE]}]` | `easyp.yaml` |
| `breaking.against_git_ref: main` | `breaking.baseline: git:main` | `easyp.yaml` |
| `breaking.ignore` | `breaking.ignore` | `easyp.yaml` |
| `breaking.use: [FILE]` | `breaking.categories: [FILE]` (default) | `easyp.yaml` |
| `deps: [repo]` | `require repo` | `protobuf.mod` |
| `deps: [repo@v1.2.3]` / `@<commit>` | `require repo v1.2.3` / `require repo <full-commit>` | `protobuf.mod` |
| `generate.inputs[].directory: proto` | `generate.paths: [proto]`, module root `.` | `easyp.gen.yaml`, `protobuf.mod` |
| `generate.inputs[].directory: {path: P, root: R}` | `roots R` | `protobuf.mod` |
| `generate.inputs[].git_repo: {url: repo@v}` | `require repo v` + `generate.modules: [repo]` | `protobuf.mod`, `easyp.gen.yaml` |
| `git_repo.sub_directory` / `root` | module-scoped `generate.modules[].paths` / lock `roots` | migrate verifies them |
| `generate.plugins` | top-level `plugins` | `easyp.gen.yaml` |
| `remote: host/p:v1` | `remote: host/p` + `version: v1` | `easyp.gen.yaml` |
| `generate.managed` | `generate.managed` (unchanged) | `easyp.gen.yaml` |
| `easyp.lock` | `protobuf.lock` (regenerated, never converted by hand) | — |

## Worked example

Before (v0):

```yaml
version: v1alpha
deps:
  - github.com/googleapis/googleapis
lint:
  use: [DEFAULT, COMMENTS]
  except: [COMMENT_FIELD]
  ignore: [proto/legacy]
  enum_zero_value_suffix: _UNSPECIFIED
breaking:
  against_git_ref: main
generate:
  inputs:
    - directory: {path: ., root: proto}
  plugins:
    - remote: plugins.beta.easyp.tech/protocolbuffers/go:v1.36.11
      out: gen/go
      opts: {paths: source_relative}
```

After `easyp migrate --module github.com/acme/contracts --resolve-lock --write` (rule lists shortened):

```yaml
# easyp.yaml
version: v1
linters:
  default: MINIMAL
  enable: [ENUM_VALUE_PREFIX, …, COMMENT_ENUM, COMMENT_MESSAGE, …]
  disable: [DIRECTORY_SAME_PACKAGE, …]
  allow_comment_ignores: false
linters-settings:
  ENUM_ZERO_VALUE_SUFFIX:
    suffix: _UNSPECIFIED
issues:
  exclude-rules:
    - path: proto/legacy
breaking:
  baseline: git:main
```

```yaml
# easyp.gen.yaml
version: v1
plugins:
  - remote: plugins.beta.easyp.tech/protocolbuffers/go
    version: v1.36.11
    out: gen/go
    opts:
      paths: source_relative
```

```text
# protobuf.mod
module github.com/acme/contracts

roots (
	proto
)

require github.com/googleapis/googleapis
```

`protobuf.lock` then pins googleapis to the commit and hash verified from the old `easyp.lock`.
