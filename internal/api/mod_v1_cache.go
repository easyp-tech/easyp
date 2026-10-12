package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/bsr"
	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	"github.com/easyp-tech/easyp/internal/config"
)

func moduleCache(ctx *cli.Context) (*gitmodules.Cache, error) {
	root, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}
	var cfg config.Config
	if err := cfg.LoadEnvironment(); err != nil {
		return nil, fmt.Errorf("LoadEnvironment: %w", err)
	}
	return gitmodules.NewWithDependencyProxy(root, bsr.StaticResolver{Logger: getLogger(ctx)}, cfg.Proxy), nil
}
