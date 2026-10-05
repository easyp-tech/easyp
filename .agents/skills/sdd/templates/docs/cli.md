<!-- scope: **/cmd/*, **/cli/*, **/commands/*, **/flags/* -->
# CLI Documentation Template

## What This Generates

- `.spec/CLI.md` — command tree, flags, arguments, configuration, exit codes, I/O contracts

## Instructions

You are a technical documentarian. Create CLI application documentation for the project in the `.spec/` directory.
Analyze: entry point files, command definitions, flag/argument parsers, configuration loaders, and shell completion scripts.

### Step 1: Identify CLI Framework

Check for the presence of:
- **Go**: cobra, urfave/cli, kong, pflag, flag (stdlib)
- **Rust**: clap, structopt, argh
- **Python**: click, typer, argparse, fire
- **Node.js**: commander, yargs, oclif, meow, citty
- **Ruby**: thor, optparse, gli
- **Dart**: args, dcli

Determine from imports, `go.mod`, `Cargo.toml`, `package.json`, `requirements.txt`, `pubspec.yaml`, etc.

### Step 2: Create .spec/CLI.md

#### Structure:

##### 1. Overview
- One sentence: what the CLI does and who uses it
- Installation command (if applicable)
- Quick start example (the most common invocation)

##### 2. Command Tree

ASCII tree of all commands and subcommands:
```
myapp
├── init              # Initialize a new project
├── run               # Run the application
│   ├── --watch       # Watch mode
│   └── --port N      # Port number
├── config
│   ├── get <key>     # Read config value
│   ├── set <key> <v> # Write config value
│   └── list          # List all config values
└── version           # Print version
```

Build from actual command registration code (e.g., `rootCmd.AddCommand()`, `@app.command()`, `.command()`).

##### 3. Commands Reference

For each command (or command group):

```markdown
### `myapp <command> [flags]`

**Description:** One sentence.

| Flag / Argument | Type | Required | Default | Description |
|----------------|------|----------|---------|-------------|
| `<positional>` | string | yes | — | What it is |
| `--flag` / `-f` | bool | no | `false` | What it does |
| `--output` / `-o` | string | no | `stdout` | Output destination |

**Examples:**
\```bash
myapp run --port 8080
myapp config set key value
\```
```

Extract flags from actual struct tags, decorators, or builder calls — do not invent.

##### 4. Configuration

- Config file locations (precedence order): CLI flags → env vars → config file → defaults
- Config file format (YAML, TOML, JSON, INI)
- Config file path resolution (XDG, `$HOME/.config/`, project-local)
- Environment variable naming convention (e.g., `MYAPP_PORT`, `MYAPP_LOG_LEVEL`)

Table:
| Setting | Flag | Env Var | Config Key | Default |
|---------|------|---------|------------|---------|

##### 5. Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Usage error (invalid flags/arguments) |
| ... | Project-specific codes |

Extract from actual code if defined (e.g., `os.Exit()`, `process.exit()`, `std::process::exit()`).

##### 6. I/O Contracts

- **stdin**: Does the CLI read from stdin? What format? (e.g., piped JSON, line-delimited)
- **stdout**: What is printed on success? (data output, structured JSON, human-readable text)
- **stderr**: What goes to stderr? (logs, progress, errors)
- **Files**: Does the CLI create/modify files? Where?

Piping and composition examples:
```bash
cat input.json | myapp process --format json > output.json
myapp list --json | jq '.[] | .name'
```

##### 7. Shell Completion

- Supported shells (bash, zsh, fish, PowerShell)
- Installation commands for each shell
- How completions are generated (static file, dynamic `completion` subcommand)

If no completion support exists, note it explicitly.

##### 8. Global Flags

Flags available on all commands:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--verbose` / `-v` | bool | `false` | Verbose output |
| `--config` / `-c` | string | `~/.config/myapp/config.yaml` | Config file path |
| `--quiet` / `-q` | bool | `false` | Suppress non-error output |

##### 9. Error Messages & Troubleshooting

Common error messages and their resolutions:
| Error | Cause | Fix |
|-------|-------|-----|

##### 10. Development

- How to run the CLI locally during development
- How to build a release binary
- How to add a new command (step-by-step referencing the framework)

## General Rules

- Language: English
- All commands, flags, and examples must come from actual source code — do not invent
- If the CLI has no subcommands (single-command tool), adapt the structure: skip the command tree, expand the flags section
- If the project is not a CLI application, do not generate this file — skip entirely
- After creating, update `.spec/README.md`: add a link under the appropriate section
