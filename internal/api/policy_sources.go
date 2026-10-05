package api

import (
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/modules"
)

// policySourceExcluded distinguishes checked sources from storage and Git metadata.
// Exclusion happens before configuration discovery; imports remain independently resolved.
func policySourceExcluded(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(relative) {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		if part != "." && (part == modules.VendorDir || strings.HasPrefix(part, ".")) {
			return true
		}
	}
	return false
}
