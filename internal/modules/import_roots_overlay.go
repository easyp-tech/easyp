package modules

import (
	"context"
	"fmt"
	"slices"

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
	// Published pins and repositories without inspections keep their fixed
	// namespace through the ordinary verified cache path.
	dependencies := remote.roots.rootCandidates(pass.lock)
	candidates = append(candidates, dependencies...)
	selected := make([][]string, len(dependencies))
	if slices.ContainsFunc(candidates, func(candidate importRootModule) bool { return !candidate.fixed }) {
		choices, err := selectImportRoots(ctx, candidates)
		if err != nil {
			return v1.Lock{}, fmt.Errorf("selectImportRoots: %w", err)
		}
		for offset, name := range pass.locals {
			if len(choices[offset]) == 0 {
				continue
			}
			entry := locals.modules[name]
			entry.Module.Roots = slices.Clone(choices[offset])
			locals.modules[name] = entry
		}
		selected = choices[len(pass.locals):]
	}
	lock, err := remote.roots.finalizeSelections(ctx, pass.lock, dependencies, selected)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("finalizeSelections: %w", err)
	}
	err = remote.roots.validateRootTransitions(ctx)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("validateRootTransitions: %w", err)
	}
	return lock, nil
}
