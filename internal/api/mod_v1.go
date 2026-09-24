package api

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/modules"
)

// Tidy executes the v1 module operation from the current directory.
func (m Mod) Tidy(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	return modules.Tidy(ctx.Context, root, cache)
}

// Download executes the v1 module operation from the current directory.
func (m Mod) Download(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	return modules.Download(ctx.Context, root, cache)
}
