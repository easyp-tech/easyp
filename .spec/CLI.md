<!-- generated: 2026-07-27, template: cli.md -->
# EasyP CLI

## Overview

EasyP is a Protocol Buffers command-line toolkit for projects that lint schemas, generate code, check API compatibility, and manage Git-based dependencies.

Build a local binary:

```bash
go build -o easyp ./cmd/easyp
```

Typical project workflow:

```bash
easyp lint --path .
easyp generate --path .
```

The CLI is implemented with `urfave/cli/v2`. Its version is available through the framework-provided `--version` flag; the code does not register a standalone `version` command.

## Command Tree

```text
easyp
├── lint (l)                         Lint Protocol Buffer files
├── generate (g)                     Generate code from Protocol Buffer files
├── breaking                         Check for API breaking changes
├── mod (m)                          Manage dependencies
│   ├── download                     Download modules into the local cache
│   ├── update                       Update module versions from configuration
│   └── vendor                       Copy dependency protos into easyp_vendor
├── init (i)                         Interactively initialize configuration
├── ls-files (ls)                    List resolved .proto files
├── validate-config (validate)       Validate an easyp configuration file
├── schema-gen                       Generate configuration JSON Schema artifacts
└── completion
    ├── bash                         Print Bash completion script
    └── zsh                          Print Zsh completion script
```

## Commands Reference

### `easyp lint [flags]`

Lints `.proto` files using the project configuration.

| Flag | Type | Required | Default | Description |
|---|---|---:|---|---|
| `--path`, `-p` | string | yes | `.` | Relative path to the directory containing `.proto` files. |
| `--root`, `-r` | string | no | config-file directory | Root directory used to search for files. |

```bash
easyp lint --path proto
easyp lint --path api --format json
```

Lint issues are written to standard output. Text output has the form `path:line:column:source message (rule)`; JSON output is one JSON object per issue.

### `easyp generate [flags]`

Generates code from `.proto` files according to `easyp.yaml`.

| Flag | Type | Required | Default | Description |
|---|---|---:|---|---|
| `--path`, `-p` | string | yes | `.` | Directory containing `.proto` files. |
| `--root`, `-r` | string | no | config-file directory | Root directory used to search for files. |
| `--descriptor_set_out` | string | no | — | Destination for a binary `FileDescriptorSet`. |
| `--include_imports` | bool | no | `false` | Include transitive dependencies in the descriptor set. |

`--path` can also be supplied through `EASYP_ROOT_GENERATE_PATH`.

```bash
easyp generate --path proto
easyp generate --path proto --descriptor_set_out descriptors.pb --include_imports
```

### `easyp breaking [flags]`

Checks the current Protocol Buffer API for breaking changes relative to a Git ref.

| Flag | Type | Required | Default | Description |
|---|---|---:|---|---|
| `--path`, `-p` | string | yes | `.` | Relative path to the directory containing `.proto` files. |
| `--against` | string | yes | `master` | Branch or Git ref to compare against. |
| `--root`, `-r` | string | no | config-file directory | Root directory used to search for files. |

```bash
easyp breaking --path proto --against main
```

When the configuration does not set `breaking_check.against_git_ref`, the value of `--against` is used. Findings use the same text and JSON formats as `lint`.

### `easyp mod <subcommand>`

Manages the dependencies defined by the configuration. All subcommands use the current working directory as the project root and take no command-specific flags.

| Subcommand | Description |
|---|---|
| `download` | Downloads modules to EasyP's local cache. |
| `update` | Updates module versions using configured versions. |
| `vendor` | Copies dependency `.proto` files to `easyp_vendor`. |

```bash
easyp mod download
easyp mod update
easyp mod vendor
```

For dependency configuration, cache behavior, lockfiles, and vendoring, see [Dependency Management](config/dependency.md).

### `easyp init [flags]`

Interactively creates an EasyP configuration using the available lint rule groups.

| Flag | Type | Required | Default | Description |
|---|---|---:|---|---|
| `--dir`, `-d` | string | yes | `.` | Directory to initialize. |

`--dir` can also be supplied through `EASYP_INIT_DIR`.

```bash
easyp init --dir .
```

### `easyp ls-files [flags]`

Lists resolved `.proto` files, including configured inputs and optionally transitive imports.

| Flag | Type | Required | Default | Description |
|---|---|---:|---|---|
| `--include-imports`, `-I` | bool | no | `true` | Include transitive imports. |

```bash
easyp ls-files
easyp ls-files --format text
```

The default output is indented JSON. Text output lists roots, files, and resolution errors.

### `easyp validate-config [flags]`

Validates the configuration file for syntax and required fields.

This command aliases the global configuration flag as a command flag.

