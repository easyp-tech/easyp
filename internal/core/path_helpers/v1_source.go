package path_helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// ShouldSkipV1SourceDir reports whether a directory under a v1 module root
// belongs to a nested module or to hidden local state rather than its sources.
func ShouldSkipV1SourceDir(root, path string) bool {
	if filepath.Clean(path) == filepath.Clean(root) {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	if HiddenOrVendorSourcePath(root, path) {
		return true
	}
	for _, name := range []string{".git", "protobuf.mod", "buf.work.yaml", "buf.yaml"} {
		if _, err := os.Stat(filepath.Join(path, name)); err == nil {
			return true
		}
	}
	return false
}

// HiddenOrVendorSourcePath reports lexical hidden and vendored descendants.
// It does not inspect metadata ownership or resolve aliases.
func HiddenOrVendorSourcePath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(relative) {
		return false
	}
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part != "." && (strings.HasPrefix(part, ".") || part == "easyp_vendor") {
			return true
		}
	}
	return false
}
