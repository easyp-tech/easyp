# EasyP Installation Guide

## Which version?

| Channel | Installs | Notes |
|---|---|---|
| `brew install easyp-tech/tap/easyp` | **v0.17.0** (stable v0) | Homebrew has no v1 formula yet |
| `go install github.com/easyp-tech/easyp/cmd/easyp@latest` | **v0.17.0** | `@latest` skips prerelease tags |
| `ghcr.io/easyp-tech/easyp:latest` | v0 | Docker `latest` is not moved by nightlies |
| `v1.0.0-nightly.YYYYMMDD.N` tags | **v1** | Explicit opt-in: Go install, release archive or Docker tag |

v1 is shipped only as nightly prerelease tags until the stable v1.0.0 release. Before choosing, check what the project uses (`easyp --version`, CI pins, and the config format — see [migration-v0-to-v1.md](./migration-v0-to-v1.md#step-1--inventory)). A v1 binary refuses v0 configuration, and v0 cannot read v1 files.

List published nightlies:

```bash
gh release list -R easyp-tech/easyp --limit 10      # or https://github.com/easyp-tech/easyp/releases
```

## v1 nightly

### Go install (Go 1.26.6+)

Install into a separate directory, so the v0 binary stays available:

```bash
EASYP_NIGHTLY_VERSION=v1.0.0-nightly.20261008.1     # pick a published tag
EASYP_NIGHTLY_BIN="$HOME/.local/share/easyp-nightly/bin"
mkdir -p "$EASYP_NIGHTLY_BIN"
GOBIN="$EASYP_NIGHTLY_BIN" go install "github.com/easyp-tech/easyp/cmd/easyp@$EASYP_NIGHTLY_VERSION"
GOBIN="$EASYP_NIGHTLY_BIN" go install "github.com/easyp-tech/easyp/cmd/easyp-mcp@$EASYP_NIGHTLY_VERSION"   # optional MCP server
"$EASYP_NIGHTLY_BIN/easyp" --version
```

Put `$EASYP_NIGHTLY_BIN` first in `PATH` for v1 projects, or call the binary by full path.

### Release archive

Archives are named `easyp-<version-without-v>-<os>-<arch>.tar.gz` (`.zip` on Windows) and contain `easyp` and `easyp-mcp`, plus a `checksums.txt`:

```bash
V=1.0.0-nightly.20261008.1
gh release download "v$V" -R easyp-tech/easyp -p "easyp-$V-linux-amd64.tar.gz" -p "easyp-$V-checksums.txt"
sha256sum -c --ignore-missing "easyp-$V-checksums.txt"
tar xzf "easyp-$V-linux-amd64.tar.gz"
sudo install "easyp-$V-linux-amd64/easyp" /usr/local/bin/easyp
```

Platforms: `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`, `linux-armv6`, `linux-armv7`, `windows-amd64`, `windows-arm64`.

### Docker

```bash
docker run --rm -v "$PWD:/workspace" -w /workspace \
  ghcr.io/easyp-tech/easyp:v1.0.0-nightly.20261008.1 lint
```

Only the exact nightly tag is published; there is no floating v1 tag.

### From source

```bash
git clone https://github.com/easyp-tech/easyp && cd easyp
go build -o easyp ./cmd/easyp
```

## v0 (stable, legacy configuration)

Use these only for projects that have not migrated yet, or to record a baseline before migrating:

```bash
brew install easyp-tech/tap/easyp
go install github.com/easyp-tech/easyp/cmd/easyp@v0.17.0
go run github.com/easyp-tech/easyp/cmd/easyp@v0.17.0 --version     # one-off, nothing installed
docker run --rm -v "$PWD:/workspace" -w /workspace ghcr.io/easyp-tech/easyp:v0.17.0 lint
```

## Verify

```bash
easyp --version          # v1.0.0-nightly.… or v0.x
easyp --help             # v1 lists: get, migrate, mod tidy
```

## Shell completions

```bash
easyp completion bash > "$(brew --prefix)/etc/bash_completion.d/easyp"
easyp completion zsh  > "${fpath[1]}/_easyp"
```

## Official documentation

[easyp.tech/docs](https://easyp.tech/docs) and the [EasyP README](https://github.com/easyp-tech/easyp#installation).
