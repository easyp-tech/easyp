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
- **Simplified configuration**: Single file for all protobuf operations
- **Full buf compatibility**: Drop-in replacement with familiar rule sets and configuration

## Our goals for Protobuf

EasyP's goal is to make Protocol Buffers development more accessible and reliable by providing a **unified toolkit** that eliminates the complexity of traditional protobuf workflows. We've built on the proven foundation of Protocol Buffers and buf's excellent design principles to create a modern development experience.

While Protocol Buffers offer significant technical advantages over REST/JSON, actually _using_ them has traditionally been more challenging than necessary. EasyP aims to change that by consolidating the entire protobuf workflow into a single, intuitive tool with Git-native dependency management and both local and remote plugin execution.

## Configuration

EasyP uses a single `easyp.yaml` file for all configuration:

```yaml
# Dependencies
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/bufbuild/protoc-gen-validate@v0.9.1

# Code generation
generate:
  inputs:
    - directory: "api"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
    - remote: api.beta.easyp.tech/community/stephenh-ts-proto:v1.178.0
      out: ./web/generated
      opts:
        env: node
        useExactTypes: false
        outputServices:
          - grpc-js
          - generic-definitions
        esModuleInterop: true
        useOptionals: true

# Linting
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
    - PACKAGE_DEFINED
    - FIELD_LOWER_SNAKE_CASE
    - MESSAGE_PASCAL_CASE

# Configuration validation

`easyp validate-config` validates `easyp.yaml` (or a custom path passed via `--config`). It expands env vars, checks required fields and types, warns on unknown keys, and exits with a non-zero status when errors are found.

```sh
# Validate the default easyp.yaml with JSON output (default)
easyp validate-config

# Validate a custom file with text output (global --format flag)
easyp --format text validate-config --config example.easyp.yaml
```

### Config Schema Integration

The source of truth for config schema + MCP tool metadata lives in [`mcp/easypconfig`](mcp/easypconfig/README.md).

- Go integration: import `github.com/easyp-tech/easyp/mcp/easypconfig` and call `RegisterTool(...)` / `Describe(...)`.
- Cross-language integration: consume generated JSON Schema artifacts in `schemas/easyp-config-v1.schema.json` and `schemas/easyp-config.schema.json`.
- Regenerate artifacts: `task schema:generate` (or `go run ./cmd/easyp-schema-gen`).

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
