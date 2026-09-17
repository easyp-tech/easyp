<!-- generated: 2026-05-14, template: deployment.md -->
# EasyP Deployment

## 1. Overview

EasyP is a CLI tool distributed as pre-built binaries, Docker images, and Homebrew packages. Releases are automated via GoReleaser and GitHub Actions.

```
tag push → CI (tests) → GoReleaser → binaries + Docker + Homebrew
```

## 2. Distribution Channels

| Channel | Target | Auto-deploy |
|---------|--------|-------------|
| GitHub Releases | All platforms (binaries) | Yes (on tag) |
| Docker (ghcr.io) | `ghcr.io/easyp-tech/easyp` | Yes (on tag) |
| Homebrew | `easyp-tech/homebrew-tap` | Yes (on tag) |
| Go Install | `go install github.com/easyp-tech/easyp/cmd/easyp@latest` | Manual |

## 3. Docker

### Dockerfile Analysis

Multi-stage build defined in `Dockerfile`:

**Stage 1: Builder** (`golang:1.25-alpine`)
- Copies `go.mod`/`go.sum`, downloads modules
- Builds binary with `CGO_ENABLED=0` and `-ldflags` for version injection
- Strips debug info for smaller binary

**Stage 2: Runtime** (`alpine:3.22`)
- Installs: `ca-certificates`, `tzdata`, `git`, `bash`
- Copies binary from builder
- Entry point: `/easyp`

**Usage:**
```bash
# Lint proto files
docker run --rm -v $(pwd):/work -w /work ghcr.io/easyp-tech/easyp lint

# Generate code
docker run --rm -v $(pwd):/work -w /work ghcr.io/easyp-tech/easyp generate

# Breaking check
docker run --rm -v $(pwd):/work -w /work ghcr.io/easyp-tech/easyp breaking --against main
```

### Multi-arch Support

| Architecture | Platform |
|-------------|----------|
| `linux/amd64` | x86_64 |
| `linux/arm64` | ARM64 |

## 4. CI/CD Pipeline

### `.github/workflows/tests.yml`

| Field | Value |
|-------|-------|
| **Trigger** | Push to `main`, all pull requests |
| **Steps** | `task init` → `task test` |
| **Purpose** | Run full test suite with race detection |

### `.github/workflows/release.yml`

| Field | Value |
|-------|-------|
| **Trigger** | Tag push (`v*`) |
| **Steps** | GoReleaser: build binaries → Docker images → Homebrew formula |
| **Secrets** | `GITHUB_TOKEN` (auto), Docker registry credentials |

### `.github/workflows/docs.yml`

| Field | Value |
|-------|-------|
| **Trigger** | Push to `main` (docs/ changed) |
| **Steps** | Build docs site → deploy |

## 5. GoReleaser Configuration

Defined in `.goreleaser.yaml` (GoReleaser v2):

### Build Targets

| OS | Architecture |
|----|-------------|
| `darwin` | `amd64`, `arm64` |
| `linux` | `amd64`, `arm64`, `arm` (v6, v7) |
| `windows` | `amd64`, `arm64` |

### Artifacts

- **Binaries** — Compressed archives for each platform
- **Docker images** — Multi-arch manifest (`linux/amd64` + `linux/arm64`)
- **Homebrew tap** — Formula pushed to `easyp-tech/homebrew-tap`

### Build Flags

```yaml
ldflags:
  - -s -w
  - -X main.version={{.Version}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.Date}}
```

## 6. Release Process

1. Create a Git tag: `git tag v1.2.3`
2. Push the tag: `git push origin v1.2.3`
3. GitHub Actions triggers `release.yml`
4. GoReleaser:
   - Builds binaries for all platforms
   - Builds and pushes Docker images to `ghcr.io`
   - Updates Homebrew formula in `easyp-tech/homebrew-tap`
   - Creates GitHub Release with changelog

### Rollback

1. Delete the GitHub Release
2. Delete the Git tag: `git push --delete origin v1.2.3`
3. Push a new tag with the fix

## 7. Local Development

```bash
# Install dev tools
task init

# Build and run
task build
./easyp lint

# Run directly
go run ./cmd/easyp lint

# Run tests
task test

# Full quality check
task quality
```
