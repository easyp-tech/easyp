---
name: epctl-commands
description: "Use when adding or changing EasyP CLI commands, handlers in internal/api, or registration in cmd/easyp/main.go. epctl-commands is the retained legacy skill name."
argument-hint: "Describe the EasyP CLI command you want to add or change"
---

# EasyP CLI Commands — Legacy Skill Name: epctl-commands

The name <code>epctl-commands</code> and its installation path are retained for compatibility. This skill routes current CLI work to <code>github.com/easyp-tech/easyp</code>. EasyP uses <code>github.com/urfave/cli/v2</code>, with handlers in <code>internal/api</code> and registration in <code>cmd/easyp/main.go</code>.

Start with [AGENTS.md](../../../AGENTS.md), the [agent rules](../../../.spec/agent-rules.md), and [.spec/CLI.md](../../../.spec/CLI.md). Read the nearest handler and its tests before extending a command.

## Source Map

| Responsibility | Actual source |
|----------------|---------------|
| Root <code>cli.App</code>, logger initialization, registration | [cmd/easyp/main.go](../../../cmd/easyp/main.go) |
| <code>Handler</code> contract: <code>Command() *cli.Command</code> | [internal/api/interface.go](../../../internal/api/interface.go) |
| Small handler example | [internal/api/schema_gen.go](../../../internal/api/schema_gen.go) |
| Group and subcommands | [internal/api/mod.go](../../../internal/api/mod.go) |
| Validation and text/JSON reports | [internal/api/validate.go](../../../internal/api/validate.go) |
| Logger lookup and core construction | [internal/api/runtime.go](../../../internal/api/runtime.go) |
| Root flags and output format selection | [internal/flags/flags.go](../../../internal/flags/flags.go), [format.go](../../../internal/flags/format.go) |

Use neighboring file names such as <code>schema_gen.go</code>, <code>mod.go</code>, and <code>mod_v1_update.go</code>. Place CLI wiring in <code>internal/api</code>; keep operations in their existing packages: <code>internal/modules</code> for dependencies, <code>internal/generation</code> for generation orchestration, <code>internal/migration</code> for migration, and <code>internal/core</code> for engines.

## urfave/cli v2 Pattern

- The root is a <code>*cli.App</code> with a <code>Commands</code> slice.
- Handlers implement <code>Command() *cli.Command</code>, often on a small exported struct.
- Actions have signature <code>func(ctx *cli.Context) error</code>. Read flags through <code>ctx.String</code>, <code>ctx.Bool</code>, and related methods; use <code>ctx.Args()</code> for positional arguments.
- Pass <code>ctx.Context</code> to operations requiring <code>context.Context</code>.
- Groups put children in <code>cli.Command.Subcommands</code>. The root app's <code>Commands</code> and a group's <code>Subcommands</code> are different fields.

This complete handler-file example adapts the existing <code>SchemaGen</code> handler and uses the real [schemagen API](../../../internal/schemagen/schemagen.go). It shows fresh command-local flags and separate assignment/error checking. It illustrates a replacement pattern for that file, not a second handler to add beside the existing one.

~~~go
package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/schemagen"
)

// SchemaGen writes JSON Schemas for the v1 YAML documents.
type SchemaGen struct{}

var _ Handler = (*SchemaGen)(nil)

// Command implements Handler.
func (s SchemaGen) Command() *cli.Command {
	return &cli.Command{
		Name:   "schema-gen",
		Usage:  "generate JSON Schemas for v1 YAML files",
		Action: s.Action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "out-dir",
				Usage: "directory for generated v1 JSON Schemas",
				Value: schemagen.DefaultOutDir,
			},
		},
	}
}

// Action writes the schemas to the selected output directory.
func (s SchemaGen) Action(ctx *cli.Context) error {
	err := schemagen.Run(schemagen.Options{OutDir: ctx.String("out-dir")})
	if err != nil {
		return fmt.Errorf("Run: %w", err)
	}
	return nil
}
~~~

