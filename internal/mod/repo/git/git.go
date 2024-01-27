package git

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod"
	"github.com/easyp-tech/easyp/internal/mod/repo"
)

var _ repo.Repo = (*gitRepo)(nil)

// gitRepo implements repo.Repo interface
type gitRepo struct {
	// remote full repository remote address with schema
	remote string
	// dir local cache directory for store repository
	dir string
}

// TODO:
// cmd/go/internal/modfetch/codehost/git.go:65 - create work dir
// cmd/go/internal/modfetch/codehost/git.go:137 - git's struct

// New returns gitRepo instance
// remote: full remote address with schema
func New(remote string) (repo.Repo, error) {
	gRepo := &gitRepo{
		remote: remote,
	}

	// TODO: create workDir
	err := mod.CreateCacheDir()
	if err != nil {
		return nil, fmt.Errorf("repo.CreateCacheDir: %w", err)
	}

	return gRepo, nil
}
