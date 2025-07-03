package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/models"
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
		revParts, err = r.readRevisionByVersion(ctx, requestedVersion)
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

// readRevisionByVersion read revision by passed version
// order of resolving:
// 1. try to read as git tag
// 2. try to read as commit's hash
// 3. return not found error
func (r *gitRepo) readRevisionByVersion(
	ctx context.Context, requestedVersion models.RequestedVersion,
) (revisionParts, error) {
	revParts, err := r.readRevisionByGitTagVersion(ctx, requestedVersion)
	if err == nil {
		return revParts, nil
	}

	if !errors.Is(err, models.ErrVersionNotFound) {
		return revisionParts{}, fmt.Errorf("r.readRevisionByGitTagVersion: %w", err)
	}

	revParts, err = r.readRevisionByCommitHash(ctx, string(requestedVersion))
	if err != nil {
		return revisionParts{}, fmt.Errorf("r.readRevisionByCommitHash: %w", err)
	}

	return revParts, nil
}

// readRevisionByGitTagVersion read revision by passed git tag
// tag has to be on the remote repository
func (r *gitRepo) readRevisionByGitTagVersion(
	ctx context.Context, requestedVersion models.RequestedVersion,
) (revisionParts, error) {
	gitTagVersion := string(requestedVersion)

	res, err := r.lsRemote(ctx, gitTagVersion)
	if err != nil {
		return revisionParts{}, models.ErrVersionNotFound
	}

	commitHash := ""

	for _, lsOut := range res {
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

	if commitHash == "" {
		return revisionParts{}, models.ErrVersionNotFound
	}

	parts := revisionParts{
		CommitHash: commitHash,
		Version:    gitTagVersion,
	}

	return parts, nil
}

// readRevisionByCommitHash read revision by passed hash of commit
func (r *gitRepo) readRevisionByCommitHash(
	ctx context.Context, commitHash string,
) (revisionParts, error) {
	// try to fetch commit
	if err := r.fetchCommit(ctx, commitHash); err != nil {
		return revisionParts{}, fmt.Errorf("r.fetchCommit: %w", models.ErrVersionNotFound)
	}

	gitTag, err := r.getTagByCommit(ctx, commitHash)
	if err != nil {
		return revisionParts{}, fmt.Errorf("r.getTagByCommit: %w", err)
	}

	if gitTag != "" {
		return revisionParts{
			CommitHash: commitHash,
			Version:    gitTag,
		}, nil
	}

	// didn't find tag for this commit, so generate version

	commitDatetime, err := r.getCommitDatetime(ctx, commitHash)
	if err != nil {
		return revisionParts{}, fmt.Errorf("r.getCommitDatetime: %w", err)
	}

	generatedVersion := models.GeneratedVersionParts{
		Datetime:   commitDatetime,
		CommitHash: commitHash,
	}

	revParts := revisionParts{
		CommitHash: commitHash,
		Version:    generatedVersion.GetVersionString(),
	}
	return revParts, nil
}

// readRevisionForLatestCommit read the latest commit
// if tag for this commit exists use its as revision's version
// otherwise generate version
func (r *gitRepo) readRevisionForLatestCommit(
	ctx context.Context,
) (revisionParts, error) {
	lines, err := r.lsRemote(ctx, gitLatestVersionRef)
	if err != nil {
		return revisionParts{}, models.ErrVersionNotFound
	}

	// got commit hash from result
	if len(lines) == 0 {
		return revisionParts{}, fmt.Errorf("invalid lines of git info: %s", lines)
	}
	parts := strings.Fields(lines[0])
	if len(parts) != 2 {
		return revisionParts{}, fmt.Errorf("invalid parts of git info: %s", lines)
	}

	headCommitHash := parts[0]

	revParts, err := r.readRevisionByCommitHash(ctx, headCommitHash)
	if err != nil {
		return revisionParts{}, fmt.Errorf("r.readRevisionByCommitHash: %w", err)
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
	if err := r.fetchCommit(ctx, generatedParts.CommitHash); err != nil {
		return revisionParts{}, fmt.Errorf("r.fetchCommit: %w", err)
	}

	parts := revisionParts{
		CommitHash: generatedParts.CommitHash,
		Version:    string(requestedVersion),
	}
	return parts, nil
}
