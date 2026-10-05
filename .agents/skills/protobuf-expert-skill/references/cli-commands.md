# EasyP CLI Commands Reference

## Global Flags

| Flag | Short | Type | Default | Env Var | Description |
|------|-------|------|---------|---------|-------------|
| `--cfg` / `--config` | — | string | `easyp.yaml` | `EASYP_CFG` | Path to config file |
| `--debug` | `-d` | bool | `false` | `EASYP_DEBUG` | Enable debug logging |
| `--format` | `-f` | enum | `text` | `EASYP_FORMAT` | Output format: `text` or `json` |

---

## lint (alias: l)

Lint `.proto` files against configured rules.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--path` | `-p` | string | `.` | Relative path to proto directory |
| `--root` | `-r` | string | project root | Root directory for file search |

**Output formats:**
- Text: `path:line:column:source message (rule_name)`
- JSON: Structured issue objects

**Exit codes:** 0 = no issues, 1 = lint issues found, 2 = critical error (e.g., import failure)

---

## generate (alias: g)

Generate code from proto files using configured plugins.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--path` | `-p` | string | `.` | Relative path to proto directory |
| `--root` | `-r` | string | project root | Root directory for file search |
| `--descriptor_set_out` | — | string | — | Output path for binary FileDescriptorSet |
| `--include_imports` | — | bool | `false` | Include transitive deps in FileDescriptorSet |

**Env:** `EASYP_ROOT_GENERATE_PATH` overrides `--path`.

---

## breaking

Detect breaking API changes by comparing against a git ref.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--path` | `-p` | string | `.` | Relative path to proto directory |
| `--against` | — | string | `master` | Git branch/ref to compare against |

**Exit codes:** 0 = no breaking changes, 1 = breaking changes found, 2 = critical error

---

## mod (alias: m)

Package manager for proto dependencies.

### mod download

Download all dependencies from `deps` and `generate.inputs[].git_repo` to local cache.

No additional flags. Exit code 1 if version not found.

### mod update

Update cached modules to versions specified in config.

No additional flags. Exit code 1 if version not found.

### mod vendor

Copy proto files from cached dependencies into `vendor/` directory.

No additional flags. Exit code 1 if version not found.

---

## init (alias: i)

Interactive configuration setup — creates `easyp.yaml`.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--dir` | `-d` | string | `.` | Directory to initialize |

**Env:** `EASYP_INIT_DIR` overrides `--dir`.

Prompts for: lint rule groups, enum zero value suffix, service suffix, breaking check ref, generate plugins, dependencies.

---

## validate-config (alias: validate)

Validate `easyp.yaml` syntax and structure.

**Output:**
- JSON: `{"valid": bool, "errors": [...], "warnings": [...]}`
- Text: tabular format with counts

**Exit codes:** 0 = valid, 1 = validation errors

---

## ls-files (alias: ls)

List `.proto` files considering inputs and imports.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--include-imports` | `-I` | bool | `true` | Include transitive import dependencies |

**Output:** JSON or text with roots, files (source, import path, absolute path), and errors.

---

## schema-gen

Generate JSON Schema artifacts for `easyp.yaml` (for IDE autocompletion).

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--out-versioned` | string | `schemas/easyp-config-v1.schema.json` | Versioned schema output |
| `--out-latest` | string | `schemas/easyp-config.schema.json` | Latest schema alias output |

---

## completion

Generate shell completion scripts.

### completion bash

Outputs bash completion function for the `easyp` command.

### completion zsh

Outputs zsh completion function for the `easyp` command.
