package api

import (
	"fmt"
	"path/filepath"
)

// breakingIgnorePaths binds ignores to their owning policy, not the Git root.
func breakingIgnorePaths(repositoryRoot, policyRoot string, ignores []string) ([]string, error) {
	result := make([]string, 0, len(ignores))
	for _, ignored := range ignores {
		if ignored == "" || !filepath.IsLocal(ignored) {
			return nil, fmt.Errorf("breaking.ignore requires a nonempty policy-relative path: %q", ignored)
		}
		rel, err := filepath.Rel(repositoryRoot, filepath.Join(policyRoot, ignored))
		if err != nil || !filepath.IsLocal(rel) {
			return nil, fmt.Errorf("breaking.ignore path leaves the repository: %q", ignored)
		}
		result = append(result, filepath.ToSlash(rel))
	}
	return result, nil
}
