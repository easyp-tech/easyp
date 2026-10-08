package modules

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// TidyResult reports import edits that have been validated and committed.
type TidyResult struct {
	Imports []ImportRewrite
}

// ImportRewrite identifies one applied consumer import mapping and its pins.
// File is relative to the owning module and names the physical write target.
type ImportRewrite struct {
	File       string
	From       string
	To         string
	Module     string
	OldVersion string
	Version    string
	OldCommit  string
	Commit     string
	OldRoots   []string
	Roots      []string
}

// TidyWithReport resolves the manifest's requirements, checks and applies unique
// pinned import mappings, and returns their report only after successful commit.
// Local replacements only validate their ephemeral graph and produce no edits.
func TidyWithReport(ctx context.Context, root string, repository Repository) (_ TidyResult, resultErr error) {
	tx, err := newResolvedFilesTransaction(root)
	if err != nil {
		return TidyResult{}, fmt.Errorf("newResolvedFilesTransaction: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, tx.close()) }()
	inputs, err := captureTidyInputs(tx.resolvedFileObservations, repository)
	if err != nil {
		return TidyResult{}, fmt.Errorf("captureTidyInputs: %w", err)
	}
	if len(inputs.module.Replaces) > 0 {
		err = validateLocalOverlay(ctx, inputs.root, inputs.module, repository, false)
		return TidyResult{}, err
	}
	if inputs.completeCacheOwnership {
		// Complete cache ownership permits a consumer checkpoint before
		// resolution; generic repositories retain their later checkpoint.
		err = inputs.captureConsumerSources()
		if err != nil {
			return TidyResult{}, fmt.Errorf("captureConsumerSources: %w", err)
		}
	}
	resolved, err := resolveV1Graph(ctx, inputs.module, inputs.previous, repository, graphResolveRequest{
		preserveHeads: true,
		transitions:   namespaceTransitionsTidyRepairs,
	})
	if err != nil {
		return TidyResult{}, fmt.Errorf("resolveV1Graph: %w", err)
	}
	lock := resolved.lockFile()
	view := newTidySourceView(tx.resolvedFileObservations, inputs.roots)
	err = repository.Install(ctx, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("Install: %w", err)
	}
	view.retainPinnedSources(resolved)
	dependencies, err := view.cachedSources(ctx, lock, repository)
	if err != nil {
		return TidyResult{}, fmt.Errorf("cachedSources: %w", err)
	}
	err = inputs.captureResolvedSources(repository, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("captureResolvedSources: %w", err)
	}
	needed, err := inputs.capturedImports()
	if err != nil {
		return TidyResult{}, fmt.Errorf("capturedImports: %w", err)
	}
	planner := newTidyImportPlanner(resolved, inputs.previous, repository)
	bindings, err := planner.tidyImportBindings(ctx, needed, view)
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyImportBindings: %w", err)
	}
	err = inputs.excludePreviousPinSources(repository)
	if err != nil {
		return TidyResult{}, fmt.Errorf("excludePreviousPinSources: %w", err)
	}
	roots := append(slices.Clone(inputs.roots), dependencies...)
	err = CheckSourceCollisions(roots)
	if err != nil {
		return TidyResult{}, fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	view.roots = roots
	tx.validateInputs = func() error {
		err := checkTidySourceSelection(inputs.observations, inputs.roots, inputs.files)
		if err != nil {
			return fmt.Errorf("checkTidySourceSelection: %w", err)
		}
		err = view.verifyOldNamespaces(ctx, repository)
		if err != nil {
			return fmt.Errorf("verifyOldNamespaces: %w", err)
		}
		return verifyTidyCache(ctx, lock, repository)
	}
	report, err := bindings.planRewrites(tx, inputs.files, view)
	if err != nil {
		return TidyResult{}, fmt.Errorf("planRewrites: %w", err)
	}
	owners, err := view.validateImports(ctx, inputs.files, report)
	if err != nil {
		return TidyResult{}, fmt.Errorf("validateImports: %w", err)
	}
	updated, err := augmentV1ManifestRequirementsWithOwners(inputs.original, inputs.module, lock, repository, owners)
	if err != nil {
		return TidyResult{}, fmt.Errorf("augmentV1ManifestRequirementsWithOwners: %w", err)
	}
	err = ctx.Err()
	if err != nil {
		return TidyResult{}, fmt.Errorf("Err: %w", err)
	}
	err = writeV1ResolvedFilesTransaction(tx, updated, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("writeV1ResolvedFilesTransaction: %w", err)
	}
	return report, nil
}
