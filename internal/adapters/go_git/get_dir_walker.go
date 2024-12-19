package go_git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/go_git"
)

func (g *GoGit) GetDirWalker(workingDir, gitRef, path string) (core.DirWalker, error) {
	repository, err := git.PlainOpen(workingDir)
	if err != nil {
		return nil, fmt.Errorf("git.PlainOpen: %w", err)
	}
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", gitRef))

	refAgainst, err := repository.Reference(refName, false)
	if err != nil {
		return nil, &core.GitRefNotFoundError{GitRef: gitRef}
	}

	commitAgainst, err := repository.CommitObject(refAgainst.Hash())
	if err != nil {
		return nil, fmt.Errorf("repository.CommitObject: %w", err)
	}

	treeAgainst, err := commitAgainst.Tree()
	if err != nil {
		return nil, fmt.Errorf("commitAgainst.Tree: %w", err)
	}

	gitTreeWalker := go_git.NewGitTreeWalker(treeAgainst, path)
	return gitTreeWalker, nil
}
