package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod/dependency"
	"github.com/easyp-tech/easyp/internal/mod/repo"
	"github.com/easyp-tech/easyp/internal/mod/utils"
)

var _ repo.Repo = (*gitRepo)(nil)

// gitRepo implements repo.Repo interface
type gitRepo struct {
	// remoteURL full repository remoteURL address with schema
	remoteURL string
	// cacheDir local cache directory for store repository
	cacheDir string

	version string
}

// Some links from go mod:
// cmd/go/internal/modfetch/codehost/git.go:65 - create work dir
// cmd/go/internal/modfetch/codehost/git.go:137 - git's struct

// New returns gitRepo instance
// remoteURL: full remoteURL address with schema
func New(ctx context.Context, dep dependency.Dependency, cacheDir string) (repo.Repo, error) {
	r := &gitRepo{
		remoteURL: getRemote(dep.Name),
		cacheDir:  cacheDir,
		version:   dep.Version,
	}

	// TODO: check if dir is already exists
	if _, err := utils.RunCmd(ctx, r.cacheDir, "git", "init", "--bare"); err != nil {
		return nil, fmt.Errorf("mod.RunCmd (init): %w", err)
	}

	_, err := utils.RunCmd(ctx, r.cacheDir, "git", "remote", "add", "origin", r.remoteURL)
	if err != nil {
		return nil, fmt.Errorf("mod.RunCmd (add origin): %w", err)
	}

	_, err = utils.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", r.version,
	)
	if err != nil {
		// it's hard to parse git stderr
		// but since previous command doesn't have any errors we can rely that version is invalid
		return nil, repo.ErrVersionNotFound
	}

	return r, nil
}

func getRemote(name string) string {
	return "https://" + name
}
