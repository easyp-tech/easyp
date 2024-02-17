package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models/dependency"
	"github.com/easyp-tech/easyp/internal/package_manager/services"
	"github.com/easyp-tech/easyp/internal/package_manager/services/repository"
)

var _ repository.Repo = (*gitRepo)(nil)

// gitRepo implements repository.Repo interface
type gitRepo struct {
	// remoteURL full repository remoteURL address with schema
	remoteURL string
	// cacheDir local cache directory for store repository
	cacheDir string

	version string
}

// Some links from go package_manager:
// cmd/go/internal/modfetch/codehost/git.go:65 - create work dir
// cmd/go/internal/modfetch/codehost/git.go:137 - git's struct

// New returns gitRepo instance
// remoteURL: full remoteURL address with schema
func New(ctx context.Context, dep dependency.Dependency, cacheDir string) (repository.Repo, error) {
	r := &gitRepo{
		remoteURL: getRemote(dep.Name),
		cacheDir:  cacheDir,
		version:   dep.Version,
	}

	// TODO: check if dir is already exists
	if _, err := services.RunCmd(ctx, r.cacheDir, "git", "init", "--bare"); err != nil {
		return nil, fmt.Errorf("package_manager.RunCmd (init): %w", err)
	}

	_, err := services.RunCmd(ctx, r.cacheDir, "git", "remote", "add", "origin", r.remoteURL)
	if err != nil {
		return nil, fmt.Errorf("package_manager.RunCmd (add origin): %w", err)
	}

	_, err = services.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", r.version,
	)
	if err != nil {
		// it's hard to parse git stderr
		// but since previous command doesn't have any errors we can rely that version is invalid
		return nil, repository.ErrVersionNotFound
	}

	return r, nil
}

func getRemote(name string) string {
	return "https://" + name
}
