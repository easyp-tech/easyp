# EasyP

[![License](https://img.shields.io/github/license/easyp-tech/easyp?color=blue)](https://github.com/easyp-tech/easyp/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/easyp-tech/easyp?include_prereleases)](https://github.com/easyp-tech/easyp/releases)
[![CI](https://github.com/easyp-tech/easyp/workflows/ci/badge.svg)](https://github.com/easyp-tech/easyp/actions?workflow=ci)

**Modern Protocol Buffers toolkit for streamlined development workflows**

The `easyp` CLI is a comprehensive tool for working with [Protocol Buffers](https://protobuf.dev). It provides:

- A **linter** that enforces good API design choices and structure
- A **breaking change detector** that ensures compatibility at the source code level
- A **generator** that invokes plugins based on configuration files
- A **package manager** with Git-based dependency management
- **Integration with remote plugins** for consistent, isolated execution

## Installation

### Homebrew

You can install `easyp` using [Homebrew](https://brew.sh) (macOS or Linux):

```sh
brew install easyp-tech/tap/easyp
```

### Go Install

```sh
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

### Other methods

For other installation methods, see our [official documentation](https://easyp.tech/docs/guide/introduction/install), which covers:

- Installing `easyp` via npm
- Using `easyp` as a Docker image
- Installing as a binary from GitHub Releases

## Quick Start

```sh
# Initialize a new project
mkdir my-proto-project && cd my-proto-project
easyp init

# Add your .proto files to the project
mkdir api
# ... add your .proto files to api/ ...

# Download dependencies and generate code
easyp mod tidy
easyp mod download
easyp generate

# Lint your proto files
easyp lint

# Check for breaking changes
easyp breaking --against main
```

## Usage

EasyP's help interface provides summaries for commands and flags:

```sh
easyp --help
```

For comprehensive usage information, consult EasyP's [documentation](https://easyp.tech), especially these guides:

* [What is EasyP?](https://easyp.tech/docs/guide/introduction/what-is) - Overview and key concepts
* [`easyp lint`](https://easyp.tech/docs/guide/cli/linter/linter) - Code linting and validation
* [`easyp breaking`](https://easyp.tech/docs/guide/cli/breaking-changes/breaking-changes) - Breaking change detection
* [`easyp mod`](https://easyp.tech/docs/guide/cli/package-manager/package-manager) - Package management
* [`easyp generate`](https://easyp.tech/docs/guide/cli/generator/generator) - Code generation
* `easyp validate-config` - Validate `easyp.yaml` structure and types (JSON or text output)
* Global flag: `--format, -f` / env `EASYP_FORMAT` (`text` or `json`) for commands that support formatted output

## Key Features

- **🔍 Comprehensive Linting** - Built-in support for buf's linting rules with customizable configurations
- **📦 Smart Package Manager** - Git-based dependency management with lock file support
- **⚡ Code Generation** - Multi-language generation with local and remote plugin support
- **🔄 Breaking Change Detection** - Automated API compatibility verification against Git branches
- **🌐 Remote Plugin Support** - Execute plugins via centralized EasyP API service
- **🎯 Developer Experience** - Auto-completion, intuitive commands, and clear error messages

## Why choose EasyP over buf.build?

While buf.build provides excellent protobuf tooling, EasyP offers several key advantages:

| Feature | EasyP | buf.build |
|---------|--------|-----------|
| **Dependencies** | Any Git repository | Buf Schema Registry (BSR) required |
| **Vendor Lock-in** | None | Tied to BSR for full features |
| **Plugin Execution** | Local + Remote plugins | local + BSR |
| **Enterprise** | Works with existing Git infrastructure | Requires BSR setup |

**Key Benefits:**
- **No infrastructure changes**: Use your existing Git repositories for proto dependencies
- **Enhanced flexibility**: Execute plugins both locally and remotely for consistent results
- **Separated configuration**: `protobuf.mod`, `easyp.gen.yaml`, and `easyp.yaml` each own one part of the workflow
- **Buf dependency roots**: Git dependencies using Buf v1 or v2 configuration can supply import roots

## Our goals for Protobuf

EasyP's goal is to make Protocol Buffers development more accessible and reliable by providing a **unified toolkit** that eliminates the complexity of traditional protobuf workflows. We've built on the proven foundation of Protocol Buffers and buf's excellent design principles to create a modern development experience.

While Protocol Buffers offer significant technical advantages over REST/JSON, actually _using_ them has traditionally been more challenging than necessary. EasyP aims to change that by consolidating the entire protobuf workflow into a single, intuitive tool with Git-native dependency management and both local and remote plugin execution.

## Configuration

The v1 pilot uses `protobuf.mod` for module identity, source roots, and dependencies. `easyp.gen.yaml` selects modules and plugins; `easyp.yaml` configures lint and breaking checks. Run `easyp get <module>[@version|@commit]` from the module directory to add a direct requirement and record its transitive dependencies as `// indirect`. The command writes pinned commits and content hashes to `protobuf.lock`; `easyp mod tidy` resolves requirements already in the manifest.

The [`easyp_config_describe` MCP tool](mcp/easypconfig/README.md) describes the editable v1 YAML files and uses the same JSON Schemas as the CLI.

In `easyp.yaml` and `easyp.gen.yaml`, EasyP expands environment variables before parsing and validation. Use `${NAME}`, `${NAME:-default}` for a fallback, or `$${NAME}` to keep a literal `${NAME}`. An unset variable without a fallback expands to an empty string. Substitutions inside YAML comments are ignored. `protobuf.mod` and `protobuf.lock` do not expand environment variables.

```text
# protobuf.mod
module github.com/acme/contracts
roots proto
require github.com/acme/weather v1.2.0
```

```yaml
# easyp.gen.yaml
version: v1
generate:
  modules: [github.com/acme/contracts]
plugins:
  - name: go
    out: ./gen/go
    opts:
      paths: source_relative
```

A plugin can select one source: `name` for a local or bundled plugin, `path` for an explicit binary, `command` for an executable and its arguments, or `remote` with a pinned `version`. Relative binary paths and commands run from the directory where `easyp generate` starts.

```yaml
# easyp.yaml
version: v1
linters:
  default: STANDARD
breaking:
  baseline: git:main
```

# Configuration validation

`easyp validate-config` recursively validates every `easyp.yaml`, `easyp.gen.yaml`, `protobuf.mod`, and `protobuf.lock` under the current directory. Use `--config` to validate one file or all matching files under a selected directory. Errors include the file path and, for YAML structure errors, the line and column. The command exits with a non-zero status when errors are found.

```sh
# Validate all EasyP configuration and module files below the current directory
easyp validate-config

# Validate one generator file with text output
easyp --format text validate-config --config easyp.gen.yaml

# Validate every matching file below a directory
easyp validate-config --config ./services
```

## Community

For help and discussion around EasyP and Protocol Buffers best practices:

- **📖 [Documentation](https://easyp.tech)** - Comprehensive guides and API reference
- **💬 [Telegram Chat](https://t.me/easyptech)** - Community discussion and support
- **🐛 [GitHub Issues](https://github.com/easyp-tech/easyp/issues)** - Bug reports and feature requests
- **✉️ [Contact](mailto:support@easyp.tech)** - Direct contact for enterprise support

## Next steps

Once you've installed `easyp`, we recommend completing the [Quick Start tutorial](https://easyp.tech/docs/guide/introduction/quickstart), which provides a hands-on overview of the core functionality. The tutorial takes about 10 minutes to complete.

After completing the tutorial, check out the [documentation](https://easyp.tech) for your specific areas of interest.

## License

EasyP is released under the [Apache License 2.0](LICENSE).

---

*Built with ❤️ for the Protocol Buffers community*
