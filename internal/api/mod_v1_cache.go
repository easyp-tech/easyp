package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/bsr"
	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
)

func moduleCache(ctx *cli.Context) (*gitmodules.Cache, error) {
	if _, err := gitcommand.Timeout(); err != nil {
		return nil, fmt.Errorf("Timeout: %w", err)
	}
	root, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}
	return gitmodules.NewWithBSRResolver(root, bsr.StaticResolver{Logger: getLogger(ctx)}).WithLogger(getLogger(ctx)), nil
}
