package modules

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// tidyOldNamespace retains the proto inventory and ownership that justified an
// unused old import name. Its cache proof and absence evidence survive until
// commit instead of consulting a mutable filesystem after verification.
type tidyOldNamespace struct {
	tx    *resolvedFilesTransaction
	roots SourceRoots
	files []string
	lock  v1.Lock
}

func (view *tidySourceView) captureOldNamespace(ctx context.Context, repository Repository, entry v1.LockedModule) (map[string]bool, error) {
	previous := v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}
	err := repository.Install(ctx, previous)
	if err != nil {
		return nil, fmt.Errorf("Install: %w", err)
	}
	directory, module, err := repository.Cached(entry)
	if err != nil {
		return nil, fmt.Errorf("Cached: %w", err)
	}
	err = view.observeModuleMetadata(directory, module)
	if err != nil {
		return nil, fmt.Errorf("observeModuleMetadata: %w", err)
	}
	roots, err := ModuleSources(directory, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	tx, err := view.dependencyScope(directory)
	if err != nil {
		return nil, fmt.Errorf("dependencyScope: %w", err)
	}
	files, names, err := captureOldProtoNamespace(tx, roots)
	if err != nil {
		return nil, fmt.Errorf("captureOldProtoNamespace: %w", err)
	}
	proof := tidyOldNamespace{tx: tx, roots: roots, files: files, lock: previous}
	err = proof.verify(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}
	view.old = append(view.old, proof)
	return names, nil
}

func captureOldProtoNamespace(tx *resolvedFilesTransaction, roots SourceRoots) ([]string, map[string]bool, error) {
	names, captured := make(map[string]bool), make(map[string]bool)
	var files []string
	for _, root := range roots {
		err := roots.WalkSelected(root, roots.FileAllowed(), func(path string) error {
			file, err := filepath.Rel(tx.requestedRoot, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			_, err = tx.capture(file, roots)
			if err != nil {
				return fmt.Errorf("capture: %w", err)
			}
			if !captured[file] {
				captured[file] = true
				files = append(files, file)
			}
			name, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			names[filepath.ToSlash(name)] = true
			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("WalkSelected: %w", err)
		}
	}
	slices.Sort(files)
	return files, names, nil
}

func (proof tidyOldNamespace) verify(ctx context.Context, repository Repository) error {
	err := proof.checkInputs()
	if err != nil {
		return fmt.Errorf("checkInputs: %w", err)
	}
	err = tidyInputVerifier(ctx, proof.lock, repository)()
	if err != nil {
		return fmt.Errorf("tidyInputVerifier: %w", err)
	}
	err = proof.checkInputs()
	if err != nil {
		return fmt.Errorf("checkInputs: %w", err)
	}
	return nil
}

func (proof tidyOldNamespace) checkInputs() error {
	err := proof.tx.verifyCapturedInputs()
	if err != nil {
		return fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	err = checkTidySourceSelection(proof.tx, proof.roots, proof.files)
	if err != nil {
		return fmt.Errorf("checkTidySourceSelection: %w", err)
	}
	return nil
}

func (view *tidySourceView) verifyOldNamespaces(ctx context.Context, repository Repository) error {
	for _, proof := range view.old {
		err := proof.verify(ctx, repository)
		if err != nil {
			return fmt.Errorf("verify: %w", err)
		}
	}
	return nil
}
