# Release distribution

EasyP is distributed as CLI/MCP binaries, a Homebrew formula and a CLI
container. Human-facing installation instructions live in README.md and the
separate easyp-tech/docs repository.

## Channels

| Channel | Selection | Published by |
|---|---|---|
| Stable | Tag v1.X.Y on main: GitHub Latest, Homebrew easyp, Go @latest, Docker latest | release.yml, on explicit tag push |
| Nightly | Exact v1.X.Y-nightly.YYYYMMDD.N tag only | release.yml, on explicit tag push |
| Local snapshot | goreleaser release --snapshot --clean | Local artifacts only |

One GoReleaser configuration serves both channels; the tag decides:

- release.prerelease is auto, so a SemVer prerelease (every nightly) is marked
  as one, and release.make_latest is true only when the tag is not a
  prerelease.
- Docker always gets the exact tag; the latest tag template renders empty for
  a prerelease, which GoReleaser skips. There is no floating nightly tag.
- The Homebrew formula (easyp-tech/homebrew-tap, Formula/easyp.rb, installing
  easyp and easyp-mcp) uses skip_upload: auto, so nightlies never reach it.

Go selects module versions from Git tags, independently of GitHub Release
flags: a stable v1.X.Y tag is eligible for @latest as soon as it is pushed, and
removing it after Go proxies have cached it is not a rollback. Tags are never
moved; a bad release is followed by the next patch version.

v0 releases (v0.17.x) belong to the v0 sources and their own pipeline; this
branch rejects v0 tags.

## Release workflow

.github/workflows/release.yml runs for tag pushes matching v1.*.* and:

1. Checks out full history, including origin/main and origin/v1.0.
2. Runs scripts/validate-release-tag.sh before registry login. It prints the
   channel or fails:
   - stable: v1.X.Y without leading zeros, and the tag commit must be an
     ancestor of origin/main;
   - nightly: v1.X.Y-nightly.YYYYMMDD.N, N positive without leading zeros, and
     the tag commit must be an ancestor of origin/main or origin/v1.0.

   Other prerelease forms (rc, beta) and other majors are rejected.
3. Sets up QEMU and Docker Buildx, logs into GHCR with GITHUB_TOKEN, and selects
   Go from go.mod (1.26.6).
4. Runs GoReleaser OSS 2.18.2 with release --clean --timeout=90m, using
   GORELEASER_AUTH_TOKEN for GitHub publishing and the Homebrew tap. A nightly
   is an ordinary SemVer prerelease, not GoReleaser Pro's --nightly feature.

A push or merge to main alone does not publish anything. There is no scheduled
or workflow_dispatch release. Runs for the same release tag are serialized.

### Publishing a nightly

Only run these commands when a release is intended. Preparing or merging a
branch does not require creating a tag. Nightlies are cut towards the next
version: v1.0.0-nightly.* before v1.0.0, v1.1.0-nightly.* after it.

```sh
git fetch origin --prune --tags
git switch main
git merge --ff-only origin/main
# Verify the selected commit and its CI results before tagging.
EASYP_NIGHTLY_TAG="v1.1.0-nightly.$(date -u +%Y%m%d).1"
# Increase the sequence if that date already has a published nightly.
git tag -a "$EASYP_NIGHTLY_TAG" -m "EasyP $EASYP_NIGHTLY_TAG"
bash scripts/validate-release-tag.sh "$EASYP_NIGHTLY_TAG"
git push origin "refs/tags/$EASYP_NIGHTLY_TAG"
```

### Publishing a stable release

A stable tag cannot be taken back, so cut it from a commit that has already
shipped as a nightly and been exercised against the plugin service
(plugins.easyp.tech).

```sh
git fetch origin --prune --tags
git switch main
git merge --ff-only origin/main
# The commit of the nightly that was verified, with green CI on main.
EASYP_TAG=v1.0.0
git tag -a "$EASYP_TAG" -m "EasyP $EASYP_TAG"
bash scripts/validate-release-tag.sh "$EASYP_TAG"   # must print: stable
git push origin "refs/tags/$EASYP_TAG"
```

Then check every channel:

```sh
gh release view --repo easyp-tech/easyp --json tagName,isLatest,isPrerelease
brew update && brew upgrade easyp-tech/tap/easyp && easyp --version
docker run --rm ghcr.io/easyp-tech/easyp:latest --version
GOPROXY=https://proxy.golang.org go list -m github.com/easyp-tech/easyp@latest
```

Finally, publish the matching easyp-tech/docs change; its version selector picks
up the new Latest release on its own.

Tags are immutable. If publication fails, inspect the workflow and existing
assets before retrying; do not move a published tag. A failed nightly gets a
new sequence, a failed stable release the next patch version.

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

Docker publishes ghcr.io/easyp-tech/easyp:<exact-tag> for linux/amd64 and
linux/arm64, plus latest for a stable release. The root Dockerfile builds the CLI from source and receives
RELEASE_VERSION from GoReleaser, so its --version matches the release. There is
no MCP container. Local task docker:build never pushes and does not require a
release version. Runtime APK versions remain pinned in Dockerfile.

## Checks before publishing

Use Go 1.26.6 and GoReleaser 2.18.2 without upgrading repository dependencies.

```sh
bash scripts/check-goreleaser.sh
go test -mod=readonly -race -count=1 ./scripts ./internal/version ./internal/core
task lint:go
task dev-tools:check
```

The normal tests workflow runs the full race suite, Go lint, Linux ARMv7 compile,
proto example and developer-task guards, and the pinned GoReleaser config check.
scripts/check-goreleaser.sh accepts exactly one GoReleaser deprecation, brews:
GoReleaser prefers homebrew_casks, but the tap ships a formula that macOS and
Linux users already upgrade through, and moving it to a cask is a separate
migration. Any other finding fails the check.

The scripts package checks the channel split (Docker latest, GitHub
prerelease/latest and Homebrew only for stable tags), version linker flags,
supported tag forms and source branches, and validation before publishing
credentials.

For packaging checks, use an isolated checkout with a temporary stable or nightly tag and
run goreleaser release --clean --skip=publish,docker. Do not create verification
tags in the user's checkout or push them. For Docker, use a unique local tag and
verify --version with an explicit RELEASE_VERSION build argument. Snapshot Task
commands use the local dist directory; run them in an isolated checkout when
existing artifacts must be preserved.

The documentation site is built and published by easyp-tech/docs. Its latest
version selector validates published stable releases and fails closed if GitHub
returns only prereleases or invalid metadata; it never promotes a nightly via
a raw-tag fallback.
