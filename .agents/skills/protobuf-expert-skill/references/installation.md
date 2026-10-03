# EasyP Installation Guide

## Homebrew (macOS / Linux)

```bash
brew install easyp-tech/tap/easyp
```

## Go Install

Requires Go 1.21+:

```bash
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

## npm

```bash
npx easyp@latest --help
```

## Docker

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/easyp-tech/easyp:latest lint
```

## Binary from GitHub Releases

Download the appropriate binary from [GitHub Releases](https://github.com/easyp-tech/easyp/releases):

```bash
# Example for Linux amd64
curl -Lo easyp https://github.com/easyp-tech/easyp/releases/latest/download/easyp_linux_amd64
chmod +x easyp
sudo mv easyp /usr/local/bin/
```

## Verify Installation

```bash
easyp --help
```

## Shell Completions

```bash
# Bash
easyp completion bash >> ~/.bashrc

# Zsh
easyp completion zsh >> ~/.zshrc
```

## Official Documentation

For the most up-to-date installation methods, see [easyp.tech/docs/guide/introduction/install](https://easyp.tech/docs/guide/introduction/install).