```bash
easyp validate-config --config easyp.yaml
easyp validate-config --format text
```

The default output is JSON with `valid`, and when applicable `errors` and `warnings`. Text output begins with `VALID: true` or `VALID: false`.

### `easyp schema-gen [flags]`

Generates versioned and latest JSON Schema artifacts for EasyP configuration.

| Flag | Type | Required | Default | Description |
|---|---|---:|---|---|
| `--out-versioned` | string | no | schemagen default | Path for the versioned schema file. |
| `--out-latest` | string | no | schemagen default | Path for the latest-schema alias file. |

```bash
easyp schema-gen
easyp schema-gen --out-versioned schemas/easyp-config-v1.schema.json --out-latest schemas/easyp-config.schema.json
```

### `easyp completion <shell>`

Prints a shell completion script to standard output.

| Argument | Allowed values | Description |
|---|---|---|
| `<shell>` | `bash`, `zsh` | Select the completion script to print. |

```bash
easyp completion bash > "$(brew --prefix)/etc/bash_completion.d/easyp"
easyp completion zsh > "${fpath[1]}/_easyp"
```

Only Bash and Zsh completion subcommands are registered. The generated scripts use the CLI's Bash-completion support for dynamic suggestions.

## Global Flags

The application registers these global flags:

| Flag | Type | Default | Environment variable | Description |
|---|---|---|---|---|
| `--cfg`, `--config` | string | `easyp.yaml` | `EASYP_CFG` | Absolute or relative configuration file path. |
| `--debug`, `-d` | bool | `false` | `EASYP_DEBUG` | Enables debug-level logs. |
| `--format`, `-f` | enum | `text` | `EASYP_FORMAT` | Output format: `text` or `json`. Commands can select a different default when the flag is unset. |
| `--version` | bool | `false` | — | Prints the application version provided by `internal/version`. |

The configuration flag is marked required by the CLI definition while its configured value defaults to `easyp.yaml`. For relative paths, EasyP resolves the configuration file from the current working directory. For `lint`, `generate`, `breaking`, and `ls-files`, the configuration file's directory is the project root unless `--root` is provided.

## Configuration

EasyP reads YAML configuration. The effective configuration path is selected in this order: `--cfg` or `--config`, `EASYP_CFG`, then the default `easyp.yaml`.

| Setting | Flag | Environment variable | Config key | Default |
|---|---|---|---|---|
| Configuration path | `--cfg`, `--config` | `EASYP_CFG` | — | `easyp.yaml` |
| Debug logging | `--debug`, `-d` | `EASYP_DEBUG` | — | `false` |
| Output format | `--format`, `-f` | `EASYP_FORMAT` | — | `text` |
| Generate input path | `generate --path`, `-p` | `EASYP_ROOT_GENERATE_PATH` | — | `.` |
| Init target directory | `init --dir`, `-d` | `EASYP_INIT_DIR` | — | `.` |
| Module cache path | — | `EASYPPATH` | — | `$HOME/.easyp` |

EasyP uses `easyp_vendor` as its vendor directory.

## Exit Codes

| Code | Meaning |
|---:|---|
| `0` | Successful command completion. |
| `1` | Lint or breaking-change findings; or a module version was not found during `mod download`, `mod update`, or `mod vendor`. |
| `2` | An imported file cannot be opened, a requested Git ref cannot be found, or the current directory is not a Git repository for `breaking`. |
| non-zero, implementation/runtime dependent | Other returned errors are passed to `log.Fatal`, which terminates the process after printing the error. |

Invalid flags and arguments are handled by `urfave/cli`; this repository does not define a separate explicit `os.Exit` code for those usage errors.

## I/O Contracts

- **stdin:** No command reads standard input directly.
- **stdout:** `lint` and `breaking` print findings; `ls-files` and `validate-config` print JSON by default; `completion` prints a script; `schema-gen` writes schema files.
- **stderr:** The logger writes text logs and errors to standard error. `--debug` includes debug-level log records.
- **files:** `init` creates configuration through an interactive prompt; `generate` writes configured generator outputs and may write a descriptor set; `mod vendor` writes `easyp_vendor`; `schema-gen` writes its two schema artifacts.

Examples for composition:

```bash
easyp lint --path proto --format json > lint-issues.json
easyp ls-files --format json | jq '.files[]'
easyp completion zsh > "${fpath[1]}/_easyp"
```

## Development

Run without creating a persistent binary:

```bash
go run ./cmd/easyp lint --path .
```

Build the local CLI:

```bash
go build -o easyp ./cmd/easyp
```

New commands are added by implementing `api.Handler`, returning a `*cli.Command`, and registering the handler in `cmd/easyp/main.go`.
