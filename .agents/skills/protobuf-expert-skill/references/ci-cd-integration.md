# EasyP CI/CD Integration

## GitHub Actions (Official)

EasyP provides official GitHub Actions at [easyp-tech/actions](https://github.com/easyp-tech/actions).

### Lint on push and PR

```yaml
name: easyp-lint
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: easyp-tech/actions/lint@v1
        with:
          version: v0.12.2
```

### Breaking change detection on PR

```yaml
name: easyp-breaking
on: [pull_request]

jobs:
  breaking:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0    # Required — needs full git history
      - uses: easyp-tech/actions/breaking@v1
        with:
          version: v0.12.2
          against: origin/main
```

### Combined workflow (lint + breaking)

```yaml
name: easyp
on:
  push:
    branches: [main, master]
  pull_request:

permissions:
  contents: read

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: easyp-tech/actions/lint@v1
        with:
          version: v0.12.2

  breaking:
    name: Breaking Changes
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: easyp-tech/actions/breaking@v1
        with:
          version: v0.12.2
          against: origin/main
```

Features of official actions:
- Pre-built Docker images (fast, no compilation)
- GitHub Annotations (errors appear on PR diff lines)
- Version pinning via `version` input

## GitLab CI

```yaml
stages:
  - proto

easyp-lint:
  stage: proto
  image: ghcr.io/easyp-tech/easyp:v0.12.2
  script:
    - easyp lint
  rules:
    - changes:
        - "**/*.proto"
        - easyp.yaml

easyp-breaking:
  stage: proto
  image: ghcr.io/easyp-tech/easyp:v0.12.2
  script:
    - easyp breaking --against origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME
  variables:
    GIT_DEPTH: 0
  rules:
    - if: $CI_MERGE_REQUEST_IID
      changes:
        - "**/*.proto"
        - easyp.yaml
```

## Makefile

```makefile
EASYP_VERSION ?= latest

.PHONY: proto-lint proto-breaking proto-generate proto-download proto-validate

proto-lint:
	easyp lint

proto-breaking:
	easyp breaking --against main

proto-generate:
	easyp generate

proto-download:
	easyp mod download

proto-validate:
	easyp validate-config

proto-all: proto-download proto-lint proto-generate
```

## Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Lint only changed proto files
if git diff --cached --name-only | grep -q '\.proto$'; then
  easyp lint
  if [ $? -ne 0 ]; then
    echo "Proto lint failed. Fix issues before committing."
    exit 1
  fi
fi
```

## Docker-based CI (generic)

For any CI system that supports Docker:

```bash
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  ghcr.io/easyp-tech/easyp:v0.12.2 \
  lint
```

## Important CI Notes

- **Breaking checks need full git history** — use `fetch-depth: 0` or `GIT_DEPTH: 0`
- **Pin the EasyP version** — avoid `latest` in CI for reproducibility
- **Run `validate-config` first** — catches config issues before lint/generate
- **Cache dependencies** — `easyp mod download` results are cached in `~/.cache/easyp`
