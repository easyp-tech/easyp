package gitmodules

import (
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func applyV1ModuleRoots(module v1.Module, roots []string) (v1.Module, error) {
	if err := v1.ValidateModuleRoots(roots); err != nil {
		return v1.Module{}, fmt.Errorf("ValidateModuleRoots: %w", err)
	}
	if len(roots) == 0 {
		return module, nil
	}
	selected := slices.Clone(roots)
	slices.Sort(selected)
	if module.RootsFromMetadata {
		authoritative := slices.Clone(module.Roots)
		slices.Sort(authoritative)
		authoritative = slices.Compact(authoritative)
		if !slices.Equal(selected, authoritative) {
			return v1.Module{}, fmt.Errorf("module %s has authoritative roots %v; cannot select %v", module.Name, module.Roots, roots)
		}
		return module, nil
	}
	module.Roots = selected
	return module, nil
}

func lockedV1ModuleRoots(module v1.Module, roots []string) []string {
	if module.RootsFromMetadata || len(roots) == 0 {
		return nil
	}
	return slices.Clone(module.Roots)
}

func v1RootSelectionKey(roots []string) string {
	if len(roots) == 0 {
		return ""
	}
	canonical := slices.Clone(roots)
	slices.Sort(canonical)
	return strings.Join(canonical, "\n")
}
