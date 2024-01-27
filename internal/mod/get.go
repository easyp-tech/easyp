package mod

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod/dependency"
	"github.com/easyp-tech/easyp/internal/mod/repo/git"
	"github.com/easyp-tech/easyp/internal/mod/utils"
)

type GetCommand struct {
}

// Get download dependency.
// module: string format: origin@version: github.com/company/repo@v1.2.3
// if version is absent use the latest
func (c *GetCommand) Get(ctx context.Context, module string) error {
	dep := dependency.ParseDependency(module)

	cacheDir, err := utils.CreateCacheDir(dep)
	if err != nil {
		return fmt.Errorf("CreateCacheDir: %w", err)
	}

	repo, err := git.New(ctx, dep, cacheDir)
	if err != nil {
		return fmt.Errorf("git.New: %w", err)
	}

	_ = repo
	return nil
}
