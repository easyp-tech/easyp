package modules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

// Vendor copies verified locked protobuf sources into one import root.
func Vendor(ctx context.Context, root string, repository Cache) error {
	_, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	if len(module.Replaces) != 0 {
		return fmt.Errorf("module %s: remove local replacements before vendoring a reproducible lock", module.Name)
	}
	dependencyRoots, err := EnsureLockedSources(ctx, root, module, repository)
	if err != nil {
		return fmt.Errorf("EnsureLockedSources: %w", err)
	}
	if err := CheckImportCollisions(root, module.Roots, dependencyRoots.Paths()); err != nil {
		return fmt.Errorf("CheckImportCollisions: %w", err)
	}
	return writeV1Vendor(root, dependencyRoots)
}

func writeV1Vendor(root string, roots SourceRoots) error {
	stage, err := os.MkdirTemp(root, ".easyp-vendor-*")
	if err != nil {
		return fmt.Errorf("MkdirTemp: %w", err)
	}
	defer func() { _ = os.RemoveAll(stage) }()
	for _, source := range roots {
		err := WalkProtoFiles(source.Path, func(path string) error {
			importPath, err := filepath.Rel(source.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			if !filepath.IsLocal(importPath) {
				return fmt.Errorf("invalid import path %q", importPath)
			}
			if err := disk.CopyRegularFile(path, filepath.Join(stage, importPath)); err != nil {
				return fmt.Errorf("CopyRegularFile: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("vendor %s: %w", source.Module, err)
		}
	}
	return replaceV1Vendor(root, stage)
}

func replaceV1Vendor(root, stage string) error {
	target := filepath.Join(root, VendorDir)
	backup := filepath.Join(root, ".easyp-vendor-backup")
	if _, err := os.Lstat(backup); !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		return fmt.Errorf("vendor backup %s already exists", backup)
	}
	info, err := os.Lstat(target)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Lstat: %w", err)
	}
	if info != nil {
		if !info.IsDir() {
			return fmt.Errorf("vendor path %s is not a regular directory", target)
		}
		if err := os.Rename(target, backup); err != nil {
			return fmt.Errorf("Rename: %w", err)
		}
	}
	if err := os.Rename(stage, target); err != nil {
		if restoreErr := restoreV1Vendor(backup, target); restoreErr != nil {
			return errors.Join(err, restoreErr)
		}
		return fmt.Errorf("Rename: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("RemoveAll: %w", err)
	}
	return nil
}

func restoreV1Vendor(backup, target string) error {
	_, err := os.Lstat(backup)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Lstat: %w", err)
	}
	if err := os.Rename(backup, target); err != nil {
		return fmt.Errorf("Rename: %w", err)
	}
	return nil
}
