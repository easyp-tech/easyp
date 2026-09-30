<!-- generated: 2026-09-30, template: deployment.md -->
# Deployment

## Overview

EasyP is a Go command-line toolkit distributed through release automation and
documented installation channels, rather than deployed as a long-running web
service.

~~~text
pull request / push to main
  → GitHub Actions tests
  → release tag vX.Y.Z
  → GitHub Actions release
  → published release artifacts and container images
~~~

The release workflow accepts stable semantic-version tags in the form
<code>v&lt;major&gt;.&lt;minor&gt;.&lt;patch&gt;</code> only. Pre-release suffixes are explicitly rejected.

## Environments

| Environment | URL / Host | Purpose | Branch / Tag | Auto-deploy |
|---|---|---|---|---|
| Local CLI | N/A | Build, test, and install the command-line tool | Local checkout | N/A |
| Staging | N/A | No staging deployment is configured | N/A | N/A |
| Production service | N/A | No long-running production service is configured | N/A | N/A |
| Release distribution | GitHub Releases, Homebrew, Docker, npm | Install EasyP artifacts | <code>vX.Y.Z</code> tag | Yes |

The repository README documents Homebrew installation, <code>go install</code>, Docker
image use, npm installation, and binary installation from GitHub Releases.
It does not define a service host, a staging URL, or runtime environment
configuration.

## Docker

The checked-in <code>Taskfile.yml</code> includes container targets for <code>easyp/base</code> and
<code>easyp/lint</code>:

~~~sh
task docker_base
task docker_lint
task docker_push
task docker
~~~

<code>task docker</code> obtains <code>GIT_TAG</code> from the latest Git tag, builds both images,
tags them as <code>latest</code> and with that tag, then pushes them.

The helper targets still expect missing <code>Docker/base/Dockerfile</code> and
<code>Docker/lint/Dockerfile</code>; their preconditions fail in this checkout. This is
an unresolved Taskfile reference, not the production release build path.

The existing root <code>Dockerfile</code> builds with <code>golang:1.26-alpine</code>,
<code>CGO_ENABLED=0</code>, <code>-trimpath</code>, and target OS/architecture build arguments. It
packages <code>/easyp</code> into <code>alpine:3.22</code> with CA certificates, timezone data, Git
and Bash, and uses <code>/easyp</code> as its entrypoint. It declares no listening port.

No Docker Compose configuration is present. Service definitions, volume
mounts, and network configuration are therefore N/A.

## CI/CD Pipeline

### <code>.github/workflows/tests.yml</code>

- **Trigger:** pushes to <code>main</code> and pull requests targeting any branch.
- **Runner:** <code>ubuntu-latest</code>.
- **Steps:** check out the repository; install Go from <code>go.mod</code> (1.26.6); install Task 3.x;
  run <code>task init</code>, <code>task test</code>, then <code>task proto:check</code>.
- **Secrets required:** <code>GITHUB_TOKEN</code>, supplied to the Task setup action.

<code>task init</code> installs the local Go development tools and resolves Go
dependencies. <code>task test</code> runs the test suite through <code>gotestsum</code> with race
detection and writes <code>coverage.out</code>.

### <code>.github/workflows/release.yml</code>

- **Trigger:** push of a stable <code>vX.Y.Z</code> tag.
- **Runner:** <code>ubuntu-latest</code>, with Go selected from <code>go.mod</code> (1.26.6).
- **Permissions:** write access to repository contents and packages.
- **Steps:** full-history checkout; QEMU setup; Docker Buildx setup; login to
  GitHub Container Registry; login to Docker Hub; install Go; validate the
  tag; run GoReleaser with <code>release --clean --timeout=90m</code>.
- **Secrets required:** <code>GITHUB_TOKEN</code>, <code>DOCKERHUB_USERNAME</code>,
  <code>DOCKERHUB_TOKEN</code>, and <code>GORELEASER_AUTH_TOKEN</code>.
- **Variable used:** <code>DOCKERHUB_IMAGE</code>, with <code>easyp/easyp</code> as the workflow
  fallback value.

<code>.goreleaser.yaml</code> configures Darwin, Windows and Linux binaries for its
listed architectures, tar.gz archives (zip on Windows), checksums, a source
archive and a Homebrew formula. Its active image target is
<code>ghcr.io/easyp-tech/easyp</code> for <code>linux/amd64</code> and <code>linux/arm64</code>, using root
<code>Dockerfile</code>. The Docker Hub image entry is commented out even though the
workflow still logs in there.

The Homebrew install snippet obtains Bash and Zsh completion scripts from the
registered <code>completion</code> command in <code>internal/api/completion.go</code>
and verifies the installed binary with <code>--version</code>.

### Documentation site

The site and its publishing workflow live in the separate <code>easyp-tech/docs</code>
repository. This checkout has neither a site directory nor a docs-publishing
workflow. Do not infer its current deployment steps from removed workflow
text; the external repository was not inspected for this audit.

### <code>.github/workflows/relator.yml</code>

This workflow is not a deployment pipeline. It sends Telegram notifications
when issues or pull requests are opened or reopened. It requires
<code>TELEGRAM_BOT_TOKEN</code>, <code>TELEGRAM_CHAT_ID</code>, and <code>GITHUB_TOKEN</code>.

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
procedure is to correct the issue and publish a new stable <code>vX.Y.Z</code> tag through
the release workflow.

## Secrets Management

| Secret / Variable | Purpose | Where Set |
|---|---|---|
| <code>GITHUB_TOKEN</code> | GitHub API, package-registry and checkout authentication | GitHub Actions |
| <code>GORELEASER_AUTH_TOKEN</code> | GoReleaser release authentication | GitHub Actions secret |
| <code>DOCKERHUB_USERNAME</code> | Docker Hub login username | GitHub Actions secret |
| <code>DOCKERHUB_TOKEN</code> | Docker Hub login token | GitHub Actions secret |
| <code>DOCKERHUB_IMAGE</code> | Docker Hub image name override | GitHub Actions variable |
| <code>TELEGRAM_BOT_TOKEN</code> | Relator notification authentication | GitHub Actions secret |
| <code>TELEGRAM_CHAT_ID</code> | Relator notification destination | GitHub Actions variable |

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
