package modules

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type tidyCachedModule struct {
	directory string
	module    v1.Module
}

func (view *tidySourceView) cachedSources(ctx context.Context, lock v1.Lock, repository Repository) (SourceRoots, error) {
	var cached []tidyCachedModule
	for _, entry := range lock.Modules {
		directory, module, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("Cached: %w", err)
		}
		directory, err = filepath.Abs(directory)
		if err != nil {
			return nil, fmt.Errorf("Abs: %w", err)
		}
		// Discover the directory, then capture and verify its metadata before
		// using the returned requirements or roots to build the source graph.
		err = view.observeModuleMetadata(directory, module)
		if err != nil {
			return nil, fmt.Errorf("observeModuleMetadata: %w", err)
		}
		cached = append(cached, tidyCachedModule{directory: directory, module: module})
	}
	err := view.observations.verifyCapturedInputs()
	if err != nil {
		return nil, fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	err = verifyTidyCache(ctx, lock, repository)
	if err != nil {
		return nil, fmt.Errorf("verifyTidyCache: %w", err)
	}
	err = view.observations.verifyCapturedInputs()
	if err != nil {
		return nil, fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	var roots SourceRoots
	for _, dependency := range cached {
		err := ValidateRequirements(dependency.module.Requires, lock)
		if err != nil {
			return nil, fmt.Errorf("ValidateRequirements: %w", err)
		}
		selected, err := ModuleSources(dependency.directory, dependency.module)
		if err != nil {
			return nil, fmt.Errorf("ModuleSources: %w", err)
		}
		roots = append(roots, selected...)
	}
	return roots, nil
}

func (view *tidySourceView) observeModuleMetadata(directory string, module v1.Module) error {
	tx, err := view.dependencyScope(directory)
	if err != nil {
		return fmt.Errorf("dependencyScope: %w", err)
	}
	for _, root := range append([]string{"."}, module.Roots...) {
		// Native module locations and Buf workspace member metadata are
		// ancestors of the selected roots. Capture those validation inputs,
		// including their absence, without reading unrelated proto bodies.
		for directory := root; ; directory = filepath.Dir(directory) {
			for _, name := range []string{v1.ModuleFile, "easyp.yaml", "buf.yaml", "buf.work.yaml", "buf.lock"} {
				_, err := tx.capture(filepath.Join(directory, name), nil)
				if err != nil {
					return fmt.Errorf("capture: %w", err)
				}
			}
			if directory == "." {
				break
			}
		}
	}
	return nil
}

func (view *tidySourceView) dependencyScope(boundary string) (*resolvedFileObservations, error) {
	tx, exists := view.scopes[boundary]
	if exists {
		return tx, nil
	}
	tx, err := view.observations.observeDirectory(boundary)
	if err != nil {
		return nil, fmt.Errorf("observeDirectory: %w", err)
	}
	view.scopes[boundary] = tx
	return tx, nil
}

type tidyCachedVerifier interface {
	VerifyCached(context.Context, v1.Lock) error
}

func verifyTidyCache(ctx context.Context, lock v1.Lock, repository Repository) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("Err: %w", err)
	}
	if verifier, supported := repository.(tidyCachedVerifier); supported {
		err = verifier.VerifyCached(ctx, lock)
		if err != nil {
			return fmt.Errorf("VerifyCached: %w", err)
		}
	} else {
		// The base Cache contract verifies installed bytes against the lock.
		// Callers recheck snapshots after this potentially repairing operation.
		err = repository.Install(ctx, lock)
		if err != nil {
			return fmt.Errorf("Install: %w", err)
		}
	}
	return nil
}
