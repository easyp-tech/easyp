package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
)

func moduleCache(ctx *cli.Context) (*gitmodules.Cache, error) {
	root, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}
	return gitmodules.New(root), nil
}
