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

	// TODO: read buf.work.yaml to determine dir with proto files

	files, err := repo.GetFiles(ctx)
	if err != nil {
		return fmt.Errorf("repo.GetFiles: %w", err)
	}

	// 1. read all files (done)
	// 2. read buf.work.yaml -> read dirs with proto files
	// 3. filter or by buf.work or is it does not exist filter only proto files

	_ = files

	return nil
}
