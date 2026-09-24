package modules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Tidy resolves v1 requirements and records exact Git commits and content
// hashes in protobuf.lock.
func Tidy(ctx context.Context, root string, repository Repository) error {
	original, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	existing, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("ReadLock: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, module, existing, repository)
	if err != nil {
		return fmt.Errorf("resolveV1LockWithPins: %w", err)
	}
	updated, err := augmentV1ManifestRequirements(original, root, module, lock, repository)
	if err != nil {
		return fmt.Errorf("augmentV1ManifestRequirements: %w", err)
	}
	return writeV1ResolvedFiles(root, original, updated, lock)
}

func resolveV1LockWithPins(ctx context.Context, root string, module v1.Module, existing v1.Lock, repository Repository) (v1.Lock, error) {
	if len(module.Replaces) > 0 {
		return v1.Lock{}, fmt.Errorf("module %s: remove local replacements before writing a reproducible lock", module.Name)
	}
	pins := make(map[string]v1.LockedModule, len(existing.Modules))
	for _, entry := range existing.Modules {
		pins[entry.Source] = entry
	}
	lock, err := Resolve(ctx, module, repository, pins)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	if err := repository.Install(ctx, lock); err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	dependencyRoots, err := cachedRoots(lock, repository)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	if err := CheckImportCollisions(root, module.Roots, dependencyRoots); err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	unresolved, err := unresolvedV1ImportsWithRoots(root, module.Roots, dependencyRoots)
	if err != nil {
		return v1.Lock{}, err
	}
	if len(unresolved) > 0 {
		return v1.Lock{}, fmt.Errorf("module %s: cannot resolve imports %v", module.Name, unresolved)
	}
	return lock, nil
}

// Download installs the exact v1 dependencies recorded in protobuf.lock.
func Download(ctx context.Context, root string, repository Cache) error {
	_, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	lock, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil {
		return fmt.Errorf("read protobuf.lock; run easyp mod tidy: %w", err)
	}
	if err := ValidateRequirements(RemoteRequirements(module), lock); err != nil {
		return err
	}
	if err := repository.Install(ctx, lock); err != nil {
		return err
	}
	_, err = cachedRoots(lock, repository)
	return err
}
