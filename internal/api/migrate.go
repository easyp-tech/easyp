package api

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	"github.com/easyp-tech/easyp/internal/migration"
)

// Migrate provides an explicit, preview-first v0 to v1 migration command.
type Migrate struct{}

var _ Handler = Migrate{}

// Command implements Handler.
func (m Migrate) Command() *cli.Command {
	return &cli.Command{
		Name:        "migrate",
		Usage:       "migrate EasyP v0 to v1 with a terminal wizard or explicit flags",
		Description: "Prepares actual v1 files without executing plugins. Flag-only dependency verification requires --resolve-lock; the wizard asks for permission instead. Legacy easyp.lock is retained. Existing native outputs are never overwritten. Without --module, a terminal starts the wizard; --module preserves the noninteractive preview. Use --interactive to request the wizard explicitly or --interactive=false to disable it.",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dir", Value: ".", Usage: "directory containing legacy easyp.yaml"},
			&cli.StringFlag{Name: "module", Usage: "canonical identity for the local protobuf module"},
			&cli.BoolFlag{Name: "interactive", DefaultText: "automatic on a terminal without --module", Usage: "use the terminal wizard; --interactive=false disables automatic prompting"},
			&cli.BoolFlag{Name: "write", Usage: "apply the validated candidates with historical backups"},
			&cli.BoolFlag{Name: "resolve-lock", Usage: "explicitly allow dependency resolution/integrity verification and cache writes, including during preview"},
		},
		Action: m.Action,
	}
}

// Action previews the entire plan before any explicitly requested application.
func (Migrate) Action(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		return fmt.Errorf("migrate accepts flags only: --dir <root> --module <identity> [--interactive] [--resolve-lock] [--write]")
	}
	reader, writer := migrationIO(ctx)
	interactive, err := migrationInteractiveMode(ctx.IsSet("module"), ctx.IsSet("interactive"), ctx.Bool("interactive"), migrationHasTerminal(reader, writer))
	if err != nil {
		return err
	}
	options := migration.Options{Dir: ctx.String("dir"), Module: ctx.String("module"), ResolveLock: ctx.Bool("resolve-lock")}
	if interactive {
		wizard := migrationWizard{
			prompt: newMigrationPrompt(reader, writer), writer: writer,
			repository: func() (migration.Repository, error) { return migrationRepository(ctx) },
		}
		return wizard.run(ctx.Context, options, ctx.IsSet("dir"), ctx.IsSet("module"))
	}
	if strings.TrimSpace(options.Module) == "" {
		return fmt.Errorf("--module is required in noninteractive mode; use --interactive in a terminal for guided migration")
	}
	if options.ResolveLock {
		options.Repository, err = migrationRepository(ctx)
		if err != nil {
			return fmt.Errorf("migrationRepository: %w", err)
		}
	}
	plan, err := migration.Build(ctx.Context, options)
	if err != nil {
		return fmt.Errorf("Build: %w", err)
	}
	if err := printMigrationPlan(writer, plan); err != nil {
		return err
	}
	if !ctx.Bool("write") {
		if _, err := io.WriteString(writer, "\nPreview only. Use --write to apply these candidates.\n"); err != nil {
			return fmt.Errorf("WriteString: %w", err)
		}
		return nil
	}
	if err := plan.Apply(); err != nil {
		return fmt.Errorf("Apply: %w", err)
	}
	if !plan.AlreadyV1() {
		if _, err := io.WriteString(writer, "\nMigration applied. Legacy backups were preserved.\n"); err != nil {
			return fmt.Errorf("WriteString: %w", err)
		}
	}
	return nil
}

func migrationIO(ctx *cli.Context) (io.Reader, io.Writer) {
	reader := ctx.App.Reader
	if reader == nil {
		reader = os.Stdin
	}
	writer := ctx.App.Writer
	if writer == nil {
		writer = os.Stdout
	}
	return reader, writer
}

func migrationRepository(ctx *cli.Context) (migration.Repository, error) {
	storage, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}
	return gitmodules.New(storage), nil
}

func printMigrationPlan(writer io.Writer, plan *migration.Plan) error {
	for _, warning := range plan.Warnings() {
		if _, err := fmt.Fprintf(writer, "Warning: %s\n", warning); err != nil {
			return fmt.Errorf("Fprintf: %w", err)
		}
	}
	for _, output := range plan.Outputs() {
		status := "candidate"
		if output.Unchanged {
			status = "unchanged"
		}
		if _, err := fmt.Fprintf(writer, "\n--- %s (%s, mode %04o) ---\n%s\n", output.Name, status, output.Mode.Perm(), output.Content); err != nil {
			return fmt.Errorf("Fprintf: %w", err)
		}
	}
	return nil
}
