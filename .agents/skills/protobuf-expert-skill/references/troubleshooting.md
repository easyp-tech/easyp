# EasyP Troubleshooting Guide

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (no issues) |
| 1 | Issues found (lint violations, breaking changes, validation errors) |
| 2 | Critical error (import failure, config error, runtime crash) |

## Common Errors

### Import not found

**Symptom:** `exit code 2` with message like `import "google/protobuf/timestamp.proto" not found`

**Causes & fixes:**
1. Missing dependency — add to `deps` in `easyp.yaml`:
   ```yaml
   deps:
     - github.com/protocolbuffers/protobuf@v25.0
   ```
2. Dependencies not downloaded — run `easyp mod download`
3. Wrong `--path` flag — ensure it points to the correct proto directory

### Config validation errors

**Symptom:** `easyp validate-config` returns errors

**Debug:**
```bash
easyp validate-config --format json
```

**Common causes:**
- Unknown keys (typos in field names)
- Missing required fields (`lint.use` is required)
- Invalid rule names in `lint.use` or `lint.except`
- Plugin missing exactly one source (`name`, `remote`, `path`, or `command`)

### Lint rule not working

**Symptom:** Expected lint violation not reported

**Check:**
1. Rule is in `lint.use` (either directly or via group)
2. Rule is NOT in `lint.except`
3. File path is NOT in `lint.ignore`
4. Rule is NOT suppressed by `// easyp:off` in the proto file
5. Run with `--debug` for verbose output: `easyp lint --debug`

### Breaking check fails with "could not get git ref"

**Symptom:** `easyp breaking --against main` fails

**Causes:**
- Shallow clone — need full history: `git fetch --unshallow` or `fetch-depth: 0` in CI
- Wrong ref name — verify with `git branch -a` or `git tag -l`
- Detached HEAD — specify explicit branch: `--against origin/main`

### Generate produces no output

**Symptom:** `easyp generate` runs but no files are created

**Check:**
1. `generate.inputs` points to directory with `.proto` files
2. Plugin binary is installed and in `$PATH` (for `name` source)
3. `out` directory exists or plugin can create it
4. Run with `--debug`: `easyp generate --debug`
5. For remote plugins, check network connectivity

### Dependency version not found

**Symptom:** `easyp mod download` fails with exit code 1

**Causes:**
- Tag doesn't exist — verify tag on the repository: `git ls-remote --tags <url>`
- Commit hash is wrong or abbreviated — use full hash
- Repository is private — ensure Git credentials are configured

### Circular import detected

**Symptom:** `PACKAGE_NO_IMPORT_CYCLE` violation

**Debug with:**
```bash
easyp ls-files --format json
```

This shows the full import graph. Refactor packages to break the cycle:
- Move shared types to a common package
- Use message wrappers instead of direct cross-package imports

## Debugging Techniques

### Verbose output

```bash
easyp lint --debug
easyp generate --debug
```

### JSON output for parsing

```bash
easyp lint --format json
easyp breaking --format json
easyp validate-config --format json
easyp ls-files --format json
```

### List resolved files

```bash
easyp ls-files                    # With imports
easyp ls-files --include-imports=false  # Without imports
```

### Validate config first

Always validate before running other commands:
```bash
easyp validate-config
```

## Environment Variables

| Variable | Overrides | Description |
|----------|----------|-------------|
| `EASYP_CFG` | `--cfg` | Path to config file |
| `EASYP_DEBUG` | `--debug` | Enable debug logging |
| `EASYP_FORMAT` | `--format` | Output format (text/json) |
| `EASYP_INIT_DIR` | `init --dir` | Init directory |
| `EASYP_ROOT_GENERATE_PATH` | `generate --path` | Generate path |
