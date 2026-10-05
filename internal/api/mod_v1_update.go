package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Update executes the v1 module operation from the current directory.
func (m Mod) Update(ctx *cli.Context) error {
	if flags.IsFrozen(ctx) {
		return fmt.Errorf("mod update is not allowed in frozen mode")
	}
	root, err := moduleWorkingDir()
	if err != nil {
		return fmt.Errorf("moduleWorkingDir: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	return modules.Update(ctx.Context, root, cache)
}
