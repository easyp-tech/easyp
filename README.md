# EasyP

[![License](https://img.shields.io/github/license/easyp-tech/easyp?color=blue)](https://github.com/easyp-tech/easyp/blob/main/LICENSE)
[![Stable release](https://img.shields.io/github/v/release/easyp-tech/easyp)](https://github.com/easyp-tech/easyp/releases/latest)
[![CI](https://github.com/easyp-tech/easyp/actions/workflows/tests.yml/badge.svg)](https://github.com/easyp-tech/easyp/actions/workflows/tests.yml)

**Modern Protocol Buffers toolkit for streamlined development workflows**

The `easyp` CLI is a comprehensive tool for working with [Protocol Buffers](https://protobuf.dev). It provides:

- A **linter** that enforces good API design choices and structure
- A **breaking change detector** that ensures compatibility at the source code level
- A **generator** that invokes plugins based on configuration files
- A **package manager** with Git-based dependency management
- **Integration with remote plugins** for consistent, isolated execution

## Installation

The latest stable release is **v0.17.0**. The source on this branch implements
**v1.0**, which is distributed only through explicitly selected nightly versions.
Merging v1 into `main` does not update the stable installation channels.

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

For archives and Docker installation, see the [installation guide](https://easyp.tech/docs/introduction/install).

### v1 nightly (explicit opt-in)

Choose an existing `v1.0.0-nightly.YYYYMMDD.N` tag from
[GitHub Releases](https://github.com/easyp-tech/easyp/releases). The tag below is a
placeholder; replace it with a published nightly. Building v1 requires Go 1.26.6
or later.

```sh
EASYP_NIGHTLY_VERSION=v1.0.0-nightly.YYYYMMDD.N
EASYP_NIGHTLY_BIN="$HOME/.local/share/easyp-nightly/bin"
mkdir -p "$EASYP_NIGHTLY_BIN"
GOBIN="$EASYP_NIGHTLY_BIN" go install "github.com/easyp-tech/easyp/cmd/easyp@$EASYP_NIGHTLY_VERSION"
GOBIN="$EASYP_NIGHTLY_BIN" go install "github.com/easyp-tech/easyp/cmd/easyp-mcp@$EASYP_NIGHTLY_VERSION"
"$EASYP_NIGHTLY_BIN/easyp" --version
```

Nightly archives contain both `easyp` and `easyp-mcp`, with a checksums file in the
same release. The CLI container uses the exact tag
`ghcr.io/easyp-tech/easyp:<nightly-tag>`. Nightlies do not update Homebrew,
GitHub Latest, or the Docker `latest` tag. `go install ...@latest` continues to
select the stable v0 release while v1 has only prerelease tags.

Go modules use the same `github.com/easyp-tech/easyp` path for v0 and v1; there is
no `/v1` suffix. The `/v2` suffix is required starting with v2. Existing pinned
v0 dependencies keep their versions after the merge, but explicitly upgrading
to v1 may require API and configuration changes. Creating a **stable Git tag**
`v1.0.0` would make it eligible for `@latest`, even without a GitHub Release, so
v1 stable tags must not be created during the nightly period. See
[Go module version numbering](https://go.dev/doc/modules/version-numbers).

Go selects one version per module path for a build. Another dependency requiring
a v1 nightly can raise the selected version through
[minimal version selection](https://go.dev/ref/mod#minimal-version-selection),
so library consumers also need a dependency graph compatible with v0.

## Quick Start

These commands describe the v1 nightly contract. Select the nightly binary
above, or build this checkout, before using them. For an existing v0 project,
run `easyp migrate` to preview its migration and `easyp migrate --write` to apply
the plan.

```sh
# Select the explicitly installed nightly for this terminal's v1 examples.
export PATH="$HOME/.local/share/easyp-nightly/bin:$PATH"
easyp --version

# Initialize a new project
mkdir my-proto-project && cd my-proto-project
easyp init --module github.com/acme/my-proto-project

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

* [What is EasyP?](https://easyp.tech/docs/introduction/what-is) - Overview and key concepts
* [`easyp lint`](https://easyp.tech/docs/cli/linter) - Code linting and validation
* [`easyp breaking`](https://easyp.tech/docs/cli/breaking-changes) - Breaking change detection
* [`easyp mod`](https://easyp.tech/docs/cli/package-manager) - Package management
* [`easyp generate`](https://easyp.tech/docs/cli/generator) - Code generation
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

The development version of EasyP v1.0 uses `protobuf.mod` for module identity, source roots, and dependencies. `easyp.gen.yaml` selects modules and plugins; `easyp.yaml` configures lint and breaking checks. Run `easyp get <module>[@version|@tag|@commit]` from the module directory to add a direct requirement and record its transitive dependencies as `// indirect`. Named Git tags are resolved to full commits before the requirement is written. The command writes pinned commits and content hashes to `protobuf.lock`; `easyp mod tidy` resolves requirements already in the manifest.

Git dependencies using Buf can resolve known BSR dependencies through fixed Git snapshots, including the Googleapis dependency of grpc-gateway and grpc-federation. The lock retains the original BSR reference, commit and digest with `resolution: compatibility_snapshot`; the CLI explicitly reports that BSR revision equivalence and digest verification are not guaranteed. Frozen commands reuse the recorded mapping. Unknown BSR modules fail explicitly. See [supported mappings and resolver boundaries](.spec/config/dependency.md#bsr-compatibility-snapshots).

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


### Working with the v1 pilot

The implementation and `.spec` tree define the current unreleased v1 contract. Some earlier pilot RFC examples were intentionally superseded; see [v1.0 pre-release contract notes](V1_RELEASE_NOTES.md) before treating an old RFC example as a compatibility requirement.

The repository's native configuration has a runnable local example. Run
<code>task proto:check</code> to validate it, lint it and verify repeatable
generation without altering the source checkout. User configuration is split
between <code>easyp.yaml</code>, <code>easyp.gen.yaml</code> and
<code>protobuf.mod</code>/<code>protobuf.lock</code>.

For a v0 project, <code>easyp migrate --module github.com/acme/contracts</code>
previews conversion; it does not write or execute plugins. See the
[documented migration contract](.spec/config/review-migration-and-polish.md)
before applying it with <code>--write</code>.

Run <code>easyp migrate</code> in a terminal for a guided migration, or add
<code>--interactive</code> to use the wizard with prefilled flags. Dependency
access and applying files require separate confirmations, both defaulting to no.
Explicit <code>--module</code> without the wizard keeps the preview-first script
interface; <code>--interactive=false</code> disables automatic prompting.
