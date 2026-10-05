package go_git

import (
	"errors"
	"fmt"
	pathpkg "path"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/go_git"
)

func (g *GoGit) GetDirWalker(workingDir, gitRef, path string) (core.DirWalker, error) {
	// Open the repository by searching upward from workingDir, so --root can be
	// any subdirectory of the git repository (not only the config directory).
	repository, err := gogit.PlainOpenWithOptions(workingDir, &gogit.PlainOpenOptions{
		DetectDotGit: true, EnableDotGitCommonDir: true,
	})
	if err != nil {
		if errors.Is(err, gogit.ErrRepositoryNotExists) {
			return nil, core.ErrRepositoryDoesNotExist
		}

		return nil, fmt.Errorf("git.PlainOpenWithOptions: %w", err)
	}

	// Determine the repo root so we can compute the correct git-tree path.
	wt, err := repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("repository.Worktree: %w", err)
	}
	repoRoot := wt.Filesystem.Root()

	// Compute the path from the repo root to workingDir.
	rel, err := filepath.Rel(repoRoot, workingDir)
	if err != nil {
		return nil, fmt.Errorf("filepath.Rel: %w", err)
	}
	if !filepath.IsLocal(rel) {
		return nil, fmt.Errorf("%w: %q is outside the git repository %q", core.ErrRootOutsideProject, workingDir, repoRoot)
	}

	relSlash := filepath.ToSlash(rel)
	gitPath := pathpkg.Join(relSlash, filepath.ToSlash(path))

	commitAgainst, err := baselineCommit(repository, gitRef)
	if err != nil {
		return nil, fmt.Errorf("baselineCommit: %w", err)
	}

	treeAgainst, err := commitAgainst.Tree()
	if err != nil {
		return nil, fmt.Errorf("commitAgainst.Tree: %w", err)
	}

	gitTreeWalker := go_git.NewGitTreeWalker(treeAgainst, relSlash, gitPath)
	return gitTreeWalker, nil
}
