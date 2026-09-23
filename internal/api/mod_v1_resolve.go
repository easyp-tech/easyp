package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
	"golang.org/x/mod/sumdb/dirhash"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type fetchedV1Module struct {
	config v1.Module
	lock   v1.LockedModule
}

// buildV1Lock traverses requirements and selects the highest minimum version
// of each module, then locks the exact Git commit and Go-style content hash.
func buildV1Lock(ctx context.Context, root v1.Module, cacheRoot string) (v1.Lock, error) {
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{}}
	selected := make(map[string]string)
	seen := make(map[string]fetchedV1Module)
	queue := append([]v1.Requirement(nil), root.Requires...)
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return v1.Lock{}, err
		}
		requirement := queue[0]
		queue = queue[1:]
		if requirement.Version == "" {
			version, err := latestV1Tag(ctx, requirement.Module)
			if err != nil {
				return v1.Lock{}, err
			}
			requirement.Version = version
		}
		if !semver.IsValid(requirement.Version) {
			return v1.Lock{}, fmt.Errorf("require %s: invalid semantic version %q", requirement.Module, requirement.Version)
		}
		if current, ok := selected[requirement.Module]; !ok || semver.Compare(requirement.Version, current) > 0 {
			selected[requirement.Module] = requirement.Version
		}
		key := requirement.Module + "@" + requirement.Version
		if _, ok := seen[key]; ok {
			continue
		}
		fetched, err := fetchV1Module(ctx, requirement.Module, requirement.Version, cacheRoot)
		if err != nil {
			return v1.Lock{}, err
		}
		seen[key] = fetched
		queue = append(queue, fetched.config.Requires...)
	}
	sources := make([]string, 0, len(selected))
	for source := range selected {
		sources = append(sources, source)
	}
	slices.Sort(sources)
	for _, source := range sources {
		lock.Modules = append(lock.Modules, seen[source+"@"+selected[source]].lock)
	}
	return lock, nil
}

func fetchV1Module(ctx context.Context, source, version, cacheRoot string) (fetchedV1Module, error) {
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return fetchedV1Module{}, err
	}
	checkout, err := os.MkdirTemp(cacheRoot, "git-*")
	if err != nil {
		return fetchedV1Module{}, err
	}
	defer os.RemoveAll(checkout)
	remote := v1GitRemote(source)
	if _, err := gitV1(ctx, "", "clone", "--quiet", "--depth=1", "--branch", version, "--no-checkout", "--", remote, checkout); err != nil {
		return fetchedV1Module{}, fmt.Errorf("clone %s@%s: %w", source, version, err)
	}
	commit, err := gitV1(ctx, checkout, "rev-parse", "--verify", "refs/tags/"+version+"^{commit}")
	if err != nil {
		return fetchedV1Module{}, fmt.Errorf("%s: %s is not a Git tag: %w", source, version, err)
	}
	commit = strings.TrimSpace(commit)
	if _, err := gitV1(ctx, checkout, "checkout", "--quiet", "--detach", commit); err != nil {
		return fetchedV1Module{}, err
	}
	module, err := moduleconfig.ReadGitDependency(checkout, source)
	if err != nil {
		return fetchedV1Module{}, fmt.Errorf("%s@%s: %w", source, version, err)
	}
	filesRaw, err := gitV1(ctx, checkout, "ls-files", "-z")
	if err != nil {
		return fetchedV1Module{}, err
	}
	var files []string
	for _, name := range strings.Split(filesRaw, "\x00") {
		if name == "" {
			continue
		}
		path := filepath.Join(checkout, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if err != nil {
			return fetchedV1Module{}, err
		}
		if !info.Mode().IsRegular() {
			return fetchedV1Module{}, fmt.Errorf("%s@%s: unsupported non-regular file %q", source, version, name)
		}
		files = append(files, name)
	}
	hash, err := dirhash.Hash1(files, func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(checkout, filepath.FromSlash(name)))
	})
	if err != nil {
		return fetchedV1Module{}, fmt.Errorf("hash %s@%s: %w", source, version, err)
	}
	return fetchedV1Module{config: module, lock: v1.LockedModule{
		Source: source, Version: version, Commit: commit, Hash: hash,
	}}, nil
}

func gitV1(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func v1GitRemote(source string) string {
	if filepath.IsAbs(source) || strings.Contains(source, "://") {
		return source
	}
	return "https://" + source
}

func v1CacheSourceKey(source string) string {
	hash := sha256.Sum256([]byte(source))
	return fmt.Sprintf("%x", hash)
}
