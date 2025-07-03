package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/repository"
)

var _ repository.Repo = (*gitRepo)(nil)

// gitRepo implements repository.Repo interface
type gitRepo struct {
	// remoteURL full repository remoteURL address with schema
	remoteURL string
	// cacheDir local cache directory for store repository
	cacheDir string
	// console for call external commands
	console Console
}

const (
	// for omitted package version. HEAD is git key word.
	gitLatestVersionRef = "HEAD"
	// tag prefix on output of ls-remote command
	gitRefsTagPrefix = "refs/tags/"
)

// Some links from go mod:
// cmd/go/internal/modfetch/codehost/git.go:65 - create work dir
// cmd/go/internal/modfetch/codehost/git.go:137 - git's struct

// Console temporary interface for console commands, must be replaced from core.Console.
type Console interface {
	RunCmd(ctx context.Context, dir string, command string, commandParams ...string) (string, error)
}

// New returns gitRepo instance
// remote: full remoteURL address without schema
func New(ctx context.Context, remote string, cacheDir string, console Console) (repository.Repo, error) {
	r := &gitRepo{
		remoteURL: getRemote(remote),
		cacheDir:  cacheDir,
		console:   console,
	}

	if _, err := os.Stat(filepath.Join(r.cacheDir, "objects")); err == nil {
		// repo is already exists
		return r, nil
	}

	if _, err := r.console.RunCmd(ctx, r.cacheDir, "git", "init", "--bare"); err != nil {
		return nil, fmt.Errorf("adapters.RunCmd (init): %w", err)
	}

	_, err := r.console.RunCmd(ctx, r.cacheDir, "git", "remote", "add", "origin", r.remoteURL)
	if err != nil {
		return nil, fmt.Errorf("adapters.RunCmd (add origin): %w", err)
	}

	return r, nil
}

func getRemote(name string) string {
	return "https://" + name
}

// getCommitDatetime returns datetime of commit
// NOTE: the commit has to be fetched!
func (r *gitRepo) getCommitDatetime(ctx context.Context, commitHash string) (string, error) {
	var lines []string

	commitDatetime, err := r.console.RunCmd(
		ctx,
		r.cacheDir,
		"git",
		"log", "-1",
		"--pretty=%ad", "--date=format:%Y%m%d%H%M%S",
		commitHash,
	)
	if err != nil {
		return "", fmt.Errorf("r.console.RunCmd: %w", err)
	}

	// got commit hash from result
	lines = strings.Split(commitDatetime, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("invalid lines of git log: %s", commitDatetime)
	}
	parts := strings.Fields(lines[0])
	if len(parts) != 1 {
		return "", fmt.Errorf("invalid parts of git log: %s", commitDatetime)
	}

	return parts[0], nil
}

func (r *gitRepo) fetchCommit(ctx context.Context, commitHash string) error {
	_, err := r.console.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", commitHash,
	)
	if err != nil {
		return fmt.Errorf("r.console.RunCmd: %w", err)
	}

	return nil
}