For engine work, follow <code>buildCore</code> and [core.New(core.Options)](../../../internal/core/core.go). Consult the actual engine signature, such as <code>Core.Lint(context.Context, DirWalker) ([]IssueInfo, error)</code> in [core/lint.go](../../../internal/core/lint.go), rather than inventing a service client or server lifecycle.

## Wiring a Command

1. Add or update the handler in <code>internal/api</code>, implementing <code>Handler</code>. Keep parsing, output, and CLI error decisions at this boundary.
2. For a top-level command, add the handler value to the existing <code>buildCommand(...)</code> call in <code>main</code>. For example, <code>api.SchemaGen{}</code> is already registered there. The helper calls each handler's <code>Command()</code>.
3. For a subcommand, extend the parent's <code>Subcommands</code> slice. Follow <code>Mod.Command</code>, which binds actions such as <code>m.Download</code> and <code>m.Update</code>.
4. Give flags accurate names, aliases, usage text, defaults, and required behavior. Preserve public spellings and inspect parent/local flag precedence.
5. Add focused action tests and registration/argument tests where needed. Use [go-testing](../go-testing/SKILL.md) for process-state and urfave flag isolation.

## Output, Logging, and Errors

- The root defines <code>--cfg</code> (alias <code>--config</code>), <code>--debug</code>, and <code>--format</code> (alias <code>-f</code>). Text/JSON support and default format are command-specific; use <code>flags.GetFormat</code> where appropriate.
- Follow the target command's existing output contract. For writer-based output, use <code>ctx.App.Writer</code> and <code>ctx.App.ErrWriter</code>, with appropriate standard-stream fallbacks when actions can be invoked directly. <code>Validate.Action</code> demonstrates reports through the application writer.
- Preserve output write failures with <code>%w</code>, including buffered flush/JSON encode failures. EasyP has no universal printer abstraction or global <code>--output</code> mode.
- Use <code>getLogger(ctx)</code> for the repository logger. The root installs it in application metadata; do not introduce unrelated global logger state.
- Wrap call failures using only the callee name, such as <code>fmt.Errorf("Run: %w", err)</code>, without a receiver/package prefix. Follow [go-code-style](../go-code-style/SKILL.md).
- Inspect the handler, <code>runtime.go</code>, and <code>main.go</code> before changing exits. Some handlers return errors, some use <code>cli.Exit</code>, and some call process-exit helpers. [.spec/ERRORS.md](../../../.spec/ERRORS.md) documents these boundaries; there is no general gRPC status mapper.

## CLI Test Isolation

Construct fresh commands, flags, mutable flag values, contexts, and writers per parallel case. urfave/cli v2 shares <code>HelpFlag</code>: use <code>HideHelp: true</code> on the test app and every command/subcommand when help is irrelevant. <code>HideHelpCommand</code> alone is insufficient; use <code>HideVersion: true</code> on the app for the shared version flag too.

Action-only tests can use a private <code>flag.FlagSet</code> with <code>cli.NewContext</code>. Tests that alter cwd, environment, or global context/CLI hooks stay sequential or run in subprocesses; paths that exit the process need subprocess coverage. See [go-testing](../go-testing/SKILL.md) and [breaking_baseline_test.go](../../../internal/api/breaking_baseline_test.go).

## Quick Checklist

- [ ] Work targets the EasyP CLI; the legacy skill name/path remain intact.
- [ ] Handler and action signatures use urfave/cli v2.
- [ ] Registration uses <code>buildCommand</code> for root commands and <code>Subcommands</code> for groups.
- [ ] Operations use actual package APIs and preserve context cancellation.
- [ ] Flags, output formats, logger use, and exit behavior match the command contract.
- [ ] Errors retain causes and use callee-only labels.
- [ ] Tests isolate mutable CLI flags and respect process-state constraints.
