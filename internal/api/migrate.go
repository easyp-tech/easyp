package api

import (
	"fmt"
	"io"
	"os"

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
		Usage:       "preview a safe EasyP v0 to v1 migration; use --write to apply",
		Description: "Prepares actual v1 files without executing plugins. Dependency verification requires --resolve-lock; legacy easyp.lock is retained. Existing native outputs are never overwritten.",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dir", Value: ".", Usage: "directory containing legacy easyp.yaml"},
			&cli.StringFlag{Name: "module", Required: true, Usage: "canonical identity for the local protobuf module"},
			&cli.BoolFlag{Name: "write", Usage: "apply the validated candidates with historical backups"},
			&cli.BoolFlag{Name: "resolve-lock", Usage: "explicitly allow dependency resolution/integrity verification and cache writes, including during preview"},
		},
		Action: m.Action,
	}
}

// Action previews the entire plan before any explicitly requested application.
func (Migrate) Action(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		return fmt.Errorf("migrate accepts flags only: --dir <root> --module <identity> [--resolve-lock] [--write]")
	}
	options := migration.Options{Dir: ctx.String("dir"), Module: ctx.String("module"), ResolveLock: ctx.Bool("resolve-lock")}
	if options.ResolveLock {
		storage, err := getEasypPath(getLogger(ctx))
		if err != nil {
			return fmt.Errorf("getEasypPath: %w", err)
		}
		options.Repository = gitmodules.New(storage)
	}
	plan, err := migration.Build(ctx.Context, options)
	if err != nil {
		return fmt.Errorf("Build: %w", err)
	}
	writer := ctx.App.Writer
	if writer == nil {
		writer = os.Stdout
	}
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
