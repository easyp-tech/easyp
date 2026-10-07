package api

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Tidy executes the v1 module operation from the current directory.
func (m Mod) Tidy(ctx *cli.Context) error {
	if flags.IsFrozen(ctx) {
		return fmt.Errorf("mod tidy is not allowed in frozen mode")
	}
	root, err := moduleWorkingDir()
	if err != nil {
		return fmt.Errorf("moduleWorkingDir: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	report, err := modules.TidyWithReport(ctx.Context, root, cache)
	if err != nil {
		return fmt.Errorf("TidyWithReport: %w", err)
	}
	if len(report.Imports) == 0 {
		return nil
	}
	writer := ctx.App.Writer
	if writer == nil {
		writer = os.Stdout
	}
	for _, change := range report.Imports {
		_, err = fmt.Fprintf(writer, "%s: import %q -> %q (%s %s commit %s roots %v -> %s commit %s roots %v)\n", change.File, change.From, change.To, change.Module, change.OldVersion, change.OldCommit, change.OldRoots, change.Version, change.Commit, change.Roots)
		if err != nil {
			return fmt.Errorf("Fprintf: %w", err)
		}
	}
	_, err = fmt.Fprintln(writer, "Review generation selectors in easyp.gen.yaml and generated SDK import paths for the new dependency namespace.")
	if err != nil {
		return fmt.Errorf("Fprintln: %w", err)
	}
	return nil
}

// Download executes the v1 module operation from the current directory.
func (m Mod) Download(ctx *cli.Context) error {
	root, err := moduleWorkingDir()
	if err != nil {
		return fmt.Errorf("moduleWorkingDir: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	if flags.IsFrozen(ctx) {
		_, err := modules.EnsureFrozenSources(ctx.Context, root, cache)
		return err
	}
	return modules.Download(ctx.Context, root, cache)
}
