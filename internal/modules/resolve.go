// Package modules coordinates v1 dependency resolution and project module files.
package modules

import (
	"context"
	"fmt"
	"slices"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Fetched contains metadata and a reproducible identity for one requested revision.
type Fetched struct {
	Module v1.Module
	Lock   v1.LockedModule
}

// Source supplies revision metadata without exposing checkout or cache layout.
type Source interface {
	Fetch(context.Context, string, string) (Fetched, error)
}

// Resolve selects the highest required semantic version, or an exact Git commit.
// It visits each revision once and preserves the existing pins for versionless requirements.
func Resolve(ctx context.Context, root v1.Module, source Source, pins map[string]v1.LockedModule) (v1.Lock, error) {
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{}}
	selected := make(map[string]string)
	seen := make(map[string]Fetched)
	resolvedHeads := make(map[string]string)
	queue := append([]v1.Requirement(nil), root.Requires...)
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return v1.Lock{}, err
		}
		requirement := queue[0]
		queue = queue[1:]
		var fetched Fetched
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
				fetched, err = source.Fetch(ctx, requirement.Module, "")
				if err != nil {
					return v1.Lock{}, err
				}
				requirement.Version = fetched.Lock.Version
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
			fetched, err = source.Fetch(ctx, requirement.Module, requirement.Version)
			if err != nil {
				return v1.Lock{}, err
			}
		}
		seen[key] = fetched
		queue = append(queue, fetched.Module.Requires...)
	}
	sources := make([]string, 0, len(selected))
	for source := range selected {
		sources = append(sources, source)
	}
	slices.Sort(sources)
	for _, source := range sources {
		lock.Modules = append(lock.Modules, seen[source+"@"+selected[source]].Lock)
	}
	return lock, nil
}
