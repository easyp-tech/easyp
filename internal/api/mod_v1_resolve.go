package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type fetchedV1Module struct {
	config v1.Module
	lock   v1.LockedModule
}

// buildV1Lock traverses requirements, selects SemVer minima or Git HEAD for
// versionless modules, then locks the exact commit and Go-style content hash.
func buildV1Lock(ctx context.Context, root v1.Module, cacheRoot string) (v1.Lock, error) {
	return buildV1LockWithPins(ctx, root, cacheRoot, nil)
}

func buildV1LockWithPins(ctx context.Context, root v1.Module, cacheRoot string, pins map[string]v1.LockedModule) (v1.Lock, error) {
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{}}
	selected := make(map[string]string)
	seen := make(map[string]fetchedV1Module)
	resolvedHeads := make(map[string]string)
	queue := append([]v1.Requirement(nil), root.Requires...)
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return v1.Lock{}, err
		}
		requirement := queue[0]
		queue = queue[1:]
		var fetched fetchedV1Module
		prefetched := false
		if requirement.Version == "" {
			switch {
			case resolvedHeads[requirement.Module] != "":
				requirement.Version = resolvedHeads[requirement.Module]
			case pins[requirement.Module].Commit != "":
				requirement.Version = pins[requirement.Module].Commit
				resolvedHeads[requirement.Module] = requirement.Version
			default:
				var err error
				fetched, err = fetchV1Module(ctx, requirement.Module, "", cacheRoot)
				if err != nil {
					return v1.Lock{}, err
				}
				requirement.Version = fetched.lock.Version
				resolvedHeads[requirement.Module] = requirement.Version
				prefetched = true
			}
		}
		if !semver.IsValid(requirement.Version) && !v1.IsCommitRef(requirement.Version) {
			return v1.Lock{}, fmt.Errorf("require %s: expected semantic version or full Git commit, got %q", requirement.Module, requirement.Version)
		}
		if current, ok := selected[requirement.Module]; ok && (v1.IsCommitRef(current) || v1.IsCommitRef(requirement.Version)) && current != requirement.Version {
			return v1.Lock{}, fmt.Errorf("module %s has conflicting requirements %s and %s", requirement.Module, current, requirement.Version)
		}
		if current, ok := selected[requirement.Module]; !ok || (!v1.IsCommitRef(current) && semver.Compare(requirement.Version, current) > 0) {
			selected[requirement.Module] = requirement.Version
		}
		key := requirement.Module + "@" + requirement.Version
		if _, ok := seen[key]; ok {
			continue
		}
		if !prefetched {
			var err error
			fetched, err = fetchV1Module(ctx, requirement.Module, requirement.Version, cacheRoot)
			if err != nil {
				return v1.Lock{}, err
			}
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

func v1CacheSourceKey(source string) string {
	hash := sha256.Sum256([]byte(source))
	return fmt.Sprintf("%x", hash)
}
