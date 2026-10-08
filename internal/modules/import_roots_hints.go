package modules

import (
	"fmt"
	"path"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func canonicalImportRootHints(roots []string) ([]string, error) {
	if len(roots) == 0 {
		return nil, nil
	}
	canonical := make([]string, 0, len(roots))
	for _, root := range roots {
		if root == "" || strings.HasPrefix(root, "/") || strings.ContainsAny(root, "\\:*?[]\x00") || slices.Contains(strings.Split(root, "/"), "..") {
			return nil, fmt.Errorf("invalid module-relative import root %q", root)
		}
		canonical = append(canonical, path.Clean(root))
	}
	slices.Sort(canonical)
	canonical = slices.Compact(canonical)
	if err := v1.ValidateModuleRoots(canonical); err != nil {
		return nil, fmt.Errorf("ValidateModuleRoots: %w", err)
	}
	return canonical, nil
}

func applyImportRootHint(module v1.Module, roots []string) (v1.Module, error) {
	if len(roots) == 0 {
		return module, nil
	}
	if module.RootsFromMetadata {
		declared := slices.Clone(module.Roots)
		slices.Sort(declared)
		if !slices.Equal(declared, roots) {
			return v1.Module{}, fmt.Errorf("module %s: import roots %v conflict with authoritative roots %v", module.Name, roots, module.Roots)
		}
		return module, nil
	}
	module.Roots = slices.Clone(roots)
	return module, nil
}
