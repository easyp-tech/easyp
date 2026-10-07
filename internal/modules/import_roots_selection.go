package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const maxImportRootLayouts = 1024

type importRootSearch struct {
	index     *importRootIndex
	preserved map[string]string
	visited   map[string]bool
	solutions []importRootLayout
	failure   error
}

func selectImportRoots(ctx context.Context, modules []importRootModule) ([][]string, error) {
	selected := make([][]string, len(modules))
	for owner, module := range modules {
		if module.fixed || module.inspection == nil {
			continue
		}
		roots, err := selectModuleImportRoots(ctx, module)
		if err != nil {
			return nil, fmt.Errorf("module %s: %w", module.name, err)
		}
		selected[owner] = roots
	}
	return selected, nil
}

func selectModuleImportRoots(ctx context.Context, module importRootModule) ([]string, error) {
	// The candidate index is deliberately limited to one source module. A
	// consumer or another dependency can only validate this choice later.
	index, err := newImportRootIndex([]importRootModule{module})
	if err != nil {
		return nil, fmt.Errorf("newImportRootIndex: %w", err)
	}
	initial := importRootLayout{roots: make([][]string, 1)}
	namespace := index.namespace(initial)
	constraints := index.constraints(namespace)
	search := importRootSearch{index: index, preserved: constraints.bindings, visited: make(map[string]bool)}
	err = search.visit(ctx, initial)
	if err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}
	if len(search.solutions) == 0 {
		if search.failure != nil {
			return nil, search.failure
		}
		return nil, fmt.Errorf("no consistent import roots; supply explicit import roots")
	}
	if len(search.solutions) > 1 {
		return nil, fmt.Errorf("ambiguous import roots: %s or %s; supply explicit import roots", search.describe(search.solutions[0]), search.describe(search.solutions[1]))
	}
	return search.solutions[0].roots[0], nil
}

func (search *importRootSearch) visit(ctx context.Context, layout importRootLayout) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("Err: %w", err)
	}
	if len(search.solutions) > 1 {
		return nil
	}
	signature := search.describe(layout)
	if search.visited[signature] {
		return nil
	}
	if len(search.visited) >= maxImportRootLayouts {
		return fmt.Errorf("ambiguous import root search exceeds %d layouts; supply explicit import roots", maxImportRootLayouts)
	}
	search.visited[signature] = true
	namespace := search.index.namespace(layout)
	constraints := search.index.constraints(namespace)
	for name, identity := range search.preserved {
		file := namespace.files[name]
		if file == nil || file.Identity != identity {
			search.failure = fmt.Errorf("import %q would change its previously resolved source; supply compatible explicit import roots", name)
			return nil
		}
	}
	if len(constraints.missing) == 0 {
		for _, file := range layout.intrinsicOwners {
			selected := false
			for candidate := range namespace.selected {
				if candidate.Identity == file.Identity {
					selected = true
					break
				}
			}
			if !selected {
				search.failure = fmt.Errorf("import roots would omit intrinsic import source %s; supply explicit import roots", file.Path)
				return nil
			}
		}
		if namespace.err != nil {
			search.failure = namespace.err
			return nil
		}
		search.solutions = append(search.solutions, layout)
		return nil
	}
	missing := constraints.missing[0]
	candidates := search.index.candidates[missing.path]
	if len(candidates) == 0 {
		search.failure = fmt.Errorf("cannot resolve imports [%s imports %q] in declared modules", missing.owner, missing.path)
		return nil
	}
	for _, candidate := range candidates {
		roots := layout.roots[candidate.module]
		if slices.Contains(roots, candidate.root) {
			continue
		}
		next := importRootLayout{roots: slices.Clone(layout.roots), intrinsicOwners: slices.Clone(layout.intrinsicOwners)}
		next.roots[candidate.module] = append(slices.Clone(roots), candidate.root)
		slices.Sort(next.roots[candidate.module])
		if missing.intrinsic != nil && !slices.Contains(next.intrinsicOwners, missing.intrinsic) {
			next.intrinsicOwners = append(next.intrinsicOwners, missing.intrinsic)
		}
		if err := search.visit(ctx, next); err != nil {
			return err
		}
	}
	return nil
}

func (search *importRootSearch) describe(layout importRootLayout) string {
	var parts []string
	for owner, roots := range layout.roots {
		if len(roots) > 0 {
			parts = append(parts, fmt.Sprintf("%s roots [%s]", search.index.modules[owner].name, strings.Join(roots, ", ")))
		}
	}
	if len(parts) == 0 {
		return "default roots [.]"
	}
	return strings.Join(parts, "; ")
}

// ResolveIntrinsicImportRoots derives roots only from imports of inspected files
// in this module. A nil result retains its current default namespace. Unused
// malformed message bodies and unknown imports do not become validation errors.
func ResolveIntrinsicImportRoots(ctx context.Context, module v1.Module, inspection *RootInspection) ([]string, error) {
	if module.RootsFromMetadata || inspection == nil {
		return nil, nil
	}
	roots, err := selectImportRoots(ctx, []importRootModule{{name: module.Name, roots: module.Roots, inspection: inspection}})
	if err != nil {
		return nil, fmt.Errorf("selectImportRoots: %w", err)
	}
	return roots[0], nil
}
