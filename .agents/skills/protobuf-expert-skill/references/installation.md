# EasyP Installation Guide

## Which version?

| Channel | Installs | Notes |
|---|---|---|
| `brew install easyp-tech/tap/easyp` | **v1** (stable) | Installs `easyp` and `easyp-mcp` |
| `go install github.com/easyp-tech/easyp/cmd/easyp@latest` | **v1** | `@latest` skips prerelease tags |
| `ghcr.io/easyp-tech/easyp:latest` | **v1** | Pin an exact tag in CI |
| `v0.17.0` (exact) | v0 | Projects that have not migrated yet |
| `v1.X.Y-nightly.YYYYMMDD.N` tags | prerelease | Explicit opt-in; never moves brew, `@latest` or Docker `latest` |

Before choosing, check what the project uses (`easyp --version`, CI pins, and the config format — see [migration-v0-to-v1.md](./migration-v0-to-v1.md#step-1--inventory)). A v1 binary refuses v0 configuration, and v0 cannot read v1 files.

List published releases:

```bash
gh release list -R easyp-tech/easyp --limit 10      # or https://github.com/easyp-tech/easyp/releases
```

## v1

### Homebrew

```bash
brew install easyp-tech/tap/easyp      # easyp and easyp-mcp
```

### Go install (Go 1.26.6+)

```bash
go install github.com/easyp-tech/easyp/cmd/easyp@latest
go install github.com/easyp-tech/easyp/cmd/easyp-mcp@latest   # optional MCP server
easyp --version
```

Pin an exact version (`@v1.0.0`) in scripts and CI.

### Release archive

Archives are named `easyp-<version-without-v>-<os>-<arch>.tar.gz` (`.zip` on Windows) and contain `easyp` and `easyp-mcp`, plus a `checksums.txt`:

```bash
V=1.0.0
gh release download "v$V" -R easyp-tech/easyp -p "easyp-$V-linux-amd64.tar.gz" -p "easyp-$V-checksums.txt"
sha256sum -c --ignore-missing "easyp-$V-checksums.txt"
tar xzf "easyp-$V-linux-amd64.tar.gz"
sudo install "easyp-$V-linux-amd64/easyp" /usr/local/bin/easyp
```

Platforms: `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`, `linux-armv6`, `linux-armv7`, `windows-amd64`, `windows-arm64`.

### Docker

```bash
docker run --rm -v "$PWD:/workspace" -w /workspace \
  ghcr.io/easyp-tech/easyp:v1.0.0 lint
```

`latest` follows the newest stable release; nightlies get only their exact tag.

### From source

```bash
git clone https://github.com/easyp-tech/easyp && cd easyp
go build -o easyp ./cmd/easyp
```

## v0 (legacy configuration)

Use these only for projects that have not migrated yet, or to record a baseline before migrating. Install v0 by exact version, so it can sit next to v1:

```bash
EASYP_V0_BIN="$HOME/.local/share/easyp-v0/bin"
mkdir -p "$EASYP_V0_BIN"
GOBIN="$EASYP_V0_BIN" go install github.com/easyp-tech/easyp/cmd/easyp@v0.17.0
go run github.com/easyp-tech/easyp/cmd/easyp@v0.17.0 --version     # one-off, nothing installed
docker run --rm -v "$PWD:/workspace" -w /workspace ghcr.io/easyp-tech/easyp:v0.17.0 lint
```

## Verify

```bash
easyp --version          # v1.x or v0.x
easyp --help             # v1 lists: get, migrate, mod tidy
```

## Shell completions

```bash
easyp completion bash > "$(brew --prefix)/etc/bash_completion.d/easyp"
easyp completion zsh  > "${fpath[1]}/_easyp"
```

## Official documentation

[easyp.tech/docs](https://easyp.tech/docs) and the [EasyP README](https://github.com/easyp-tech/easyp#installation).
