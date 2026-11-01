package modules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// Metadata-free root selection depends on the producer's full intrinsic import
// map, including sources the consumer does not import. Retain that evidence
// separately from tidy's reachable namespace and its fixed-root checks.
type tidyLocalRoots struct {
	directory string
	module    v1.Module
	candidate importRootModule
}

type tidyLocalMetadata struct {
	directory string
	name      string
	files     map[string]bool
}

func (view *tidySourceView) readLocalReplacement(directory, name string) (EffectiveModule, error) {
	directory, err := filepath.Abs(directory)
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("Abs: %w", err)
	}
	err = view.observeModuleMetadata(directory, v1.Module{})
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("observeModuleMetadata: %w", err)
	}
	scope, err := view.dependencyScope(directory)
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("dependencyScope: %w", err)
	}
	files := make(map[string]bool)
	entry, err := readLocalReplacementWithReader(directory, name, func(file string) ([]byte, error) {
		state, err := scope.capture(file, nil)
		if err != nil {
			return nil, fmt.Errorf("capture: %w", err)
		}
		if !state.exists {
			return nil, &os.PathError{Op: "open", Path: filepath.Join(directory, file), Err: os.ErrNotExist}
		}
		files[file] = true
		return state.data, nil
	})
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("readLocalReplacementWithReader: %w", err)
	}
	err = view.observeModuleMetadata(directory, entry.Module)
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("observeModuleMetadata: %w", err)
	}
	manifest, _ := scope.capturedFile(v1.ModuleFile)
	if !v1.IsModuleManifest(manifest.data) {
		view.compatibility = append(view.compatibility, tidyLocalMetadata{directory: directory, name: name, files: files})
	}
	return entry, nil
}

// Fallback metadata adaptation scans nested configs too. Revisit the same
// bounded metadata tree to catch new directories that could change module
// identity, roots or requirements without changing the proto inventory.
func (view *tidySourceView) verifyLocalMetadata(ctx context.Context) error {
	for _, evidence := range view.compatibility {
		scope, err := view.dependencyScope(evidence.directory)
		if err != nil {
			return fmt.Errorf("dependencyScope: %w", err)
		}
		seen := make(map[string]bool)
		_, err = moduleconfig.ReadLocalDependencyWithReader(evidence.directory, evidence.name, func(file string) ([]byte, error) {
			err := ctx.Err()
			if err != nil {
				return nil, fmt.Errorf("Err: %w", err)
			}
			state, err := scope.capture(file, nil)
			if err != nil {
				return nil, fmt.Errorf("capture: %w", err)
			}
			if !state.exists {
				return nil, &os.PathError{Op: "open", Path: filepath.Join(evidence.directory, file), Err: os.ErrNotExist}
			}
			if !evidence.files[file] {
				return nil, fmt.Errorf("local metadata inventory for %s changed: %w", evidence.name, sourceview.ErrChanged)
			}
			seen[file] = true
			return state.data, nil
		})
		if err != nil {
			return fmt.Errorf("ReadLocalDependencyWithReader: %w", err)
		}
		if len(seen) != len(evidence.files) {
			return fmt.Errorf("local metadata inventory for %s changed: %w", evidence.name, sourceview.ErrChanged)
		}
	}
	return nil
}

func (view *tidySourceView) inspectLocalRoots(ctx context.Context, directory string, module v1.Module, hinted bool) (importRootModule, error) {
	candidate, err := inspectLocalRootCandidates(ctx, directory, module, hinted)
	if err != nil {
		return importRootModule{}, fmt.Errorf("inspectLocalRootCandidates: %w", err)
	}
	if candidate.fixed {
		return candidate, nil
	}
	scope, err := view.dependencyScope(directory)
	if err != nil {
		return importRootModule{}, fmt.Errorf("dependencyScope: %w", err)
	}
	for _, file := range candidate.inspection.Files {
		state, err := scope.capture(filepath.FromSlash(file.Path), nil)
		if err != nil {
			return importRootModule{}, fmt.Errorf("capture: %w", err)
		}
		physical := filepath.Join(scope.canonicalRoot, filepath.FromSlash(state.path))
		if !state.exists || physical != file.Identity || !bytes.Equal(state.data, file.Content) {
			return importRootModule{}, fmt.Errorf("local root evidence %s changed: %w", file.Path, sourceview.ErrChanged)
		}
	}
	view.inferred = append(view.inferred, tidyLocalRoots{directory: directory, module: module, candidate: candidate})
	return candidate, nil
}

func (view *tidySourceView) verifyLocalRoots(ctx context.Context) error {
	for _, evidence := range view.inferred {
		current, err := inspectLocalRootCandidates(ctx, evidence.directory, evidence.module, false)
		if err != nil {
			return fmt.Errorf("inspectLocalRootCandidates: %w", err)
		}
		sameFiles := slices.EqualFunc(evidence.candidate.inspection.Files, current.inspection.Files, func(before, after RootProtoFile) bool {
			return before.Path == after.Path && before.Identity == after.Identity && bytes.Equal(before.Content, after.Content)
		})
		sameProblems := slices.EqualFunc(evidence.candidate.inspection.Problems, current.inspection.Problems, func(before, after RootPathProblem) bool {
			return before.Path == after.Path
		})
		if !sameFiles || !sameProblems {
			return fmt.Errorf("local root evidence for %s changed: %w", evidence.module.Name, sourceview.ErrChanged)
		}
	}
	return nil
}
