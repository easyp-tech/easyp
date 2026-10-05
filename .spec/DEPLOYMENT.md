# Release distribution

EasyP is distributed as CLI/MCP binaries and a CLI container. Human-facing
installation instructions live in README.md and the separate easyp-tech/docs
repository. The v1 source can be merged into main while the latest stable
release remains v0.17.0.

## Channels

| Channel | Selection | Published by this branch |
|---|---|---|
| Stable | GitHub Releases latest, Homebrew easyp, Go @latest, Docker latest | No |
| v1 nightly | Exact v1.0.0-nightly.YYYYMMDD.N tag | Yes, on explicit tag push |
| Local snapshot | goreleaser release --snapshot --clean | Local artifacts only |

The nightly GoReleaser configuration sets release.prerelease to true and
release.make_latest to false, contains no Homebrew publisher, and has exactly
one Docker tag template: the Git tag. There is no floating nightly tag. Existing
stable release assets, Homebrew formula and Docker latest remain on v0.

Do not create a stable v1.0.0 Git tag during this period. Go selects module
versions from Git tags, independently of GitHub Release flags. A stable v1 tag
would become eligible for @latest even if its GitHub release were hidden or
marked as a prerelease. Removing a tag after Go proxies have cached it is not
an effective rollback.

## Nightly workflow

.github/workflows/release.yml runs only for v1.0.0-nightly.* tag pushes. It:

1. Checks out full history, including origin/main and origin/v1.0.
2. Runs scripts/validate-release-tag.sh before registry login. It accepts
   v1.0.0-nightly.YYYYMMDD.N, where N is positive without leading zeros, and
   requires the tag commit to be an ancestor of origin/main or origin/v1.0.
   Stable tags and other prerelease forms are rejected.
3. Sets up QEMU and Docker Buildx, logs into GHCR with GITHUB_TOKEN, and selects
   Go from go.mod (1.26.6).
4. Runs GoReleaser OSS 2.18.2 with release --clean --timeout=90m, using
   GORELEASER_AUTH_TOKEN for GitHub publishing. This is an ordinary SemVer
   prerelease, not GoReleaser Pro's --nightly feature.

A push or merge to main alone does not publish anything. There is no scheduled
or workflow_dispatch release in this branch. GitHub requires those triggers
in the default branch; add them there explicitly if automatic nightlies are
wanted later. Runs for the same release tag are serialized.

### Publishing a nightly

Only run these commands when a release is intended. Preparing or merging the
branch does not require creating a tag.

```sh
git fetch origin --prune --tags
git switch main
git merge --ff-only origin/main
# Verify the selected commit and its CI results before tagging.
EASYP_NIGHTLY_TAG="v1.0.0-nightly.$(date -u +%Y%m%d).1"
# Increase the sequence if that date already has a published nightly.
git tag -a "$EASYP_NIGHTLY_TAG" -m "EasyP $EASYP_NIGHTLY_TAG"
bash scripts/validate-release-tag.sh "$EASYP_NIGHTLY_TAG"
git push origin "refs/tags/$EASYP_NIGHTLY_TAG"
```

Tags are immutable. If publication fails, inspect the workflow and existing
assets before retrying; do not move a published tag. Corrections get a new
sequence. Stable v0 maintenance must use v0 sources and its release workflow;
this branch deliberately has no stable release publisher.

## Artifacts and version metadata

GoReleaser builds easyp and easyp-mcp for Darwin, Windows and Linux with the
architecture matrix in .goreleaser.yaml. Each platform archive contains both
binaries, LICENSE and README.md. Windows uses zip; other platforms use tar.gz.
Checksums and a source archive are uploaded alongside them. The source archive
respects .gitattributes; .agents is export-ignore.

Both binaries receive internal/version.releaseVersion through the linker's -X
flag. Their System() metadata and compiler metadata use that release tag; Go
install and ordinary source builds keep embedded module/VCS version fallback.
Snapshots receive the snapshot version rather than an older stable tag.

Docker publishes ghcr.io/easyp-tech/easyp:<exact-nightly-tag> for linux/amd64
and linux/arm64. The root Dockerfile builds the CLI from source and receives
RELEASE_VERSION from GoReleaser, so its --version matches the release. There is
no MCP container. Local task docker:build never pushes and does not require a
release version. Runtime APK versions remain pinned in Dockerfile.

## Checks before publishing

Use Go 1.26.6 and GoReleaser 2.18.2 without upgrading repository dependencies.

```sh
goreleaser check
go test -mod=readonly -race -count=1 ./scripts ./internal/version ./internal/core
task lint:go
task dev-tools:check
```

The normal tests workflow runs the full race suite, Go lint, Linux ARMv7 compile,
proto example and developer-task guards, and the pinned GoReleaser config check.
The scripts package checks immutable Docker tags, GitHub prerelease/latest
settings, absence of Homebrew publishing, version linker flags, supported tag
forms and source branches, and validation before publishing credentials.

For packaging checks, use an isolated checkout with a temporary nightly tag and
run goreleaser release --clean --skip=publish,docker. Do not create verification
tags in the user's checkout or push them. For Docker, use a unique local tag and
verify --version with an explicit RELEASE_VERSION build argument. Snapshot Task
commands use the local dist directory; run them in an isolated checkout when
existing artifacts must be preserved.

The documentation site is built and published by easyp-tech/docs. Its latest
version selector validates published stable releases and fails closed if GitHub
returns only prereleases or invalid metadata; it never promotes a nightly via
a raw-tag fallback.
