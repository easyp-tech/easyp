package git

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/easyp-tech/easyp/internal/mod/adapters"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

type revisionParts struct {
	CommitHash string
	Version    string
}

// ReadRevision read actual revision from remote repository
// Cases:
//  1. requestedVersion is git tag: just get commit with this tag
//  2. requestedVersion is omitted: get the latest commit, try to read its tag
//     if tag does not exist generate version
//  3. requestedVersion is generated: get commit from its version
func (r *gitRepo) ReadRevision(ctx context.Context, requestedVersion models.RequestedVersion) (models.Revision, error) {
	var revParts revisionParts
	var err error

	switch {
	case requestedVersion.IsGenerated():
		revParts, err = r.readRevisionByGeneratedVersion(ctx, requestedVersion)
		if err != nil {
			return models.Revision{}, fmt.Errorf("r.readRevisionByGeneratedVersion: %w", err)
		}
	case requestedVersion.IsOmitted():
		revParts, err = r.readRevisionForLatestCommit(ctx)
		if err != nil {
			return models.Revision{}, fmt.Errorf("r.readRevisionForLatestCommit: %w", err)
		}
	default:
		// in other case use readRevisionByGitTagVersion
		revParts, err = r.readRevisionByGitTagVersion(ctx, requestedVersion)
		if err != nil {
			return models.Revision{}, fmt.Errorf("r.readRevisionByGitTagVersion: %w", err)
		}
	}

	if revParts.CommitHash == "" {
		return models.Revision{}, models.ErrVersionNotFound
	}

	revision := models.Revision{
		CommitHash: revParts.CommitHash,
		Version:    revParts.Version,
	}
	slog.Debug("Revision", "value", revision)

	return revision, nil
}

// readRevisionByTag read revision by passed git tag
// tag has to be on the remote repository
func (r *gitRepo) readRevisionByGitTagVersion(
	ctx context.Context, requestedVersion models.RequestedVersion,
) (revisionParts, error) {
	gitTagVersion := string(requestedVersion)

	res, err := adapters.RunCmd(ctx, r.cacheDir, "git", "ls-remote", "origin", gitTagVersion)
	if err != nil {
		return revisionParts{}, models.ErrVersionNotFound
	}

	commitHash := ""

	for _, lsOut := range strings.Split(res, "\n") {
		rev := strings.Fields(lsOut)
		if len(rev) != 2 {
			continue
		}

		if strings.HasPrefix(rev[1], gitRefsTagPrefix) &&
			strings.TrimPrefix(rev[1], gitRefsTagPrefix) == gitTagVersion {
			commitHash = rev[0]
			break
		}
	}

	parts := revisionParts{
		CommitHash: commitHash,
		Version:    gitTagVersion,
	}
	return parts, nil
}

// readRevisionForLatestCommit read the latest commit
// if tag for this commit exists use its as revision's version
// otherwise generate version
func (r *gitRepo) readRevisionForLatestCommit(
	ctx context.Context,
) (revisionParts, error) {
	headInfo, err := adapters.RunCmd(
		ctx, r.cacheDir, "git", "ls-remote", "origin", gitLatestVersionRef,
	)
	if err != nil {
		return revisionParts{}, models.ErrVersionNotFound
	}

	// got commit hash from result
	lines := strings.Split(headInfo, "\n")
	if len(lines) == 0 {
		return revisionParts{}, fmt.Errorf("invalid lines of git info: %s", headInfo)
	}
	parts := strings.Fields(lines[0])
	if len(parts) != 2 {
		return revisionParts{}, fmt.Errorf("invalid parts of git info: %s", headInfo)
	}

	commitHash := parts[0]
	version := ""

	// try to get git tag for this commit
	tagInfo, err := adapters.RunCmd(ctx, r.cacheDir, "git", "ls-remote", "origin")
	if err != nil {
		return revisionParts{}, fmt.Errorf("adapters.RunCmd (ls-remote tagInfo): %w", err)
	}

	for _, lsOut := range strings.Split(tagInfo, "\n") {
		rev := strings.Fields(lsOut)
		if len(rev) != 2 {
			continue
		}

		if rev[0] != commitHash {
			continue
		}

		if strings.HasPrefix(rev[1], gitRefsTagPrefix) {
			version = strings.TrimPrefix(rev[1], gitRefsTagPrefix)
			break
		}
	}

	if version != "" {
		// version was found. return it
		revParts := revisionParts{
			CommitHash: commitHash,
			Version:    version,
		}
		return revParts, nil
	}
	// didn't find tag for this commit, so generate version

	// fetch commit by its hash
	_, err = adapters.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", commitHash,
	)
	if err != nil {
		return revisionParts{}, fmt.Errorf("adapters.RunCmd (fetch): %w", err)
	}

	// get commit's datetime
	commitDatetime, err := adapters.RunCmd(
		ctx,
		r.cacheDir,
		"git",
		"log", "-1",
		"--pretty=%ad", "--date=format:%Y%m%d%H%M%S",
		commitHash,
	)
	if err != nil {
		return revisionParts{}, fmt.Errorf("adapters.RunCmd (log): %w", err)
	}

	// got commit hash from result
	lines = strings.Split(commitDatetime, "\n")
	if len(lines) == 0 {
		return revisionParts{}, fmt.Errorf("invalid lines of git log: %s", commitDatetime)
	}
	parts = strings.Fields(lines[0])
	if len(parts) != 1 {
		return revisionParts{}, fmt.Errorf("invalid parts of git log: %s", commitDatetime)
	}

	generatedVersion := models.GeneratedVersionParts{
		Datetime:   parts[0],
		CommitHash: commitHash,
	}

	revParts := revisionParts{
		CommitHash: commitHash,
		Version:    generatedVersion.GetVersionString(),
	}
	return revParts, nil
}

// readRevisionByGeneratedVersion check if commit in generated version exists
func (r *gitRepo) readRevisionByGeneratedVersion(
	ctx context.Context, requestedVersion models.RequestedVersion,
) (revisionParts, error) {
	generatedParts, err := requestedVersion.GetParts()
	if err != nil {
		return revisionParts{}, fmt.Errorf("requestedVersion.GetParts: %w", err)
	}

	// fetch by passed commit hash
	_, err = adapters.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", generatedParts.CommitHash,
	)
	if err != nil {
		// do not parse error, just return as commit was not found
		return revisionParts{}, nil
	}

	parts := revisionParts{
		CommitHash: generatedParts.CommitHash,
		Version:    string(requestedVersion),
	}
	return parts, nil
}
