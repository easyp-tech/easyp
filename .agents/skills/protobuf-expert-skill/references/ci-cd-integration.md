# EasyP v1 CI/CD Integration

Principles for v1 pipelines:

- **Pin the exact EasyP version.** v1 ships as `v1.0.0-nightly.YYYYMMDD.N`; `latest` tags and `@latest` still mean v0.17.0.
- **Use `--frozen`.** It is never enabled automatically, not even when `CI=true`. With it, CI uses the committed `protobuf.mod` + `protobuf.lock` exactly and fails if they are stale, instead of re-resolving.
- **Commit `protobuf.mod` and `protobuf.lock`.** A module without dependencies needs an empty lock (`modules: []`).
- **Breaking needs history.** Use `fetch-depth: 0`.
- **Cache `~/.easyp/v1`**, keyed on `protobuf.lock`.

```bash
# the canonical v1 CI sequence
easyp --frozen mod download
easyp --frozen lint
easyp --frozen breaking --against origin/main
easyp --frozen generate && git diff --exit-code     # generated code is up to date
```

## GitHub Actions — official actions

[easyp-tech/actions](https://github.com/easyp-tech/actions) run the Docker image `ghcr.io/easyp-tech/easyp:<version>`. They take these inputs:
- `version`: the image tag;
- `directory`: passed as `--path`;
- `config`: passed as `--cfg`;
- `against`: breaking only, default `origin/main`.

Each action runs `mod download`, then `lint` or `breaking`. **Always set `version`.** The default is `latest`, which is v0.

```yaml
name: easyp
on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

env:
  EASYP_VERSION: v1.0.0-nightly.20261008.1

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: easyp-tech/actions/lint@v1
        with:
          version: ${{ env.EASYP_VERSION }}
          directory: proto

  breaking:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: easyp-tech/actions/breaking@v1
        with:
          version: ${{ env.EASYP_VERSION }}
          directory: proto
          against: origin/${{ github.base_ref }}
```

## GitHub Actions — direct binary (with generation check and cache)

```yaml
name: proto
on: [push, pull_request]

env:
  EASYP_VERSION: v1.0.0-nightly.20261008.1

jobs:
  proto:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - uses: actions/cache@v4
        with:
          path: ~/.easyp/v1
          key: easyp-${{ env.EASYP_VERSION }}-${{ hashFiles('**/protobuf.lock') }}
      - name: Install EasyP
        run: go install "github.com/easyp-tech/easyp/cmd/easyp@${EASYP_VERSION}"
      - run: easyp validate-config --format text
      - run: easyp --frozen mod download
      - run: easyp --frozen lint
      - name: Breaking
        if: github.event_name == 'pull_request'
        run: easyp --frozen breaking --against "origin/${{ github.base_ref }}"
      - name: Generated code is up to date
        run: |
          easyp --frozen generate        # add --all for multi-project repos
          git diff --exit-code
```

Local plugins (`name: go`, `name: go-grpc`) must be on `PATH` in the job. Install them with pinned versions (`go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11`), or use `remote:` plugins.

## GitLab CI

```yaml
variables:
  EASYP_IMAGE: ghcr.io/easyp-tech/easyp:v1.0.0-nightly.20261008.1
  EASYPPATH: $CI_PROJECT_DIR/.easyp

.easyp:
  image:
    name: $EASYP_IMAGE
    entrypoint: [""]
  cache:
    key:
      files: [protobuf.lock]
    paths: [.easyp/v1]
  before_script:
    - easyp --frozen mod download

proto-lint:
  extends: .easyp
  script:
    - easyp --frozen lint

proto-breaking:
  extends: .easyp
  variables:
    GIT_DEPTH: 0
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
  script:
    - git fetch origin "$CI_MERGE_REQUEST_TARGET_BRANCH_NAME"
    - easyp --frozen breaking --against "origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME"
```

## Makefile / Taskfile

```makefile
EASYP ?= easyp

.PHONY: proto-deps proto-lint proto-breaking proto-gen proto-check
proto-deps:      ; $(EASYP) mod tidy                      # after editing protobuf.mod
proto-lint:      ; $(EASYP) lint
proto-breaking:  ; $(EASYP) breaking --against origin/main
proto-gen:       ; $(EASYP) generate
proto-check:     ; $(EASYP) --frozen lint && $(EASYP) --frozen generate && git diff --exit-code
```

```yaml
# Taskfile.yml
version: '3'
tasks:
  proto:lint:     { cmds: [easyp lint] }
  proto:breaking: { cmds: [easyp breaking --against origin/main] }
  proto:gen:      { cmds: [easyp generate] }
  proto:check:
    cmds:
      - easyp --frozen lint
      - easyp --frozen generate
      - git diff --exit-code
```

## Pre-commit hook

```sh
#!/bin/sh
# .git/hooks/pre-commit — lint when protos or EasyP files are staged
if git diff --cached --name-only | grep -qE '\.proto$|(^|/)(easyp(\.gen)?\.yaml|protobuf\.(mod|lock))$'; then
  easyp lint || exit 1
fi
```

## Docker (any CI)

```bash
docker run --rm -v "$PWD:/workspace" -w /workspace \
  -e EASYPPATH=/workspace/.easyp \
  ghcr.io/easyp-tech/easyp:v1.0.0-nightly.20261008.1 --frozen lint
```

## Migrating a v0 pipeline

| v0 | v1 |
|---|---|
| `version: v0.x` / image `:v0.x` / `@latest` | the exact nightly tag |
| `easyp mod download` | `easyp --frozen mod download` |
| `easyp generate -p proto -r .` | `easyp --frozen generate` (`--project` / `--all`) |
| cache path `~/.easyp`, key `easyp.lock` | path `~/.easyp/v1`, key `protobuf.lock` |
| `breaking --against` required | optional when `breaking.baseline` is set |

Full procedure: [migration-v0-to-v1.md](./migration-v0-to-v1.md#step-9--ci-scripts-and-cleanup).
