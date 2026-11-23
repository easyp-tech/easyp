package core

import (
	"context"
	"fmt"
	"io"
)

var ErrRootOutsideProject = fmt.Errorf("breaking check root must be inside the git repository")

type BreakingCheckConfig struct {
	// IgnoreDirs lists directories excluded from breaking comparison.
	IgnoreDirs []string
	// Categories selects descriptor-backed compatibility profiles. Empty defaults to FILE.
	Categories []string
	// IgnoreUnstable skips packages whose final component is an unstable version.
	IgnoreUnstable bool
}

// CompareBreaking compares explicitly scoped inputs with revision-specific imports.
// Neither filesystem can fall back to the other revision's dependency roots.
func (c *Core) CompareBreaking(ctx context.Context, current, against DirWalker, againstImportRoots []string) ([]IssueInfo, error) {
	return c.CompareBreakingWithImports(ctx, current, against, BreakingImports{Paths: againstImportRoots})
}

// BreakingImports binds one revision's import paths to its filter and opener.
type BreakingImports struct {
	Paths       []string
	FileAllowed func(string) bool
	Open        func(string) (io.ReadCloser, error)
}

// CompareBreakingWithImports keeps source access isolated between revisions.
func (c *Core) CompareBreakingWithImports(ctx context.Context, current, against DirWalker, imports BreakingImports) ([]IssueInfo, error) {
	baseline := *c
	baseline.importRoots = imports.Paths
	baseline.importFileAllowed = imports.FileAllowed
	baseline.sourceFileOpen = imports.Open
	categories := c.breakingCheckConfig.Categories
	if len(categories) == 0 {
		categories = []string{"FILE"}
	}
	profiles, err := selectedBreakingProfiles(categories)
	if err != nil {
		return nil, err
	}
	return c.compareBreakingProfiles(ctx, current, against, &baseline, profiles)
}
