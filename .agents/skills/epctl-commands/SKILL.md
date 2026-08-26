---
name: epctl-commands
description: "Pattern for adding CLI commands to epctl utility. Use when: creating new epctl subcommands, extending CLI functionality, working with cmd/epctl/ files. Covers file structure, urfave/cli v3 patterns, output formatting, and command wiring."
argument-hint: "Describe the CLI command you want to add"
---

# epctl CLI Commands — Pattern Guide

You follow this pattern when adding or modifying CLI commands for the `epctl` utility.

## When to Use

- Adding a new subcommand to `epctl`
- Modifying existing CLI commands
- Working with files in `cmd/epctl/`
- Creating new command groups (like `plugins`, `config`)

## Project Structure

```
cmd/epctl/
├── main.go                        # Entry point: app.Execute()
└── internal/
    ├── output/
    │   └── printer.go             # Printer (text/JSON dual output)
    └── app/
        ├── root.go                # Root command + getPrinter() helper
        ├── builder.go             # PluginBuilder (business logic)
        ├── path_filter.go         # PathFilter (business logic)
        ├── plugins_build.go       # epctl plugins build
        ├── plugins_register.go    # epctl plugins register
        ├── plugins_list.go        # epctl plugins list
        └── config_validate.go     # epctl config validate
```

### Shared vs CLI-only

| Location | Scope | Example |
|----------|-------|---------|
| `internal/config/` | Shared (server + CLI) | Config types, `Validate()` |
| `cmd/epctl/internal/` | CLI-only | Printer, commands, builder |

**Rule:** if a package is only used by `epctl`, it lives in `cmd/epctl/internal/`. If both server and CLI use it, it stays in `internal/`.

## CLI Framework

**urfave/cli v3** (`github.com/urfave/cli/v3`).

Key patterns:
- Commands are `*cli.Command` structs
- Action signature: `func(ctx context.Context, cmd *cli.Command) error`
- Global flags defined on root command propagate to all subcommands
- Global `--output` / `-o` flag controls text vs JSON output

## File Naming

Pattern: `<group>_<action>.go`

| File | Command |
|------|---------|
| `plugins_build.go` | `epctl plugins build` |
| `plugins_register.go` | `epctl plugins register` |
| `plugins_list.go` | `epctl plugins list` |
| `config_validate.go` | `epctl config validate` |

If a file defines the parent group command (e.g., `newPluginsCmd()`), it goes in the first action file of that group (currently `plugins_build.go`).

## Command Pattern

### Constructor + runner

Every command consists of two functions:

```go
// Constructor — returns the *cli.Command definition.
func new<Group><Action>Cmd() *cli.Command {
    return &cli.Command{
        Name:      "<action>",
        Usage:     "Short description",
        ArgsUsage: "[optional args]",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:  "some-flag",
                Value: "default",
                Usage: "what it does",
            },
        },
        Action: run<Group><Action>,
    }
}

// Runner — implements the command logic.
func run<Group><Action>(ctx context.Context, cmd *cli.Command) error {
    printer := getPrinter(cmd)

    // 1. Read flags
    someFlag := cmd.String("some-flag")

    // 2. Business logic
    result, err := doSomething(ctx, someFlag)
    if err != nil {
        return fmt.Errorf("doSomething: %w", err)
    }

    // 3. Output via Printer
    return printer.Table(headers, rows)
}
```

### Group command

```go
func newSomeGroupCmd() *cli.Command {
    return &cli.Command{
        Name:  "<group>",
        Usage: "Group description",
        Commands: []*cli.Command{
            newSomeGroupActionCmd(),
            newSomeGroupOtherCmd(),
        },
    }
}
```

## Output via Printer

**Every command MUST use `getPrinter(cmd)`** for output. Never use `fmt.Println` directly.

```go
printer := getPrinter(cmd)

// Simple message
printer.Message("Done!")              // text: "Done!\n"    json: {"message":"Done!"}

// Table
printer.Table(
    []string{"NAME", "STATUS"},       // headers
    [][]string{{"foo", "ok"}},        // rows
)
// text: tabwriter table             json: [{"NAME":"foo","STATUS":"ok"}]

// Error
printer.Error(err)                    // text: "Error: ...\n"  json: {"error":"..."}
```

## Wiring a New Command

1. Create `cmd/epctl/internal/app/<group>_<action>.go`
2. Implement `new<Group><Action>Cmd()` + `run<Group><Action>()`
3. Add to parent group's `Commands` slice:

```go
// In the group command constructor:
Commands: []*cli.Command{
    newExistingCmd(),
    newYourNewCmd(),  // ← add here
},
```

4. If it's a new group, add to root in `root.go`:

```go
Commands: []*cli.Command{
    newPluginsCmd(),
    newConfigCmd(),
    newYourGroupCmd(),  // ← add here
},
```

## Connecting to EasyP Service

Commands that talk to the service use the SDK:

```go
client, err := sdk.NewClient(addr, sdk.WithInsecure())
if err != nil {
    return fmt.Errorf("cannot connect to %s: %w", addr, err)
}

defer func() { _ = client.Close() }()
```

Standard flags for service-connected commands:
```go
&cli.StringFlag{
    Name:  "addr",
    Value: "localhost:8080",
    Usage: "gRPC server address",
},
```

## Quick Checklist

Before submitting a new CLI command, verify:

- [ ] File named `<group>_<action>.go` in `cmd/epctl/internal/app/`
- [ ] Uses `getPrinter(cmd)` for all output
- [ ] Constructor `new<Group><Action>Cmd()` returns `*cli.Command`
- [ ] Runner `run<Group><Action>(ctx, cmd) error` handles business logic
- [ ] Wired into parent group's `Commands` slice
- [ ] Flags have defaults and usage descriptions
- [ ] Errors wrapped with `fmt.Errorf("<function>: %w", err)`
