package modules

import (
	"fmt"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// manifestRequirementVersions keeps an inferred Git HEAD out of protobuf.mod.
// Exact commits declared by a dependency remain explicit requirements.
func manifestRequirementVersions(lock v1.Lock, repository Cache) (map[string]string, error) {
	explicitCommits := make(map[string]bool)
	for _, entry := range lock.Modules {
		_, dependency, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("cached %s: %w", entry.Source, err)
		}
		for _, requirement := range dependency.Requires {
			if v1.IsCommitRef(requirement.Version) {
				explicitCommits[requirement.Module] = true
			}
		}
	}
	versions := make(map[string]string, len(lock.Modules))
	for _, entry := range lock.Modules {
		version := entry.Version
		if v1.IsCommitRef(version) && !explicitCommits[entry.Source] {
			version = ""
		}
		versions[entry.Source] = version
	}
	return versions, nil
}
