package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func finalizeEffectiveImportRoots(ctx context.Context, pass resolutionPass, locals *localReplacements, remote *effectiveRemoteSource) (v1.Lock, error) {
	var candidates []importRootModule
	for _, name := range pass.locals {
		entry := locals.modules[name]
		candidate, err := inspectLocalRootCandidates(ctx, entry.Directory, entry.Module, len(locals.hints[name]) > 0)
		if err != nil {
			return v1.Lock{}, fmt.Errorf("inspectLocalRootCandidates: %s: %w", name, err)
		}
		candidates = append(candidates, candidate)
	}
	var dependencies []importRootModule
	for _, locked := range pass.lock.Modules {
		fetched := remote.roots.fetched[[2]string{locked.Source, strings.ToLower(locked.Commit)}]
		candidate := importRootModule{name: locked.Source, roots: fetched.Module.Roots, inspection: fetched.Inspection}
		candidate.fixed = remote.roots.fixedRootScope(locked, fetched)
		if fetched.Inspection == nil {
			// Published pins and repositories without inspections retain
			// their ordinary verified cache path and fixed namespace.
			candidate.fixed = true
		}
		dependencies = append(dependencies, candidate)
		candidates = append(candidates, candidate)
	}
	if !slices.ContainsFunc(candidates, func(candidate importRootModule) bool { return !candidate.fixed }) {
		return remote.roots.finalizeSelections(ctx, pass.lock, dependencies, make([][]string, len(dependencies)))
	}
	selected, err := selectImportRoots(ctx, candidates)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("selectImportRoots: %w", err)
	}
	for offset, name := range pass.locals {
		if len(selected[offset]) == 0 {
			continue
		}
		entry := locals.modules[name]
		entry.Module.Roots = slices.Clone(selected[offset])
		locals.modules[name] = entry
	}
	return remote.roots.finalizeSelections(ctx, pass.lock, dependencies, selected[len(pass.locals):])
}
