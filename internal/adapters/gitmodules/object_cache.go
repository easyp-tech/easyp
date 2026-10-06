package gitmodules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// checkoutCachedCommit fetches one snapshot into a reusable bare object store.
// Temporary checkouts borrow only local cached objects; no second network clone
// is necessary when Fetch is followed by Install or the same commit is reused.
func checkoutCachedCommit(ctx context.Context, checkout string, entry v1.LockedModule, candidate v1GitModuleCandidate) (bool, error) {
	root := filepath.Join(filepath.Dir(checkout), "objects")
	if err := os.MkdirAll(root, 0o700); err != nil {
		return false, fmt.Errorf("MkdirAll: %w", err)
	}
	repository := filepath.Join(root, v1CacheSourceKey(candidate.remote))
	unlock, err := lockObjectRepository(ctx, repository+".lock")
	if err != nil {
		return false, fmt.Errorf("lockObjectRepository: %w", err)
	}
	defer unlock()
	if _, err := os.Stat(filepath.Join(repository, "HEAD")); os.IsNotExist(err) {
		args := []string{"init", "--quiet", "--bare"}
		if len(entry.Commit) == 64 {
			args = append(args, "--object-format=sha256")
		}
		args = append(args, repository)
		if _, err := gitV1(ctx, "", args...); err != nil {
			return false, err
		}
		if _, err := gitV1(ctx, repository, "config", "gc.auto", "0"); err != nil {
			return false, err
		}
	} else if err != nil {
		return false, fmt.Errorf("Stat: %w", err)
	}
	commit := strings.ToLower(entry.Commit)
	if _, err := gitV1(ctx, repository, "cat-file", "-e", commit+"^{commit}"); err != nil {
		// Depth-one fetch works even for a historical commit when the server permits
		// requesting its object ID. Do not silently replace an unavailable pin by HEAD.
		_, fetchErr := gitV1(ctx, repository, "-c", "fetch.fsckObjects=true", "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--depth=1", "--", candidate.remote, commit)
		if fetchErr != nil {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			// Some servers only expose advertised refs. Fetch reachable history as a
			// compatibility fallback, then verify the exact requested commit again.
			args := []string{"-c", "fetch.fsckObjects=true", "fetch", "--quiet", "--no-tags", "--no-write-fetch-head"}
			if _, err := os.Stat(filepath.Join(repository, "shallow")); err == nil {
				args = append(args, "--unshallow")
			}
			args = append(args, "--", candidate.remote, "+refs/heads/*:refs/remotes/easyp/*", "+refs/tags/*:refs/tags/*")
			if _, err := gitV1(ctx, repository, args...); err != nil {
				return false, fmt.Errorf("fetch pinned commit (shallow: %v; fallback: %w)", fetchErr, err)
			}
		}
	}
	resolved, err := gitV1(ctx, repository, "rev-parse", "--verify", commit+"^{commit}")
	if err != nil {
		return true, fmt.Errorf("locked commit %s is unavailable in the fetched repository history; restore access to the pinned revision rather than replacing it with HEAD: %w", commit, err)
	}
	if strings.TrimSpace(resolved) != commit {
		return true, fmt.Errorf("resolved commit does not match %s", commit)
	}
	if _, err := gitV1(ctx, repository, "update-ref", "refs/easyp/"+commit, commit); err != nil {
		return true, err
	}
	if _, err := gitV1(ctx, repository, "update-ref", "refs/heads/easyp-cache", commit); err != nil {
		return true, err
	}
	if _, err := gitV1(ctx, repository, "symbolic-ref", "HEAD", "refs/heads/easyp-cache"); err != nil {
		return true, err
	}
	if _, err := gitV1(ctx, "", "clone", "--quiet", "--shared", "--no-checkout", "--", repository, checkout); err != nil {
		return true, fmt.Errorf("clone cached objects: %w", err)
	}
	if _, err := gitV1(ctx, checkout, "read-tree", commit); err != nil {
		return true, err
	}
	return true, nil
}

// OS locks are released on process exit and shared across concurrent CLI runs.
func lockObjectRepository(ctx context.Context, path string) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	for {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		acquired, err := tryObjectLock(file)
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		if acquired {
			return func() { _ = releaseObjectLock(file); _ = file.Close() }, nil
		}
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = file.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
