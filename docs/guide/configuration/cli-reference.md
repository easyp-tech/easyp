# CLI Reference

[[toc]]

## Overview

This reference documents all EasyP commands, their flags, and usage patterns. EasyP follows a command-subcommand structure similar to other modern CLI tools.

## Global Flags

These flags can be used with any EasyP command:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-cfg` | string | `easyp.yaml` | Path to configuration file |
| `--verbose` | `-v` | bool | `false` | Enable verbose output |
| `--debug` | `-d` | bool | `false` | Enable debug logging |
| `--quiet` | `-q` | bool | `false` | Suppress non-error output |
| `--no-color` | | bool | `false` | Disable colored output |
| `--help` | `-h` | bool | `false` | Show help for command |
| `--version` | | bool | `false` | Show EasyP version |

### Examples

```bash
# Use custom config file
easyp -cfg custom.easyp.yaml lint

# Enable verbose output
easyp -v generate

# Debug mode with custom config
easyp -d -cfg prod.yaml breaking

# Quiet mode for CI
easyp -q lint
```

## Commands

### `easyp init`

Initialize a new EasyP project by creating configuration files.

#### Synopsis

```bash
easyp init [flags]
```

#### Description

Creates a new `easyp.yaml` configuration file and `easyp.lock` file in the current directory. If files already exist, it will prompt for confirmation before overwriting.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | Current directory name | Project name |
| `--template` | string | `basic` | Configuration template to use |
| `--force` | bool | `false` | Overwrite existing files without prompting |
| `--minimal` | bool | `false` | Create minimal configuration |

#### Templates

- `basic` - Standard configuration with common settings
- `strict` - Strict linting and quality checks
- `microservice` - Microservices architecture setup
- `monorepo` - Monorepo configuration
- `migration` - For migrating from other tools

#### Examples

```bash
# Initialize with default settings
easyp init

# Initialize with project name
easyp init --name my-api

# Use strict template
easyp init --template strict

# Force overwrite existing config
easyp init --force

# Create minimal config
easyp init --minimal
```

#### Output

Creates the following files:
- `easyp.yaml` - Main configuration file
- `easyp.lock` - Lock file for dependencies (empty initially)

---

### `easyp lint`

Run linter on protobuf files to check for style and consistency issues.

#### Synopsis

```bash
easyp lint [paths...] [flags]
```

#### Description

Analyzes protobuf files according to configured linting rules. If no paths are specified, lints all proto files in configured directories.

#### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `paths` | No | Specific files or directories to lint |

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `text` | Output format |
| `--output` | string | `stdout` | Output destination |
| `--error-format` | string | `text` | Error message format |
| `--config-path` | string | `.` | Config search path |
| `--fix` | bool | `false` | Auto-fix fixable issues (future) |

#### Output Formats

- `text` - Human-readable text output (default)
- `json` - JSON format for programmatic processing
- `msvs` - Visual Studio compatible format
- `junit` - JUnit XML format for CI systems
- `github` - GitHub Actions annotation format
- `gitlab` - GitLab CI compatible format

#### Examples

```bash
# Lint all configured proto files
easyp lint

# Lint specific directory
easyp lint proto/

# Lint specific file
easyp lint proto/user/user.proto

# Output as JSON
easyp lint --format json

# Output to file
easyp lint --output lint-results.txt

# GitHub Actions format for CI
easyp lint --format github

# Multiple paths
easyp lint proto/ api/ internal/
```

#### Exit Codes

- `0` - Success, no issues found
- `1` - Linting issues found
- `2` - Configuration or execution error

---

### `easyp breaking`

Check for breaking changes in your protobuf APIs.

#### Synopsis

```bash
easyp breaking [flags]
```

#### Description

Compares current protobuf files against a previous version (typically another Git branch) to identify backward-incompatible changes.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--against` | string | From config or `main` | Git reference to compare against |
| `--format` | string | `text` | Output format |
| `--output` | string | `stdout` | Output destination |
| `--fail-level` | string | `WIRE` | Minimum breaking level to fail |
| `--exclude-beta` | bool | `false` | Exclude beta packages from check |
| `--limit` | int | `100` | Maximum number of issues to report |

#### Breaking Levels

