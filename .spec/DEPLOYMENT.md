<!-- generated: 2026-07-27, template: deployment.md -->
# Deployment

## Overview

EasyP is a Go command-line toolkit distributed through release automation and
documented installation channels, rather than deployed as a long-running web
service.

```text
pull request / push to main
  → GitHub Actions tests
  → release tag vX.Y.Z
  → GitHub Actions release + documentation build
  → published release artifacts and container images
```

The release workflow accepts stable semantic-version tags in the form
`v<major>.<minor>.<patch>` only. Pre-release suffixes are explicitly rejected.

## Environments

| Environment | URL / Host | Purpose | Branch / Tag | Auto-deploy |
|---|---|---|---|---|
| Local CLI | N/A | Build, test, and install the command-line tool | Local checkout | N/A |
| Staging | N/A | No staging deployment is configured | N/A | N/A |
| Production service | N/A | No long-running production service is configured | N/A | N/A |
| Release distribution | GitHub Releases, Homebrew, Docker, npm | Install EasyP artifacts | `vX.Y.Z` tag | Yes |

The repository README documents Homebrew installation, `go install`, Docker
image use, npm installation, and binary installation from GitHub Releases.
It does not define a service host, a staging URL, or runtime environment
configuration.

## Docker

The checked-in `Taskfile.yml` includes container targets for `easyp/base` and
`easyp/lint`:

```sh
task docker_base
task docker_lint
task docker_push
task docker
```

`task docker` obtains `GIT_TAG` from the latest Git tag, builds both images,
tags them as `latest` and with that tag, then pushes them.

The Taskfile expects Dockerfiles at `Docker/base/Dockerfile` and
`Docker/lint/Dockerfile`. Those paths are not present in this checkout, so
their base images, build stages, build arguments beyond
`EASYP_BASE_VERSION=latest`, exposed ports, entrypoints, and image-size
optimizations cannot be documented from the allowed source files.

No Docker Compose configuration is present. Service definitions, volume
mounts, and network configuration are therefore N/A.

## CI/CD Pipeline

### `.github/workflows/tests.yml`

- **Trigger:** pushes to `main` and pull requests targeting any branch.
- **Runner:** `ubuntu-latest`.
- **Steps:** check out the repository; install Go 1.24; install Task 3.x;
  run `task init`; run `task test`.
- **Secrets required:** `GITHUB_TOKEN`, supplied to the Task setup action.

`task init` installs the local Go development tools and resolves Go
dependencies. `task test` runs the test suite through `gotestsum` with race
detection and writes `coverage.out`.

### `.github/workflows/release.yml`

- **Trigger:** push of a stable `vX.Y.Z` tag.
- **Runner:** `ubuntu-latest` with Go 1.24.
- **Permissions:** write access to repository contents and packages.
- **Steps:** full-history checkout; QEMU setup; Docker Buildx setup; login to
  GitHub Container Registry; login to Docker Hub; install Go; validate the
  tag; run GoReleaser with `release --clean --timeout=90m`.
- **Secrets required:** `GITHUB_TOKEN`, `DOCKERHUB_USERNAME`,
  `DOCKERHUB_TOKEN`, and `GORELEASER_AUTH_TOKEN`.
- **Variable used:** `DOCKERHUB_IMAGE`, with `easyp/easyp` as the workflow
  fallback value.

The workflow's QEMU and Buildx setup supports the GoReleaser release process;
the exact artifact list and image platforms are not specified by the allowed
sources.

### `.github/workflows/docs.yml`

- **Trigger:** push of a stable `vX.Y.Z` tag.
- **Runner:** `ubuntu-latest`.
- **Steps:** full-history checkout; install Node.js 20; install dependencies
  in `docs`; build VitePress; verify English and Russian search indexes;
  detect generated documentation changes; commit and push changed `docs/dist/`
  files to the repository default branch.
- **Secrets required:** `GITHUB_TOKEN`.

The workflow reports that an external service pulls the changes for
`easyp.tech`; that external deployment is not configured in this repository.

### `.github/workflows/relator.yml`

This workflow is not a deployment pipeline. It sends Telegram notifications
when issues or pull requests are opened or reopened. It requires
`TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`, and `GITHUB_TOKEN`.

## Rollout Strategy

N/A. EasyP has no configured application-service rollout, Kubernetes
deployment, blue-green deployment, canary deployment, or zero-downtime
strategy in the allowed sources. Distribution occurs by publishing a stable
release tag through the release workflow.

## Health Checks

N/A. EasyP is a CLI, and no HTTP health, readiness, liveness, or startup probe
is defined in the allowed sources.

## Rollback Procedure

No automated rollback procedure is defined. For a faulty published CLI
release, the repository documentation and workflow sources do not specify an
artifact withdrawal or rollback command. The actionable repository-level
procedure is to correct the issue and publish a new stable `vX.Y.Z` tag through
the release workflow.

## Secrets Management

| Secret / Variable | Purpose | Where Set |
|---|---|---|
| `GITHUB_TOKEN` | GitHub API, package-registry, checkout, and documentation workflow authentication | GitHub Actions |
| `GORELEASER_AUTH_TOKEN` | GoReleaser release authentication | GitHub Actions secret |
| `DOCKERHUB_USERNAME` | Docker Hub login username | GitHub Actions secret |
| `DOCKERHUB_TOKEN` | Docker Hub login token | GitHub Actions secret |
| `DOCKERHUB_IMAGE` | Docker Hub image name override | GitHub Actions variable |
| `TELEGRAM_BOT_TOKEN` | Relator notification authentication | GitHub Actions secret |
| `TELEGRAM_CHAT_ID` | Relator notification destination | GitHub Actions variable |

Secret rotation, vault integration, and runtime secret injection are not
documented in the allowed sources.

## Infrastructure Requirements

N/A. No infrastructure resource requirements, replicas, storage allocations,
or orchestration manifests are configured for EasyP in the allowed sources.

## Monitoring and Alerts

N/A for runtime monitoring because EasyP has no deployed service in this
repository. GitHub Actions workflow results are the available build and release
execution record. The Relator workflow provides issue and pull-request
notifications, but no monitoring dashboard or alert policy is configured.
