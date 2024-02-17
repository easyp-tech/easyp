package mod

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/dependency"
	"github.com/easyp-tech/easyp/internal/package_manager/repo/git"
)

// Get download dependency.
// module: string format: origin@version: github.com/company/repo@v1.2.3
// if version is absent use the latest
func (c *Mod) Get(ctx context.Context, module string) error {
	dep := dependency.ParseDependency(module)

	cacheDir, err := c.dirs.CacheDir(dep.Name)
	if err != nil {
		return fmt.Errorf("CreateCacheDir: %w", err)
	}

	repo, err := git.New(ctx, dep, cacheDir)
	if err != nil {
		return fmt.Errorf("git.New: %w", err)
	}

	// TODO: read HEAD and determine commit (if version is absent)
	// TODO: create ref struct for storage version (commit)
	// TODO: lock file: cmd/go/internal/lockedfile/mutex.go:46

	// TODO: read buf.work.yaml to determine dir with proto files

	files, err := repo.GetFiles(ctx)
	if err != nil {
		return fmt.Errorf("repo.GetFiles: %w", err)
	}

	protoDirs := filterDirs(files)

	archive, err := repo.Archive(ctx, protoDirs...)
	if err != nil {
		return fmt.Errorf("repo.Archive: %w", err)
	}

	_ = archive

	return nil
}
