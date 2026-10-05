package go_git

import (
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/easyp-tech/easyp/internal/core"
)

// baselineCommit resolves tags and commit revisions as well as branches.
// Existing bare branch names retain precedence over an identically named tag.
func baselineCommit(repository *gogit.Repository, ref string) (*object.Commit, error) {
	if strings.TrimSpace(ref) == "" {
		return nil, &core.GitRefNotFoundError{GitRef: ref}
	}
	revision := ref
	branch := plumbing.NewBranchReferenceName(ref)
	if _, err := repository.Reference(branch, true); err == nil {
		revision = branch.String()
	}
	hash, err := repository.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, fmt.Errorf("ResolveRevision: %w", &core.GitRefNotFoundError{GitRef: ref})
	}
	commit, err := repository.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("CommitObject: %w", err)
	}
	return commit, nil
}
