package modules

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// Replacement validation shares tidy's reachable resolver without proposing
// consumer edits or serializing the effective graph into the published lock.
func validateTidyOverlay(ctx context.Context, inputs *tidyInputs, repository Repository) error {
	manifest, err := inputs.observations.capture(v1.ModuleFile, nil)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	if !bytes.Equal(manifest.data, inputs.original) {
		return fmt.Errorf("protobuf.mod changed while reading: %w", sourceview.ErrChanged)
	}
	_, err = inputs.observations.capture(v1.LockFile, nil)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	view := newTidySourceView(inputs.observations, nil)
	locals := newLocalReplacements(inputs.root, inputs.module, nil)
	locals.read = view.readLocalReplacement
	locals.inspect = view.inspectLocalRoots
	graph, err := resolveEffectiveGraph(ctx, inputs.root, inputs.module, repository, locals, false)
	if err != nil {
		return fmt.Errorf("resolveEffectiveGraph: %w", err)
	}
	inputs.roots, err = ModuleSources(inputs.root, inputs.module)
	if err != nil {
		return fmt.Errorf("ModuleSources: %w", err)
	}
	err = inputs.captureResolvedSources(repository, graph.remoteLock)
	if err != nil {
		return fmt.Errorf("captureResolvedSources: %w", err)
	}
	if inputs.completeCacheOwnership {
		err = inputs.captureConsumerSources()
		if err != nil {
			return fmt.Errorf("captureConsumerSources: %w", err)
		}
	}
	view.roots = append(slices.Clone(inputs.roots), graph.Sources...)
	names := make([]string, 0, len(graph.Modules))
	for name := range graph.Modules {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		module := graph.Modules[name]
		err = view.observeModuleMetadata(module.Directory, module.Module)
		if err != nil {
			return fmt.Errorf("observeModuleMetadata: %w", err)
		}
	}
	_, err = view.validateImports(ctx, inputs.files, TidyResult{})
	if err != nil {
		return fmt.Errorf("validateImports: %w", err)
	}
	if len(graph.remoteLock.Modules) > 0 {
		err = verifyTidyCache(ctx, graph.remoteLock, repository)
		if err != nil {
			return fmt.Errorf("verifyTidyCache: %w", err)
		}
	}
	err = view.verifyImportCandidates(ctx)
	if err != nil {
		return fmt.Errorf("verifyImportCandidates: %w", err)
	}
	err = checkTidySourceSelection(inputs.observations, inputs.roots, inputs.files)
	if err != nil {
		return fmt.Errorf("checkTidySourceSelection: %w", err)
	}
	err = view.verifyLocalRoots(ctx)
	if err != nil {
		return fmt.Errorf("verifyLocalRoots: %w", err)
	}
	err = view.verifyLocalMetadata(ctx)
	if err != nil {
		return fmt.Errorf("verifyLocalMetadata: %w", err)
	}
	err = inputs.observations.verifyCapturedInputs()
	if err != nil {
		return fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	return nil
}