- `WIRE` - Wire format compatibility (most strict)
- `WIRE_JSON` - Wire and JSON format compatibility
- `PACKAGE` - Package-level compatibility
- `FILE` - File-level compatibility (least strict)

#### Examples

```bash
# Check against main branch (default)
easyp breaking

# Check against specific branch
easyp breaking --against origin/production

# Check against specific tag
easyp breaking --against v1.0.0

# Check against commit
easyp breaking --against abc123def

# Output as JSON
easyp breaking --format json

# Only fail on WIRE_JSON or higher
easyp breaking --fail-level WIRE_JSON

# Exclude beta packages
easyp breaking --exclude-beta
```

#### Exit Codes

- `0` - No breaking changes found
- `1` - Breaking changes detected
- `2` - Configuration or execution error

---

### `easyp generate`

Generate code from protobuf files using configured plugins.

#### Synopsis

```bash
easyp generate [flags]
```

#### Description

Runs configured code generation plugins on your protobuf files, creating language-specific code.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--plugin` | []string | All configured | Specific plugins to run |
| `--input` | []string | All configured | Specific inputs to process |
| `--clean` | bool | `false` | Clean output directories first |
| `--dry-run` | bool | `false` | Show what would be generated |
| `--parallel` | int | CPU cores | Number of parallel jobs |
| `--timeout` | duration | `5m` | Plugin execution timeout |

#### Examples

```bash
# Generate with all configured plugins
easyp generate

# Run specific plugin only
easyp generate --plugin go

# Run multiple specific plugins
easyp generate --plugin go --plugin go-grpc

# Clean before generating
easyp generate --clean

# Dry run to see what would happen
easyp generate --dry-run

# Limit parallelism
easyp generate --parallel 2

# Custom timeout for slow plugins
easyp generate --timeout 10m
```

#### Output

Generated files are placed in directories specified by each plugin's `out` configuration.

---

### `easyp mod`

Manage protobuf package dependencies.

#### Subcommands

- `download` - Download dependencies to local cache
- `update` - Update dependencies to latest versions
- `vendor` - Copy dependencies to local vendor directory
- `tidy` - Remove unused dependencies (future)
- `graph` - Show dependency graph (future)

---

### `easyp mod download`

Download all configured dependencies to local cache.

#### Synopsis

```bash
easyp mod download [flags]
```

#### Description

Downloads all dependencies specified in `deps` configuration. Uses `easyp.lock` if present, otherwise resolves versions from configuration.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--verify` | bool | `true` | Verify checksums |
| `--force` | bool | `false` | Re-download even if cached |
| `--parallel` | int | `4` | Parallel downloads |
| `--retry` | int | `3` | Number of retries on failure |

#### Examples

```bash
# Download all dependencies
easyp mod download

# Force re-download
easyp mod download --force

# Skip verification (not recommended)
easyp mod download --verify=false

# More parallel downloads
easyp mod download --parallel 8
```

---

### `easyp mod update`

Update dependencies to their latest versions.

#### Synopsis

```bash
easyp mod update [modules...] [flags]
```

#### Description

Updates specified modules (or all modules if none specified) to their latest versions according to version constraints in configuration.

#### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `modules` | No | Specific modules to update |

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | `false` | Update all dependencies |
| `--major` | bool | `false` | Allow major version updates |
| `--dry-run` | bool | `false` | Show what would be updated |

#### Examples

```bash
# Update all dependencies
easyp mod update --all

# Update specific module
easyp mod update github.com/googleapis/googleapis

# Update multiple modules
easyp mod update github.com/googleapis/googleapis github.com/grpc-ecosystem/grpc-gateway

# Dry run to see changes
easyp mod update --all --dry-run

# Allow major version updates
easyp mod update --all --major
```

---

### `easyp mod vendor`

Copy all dependencies to local vendor directory.

#### Synopsis

```bash
easyp mod vendor [flags]
```

#### Description

Creates a `easyp_vendor/` directory containing all dependency proto files for offline use or committing to version control.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dir` | string | `easyp_vendor` | Vendor directory path |
| `--prune` | bool | `false` | Remove unused files |
| `--verify` | bool | `true` | Verify checksums |

#### Examples

```bash
# Vendor all dependencies
easyp mod vendor

# Use custom directory
easyp mod vendor --dir vendor/proto

