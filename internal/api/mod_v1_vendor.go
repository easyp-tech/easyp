package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/modules"
)

// Vendor executes the v1 module operation from the current directory.
func (m Mod) Vendor(ctx *cli.Context) error {
	root, err := moduleWorkingDir()
	if err != nil {
		return fmt.Errorf("moduleWorkingDir: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	return modules.Vendor(ctx.Context, root, cache)
}
