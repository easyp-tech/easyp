package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/bsr"
	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	"github.com/easyp-tech/easyp/internal/config"
)

func moduleCache(ctx *cli.Context) (*gitmodules.Cache, error) {
	if _, err := gitcommand.Timeout(); err != nil {
		return nil, fmt.Errorf("Timeout: %w", err)
	}
	root, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}
	var cfg config.Config
	if err := cfg.LoadEnvironment(); err != nil {
		return nil, fmt.Errorf("LoadEnvironment: %w", err)
	}
	return gitmodules.NewWithDependencyProxy(root, bsr.StaticResolver{Logger: getLogger(ctx)}, cfg.Proxy).WithLogger(getLogger(ctx)), nil
}