# Prune unused files
easyp mod vendor --prune
```

---

### `easyp completion`

Generate shell completion scripts.

#### Synopsis

```bash
easyp completion <shell> [flags]
```

#### Description

Generates shell completion scripts for bash, zsh, fish, or powershell.

#### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `shell` | Yes | Shell type: bash, zsh, fish, powershell |

#### Examples

```bash
# Bash completion
source <(easyp completion bash)

# Zsh completion
source <(easyp completion zsh)

# Fish completion
easyp completion fish | source

# PowerShell completion
easyp completion powershell | Out-String | Invoke-Expression

# Save to file (bash)
easyp completion bash > /etc/bash_completion.d/easyp

# Save to file (zsh)
easyp completion zsh > "${fpath[1]}/_easyp"
```

---

### `easyp version`

Display version information.

#### Synopsis

```bash
easyp version [flags]
```

#### Description

Shows EasyP version, build information, and compatibility details.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |
| `--check-update` | bool | `false` | Check for updates |

#### Examples

```bash
# Show version
easyp version

# Output as JSON
easyp version --json

# Check for updates
easyp version --check-update
```

#### Output Example

```
EasyP version: v0.5.0
Build date: 2024-01-15
Go version: go1.22
Git commit: abc123def
OS/Arch: darwin/arm64
```

## Environment Variables

EasyP recognizes the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `EASYPPATH` | Cache directory location | `~/.easyp` |
| `EASYP_CONFIG` | Default config file name | `easyp.yaml` |
| `EASYP_COLOR` | Force color output | Auto-detect |
| `NO_COLOR` | Disable all color output | `false` |
| `EASYP_DEBUG` | Enable debug logging | `false` |
| `EASYP_TIMEOUT` | Global timeout for operations | `5m` |
| `EASYP_PARALLEL` | Default parallelism | CPU cores |
| `HTTP_PROXY` | HTTP proxy for downloads | |
| `HTTPS_PROXY` | HTTPS proxy for downloads | |
| `NO_PROXY` | Bypass proxy for domains | |

### Examples

```bash
# Custom cache directory
export EASYPPATH=/opt/easyp-cache
easyp mod download

# Enable debug globally
export EASYP_DEBUG=1
easyp lint

# Disable colors in CI
export NO_COLOR=1
easyp breaking

# Custom timeout
export EASYP_TIMEOUT=10m
easyp generate

# Use proxy
export HTTPS_PROXY=http://proxy.company.com:8080
easyp mod download
```

## Exit Codes

EasyP uses consistent exit codes across all commands:

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | Command-specific failure (issues found, breaking changes, etc.) |
| `2` | Configuration error |
| `3` | Execution error |
| `4` | Network error |
| `5` | Filesystem error |
| `124` | Timeout |
| `130` | Interrupted (Ctrl+C) |

## Command Aliases

Some commands have shorter aliases for convenience:

| Command | Alias |
|---------|-------|
| `generate` | `gen` |
| `breaking` | `break` |
| `completion` | `comp` |

## Configuration Precedence

Configuration values are resolved in the following order (highest to lowest priority):

1. Command-line flags
2. Environment variables
3. Configuration file (`easyp.yaml`)
4. Default values

## Examples by Workflow

### Initial Setup

```bash
# Initialize project
easyp init --name my-api

# Download dependencies
easyp mod download

# Run initial lint
easyp lint

# Generate code
easyp generate
```

### Development Workflow

```bash
# Check for issues
easyp lint

# Check breaking changes
easyp breaking --against main

# Generate code
easyp generate --clean

# Update dependencies
easyp mod update --all
```

### CI/CD Pipeline

```bash
# Lint with CI-friendly output
easyp -q lint --format github

# Strict breaking check
easyp breaking --against origin/main --fail-level WIRE

# Generate all code
easyp generate --clean --parallel 4

# Vendor for reproducible builds
easyp mod vendor
```

### Debugging Issues

```bash
# Enable debug logging
easyp -d lint

# Verbose output
easyp -v generate

# Dry run to see what would happen
easyp generate --dry-run

# Check specific file
easyp lint proto/problematic.proto
```

This CLI reference covers all EasyP commands and their options. For more detailed usage examples and workflows, see the respective feature documentation sections.