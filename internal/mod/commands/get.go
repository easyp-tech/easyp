package commands

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod/dependency"
	"github.com/easyp-tech/easyp/internal/mod/repo/git"
)

// Get download dependency.
// module: string format: origin@version: github.com/company/repo@v1.2.3
// if version is absent use the latest
func (c *Commands) Get(ctx context.Context, module string) error {
	dep := dependency.ParseDependency(module)

	cacheDir, err := c.dirs.CacheDir(dep)
	if err != nil {
		return fmt.Errorf("CreateCacheDir: %w", err)
	}

	repo, err := git.New(ctx, dep, cacheDir)
	if err != nil {
		return fmt.Errorf("git.New: %w", err)
	}

	files, err := repo.GetFiles(ctx)
	if err != nil {
		return fmt.Errorf("repo.GetFiles: %w", err)
	}

	_ = files

	return nil
}
