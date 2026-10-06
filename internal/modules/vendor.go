package modules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"io"
)

// Vendor copies the effective dependency sources into one import root.
// Local replacements are copied without writing or creating the published lock.
func Vendor(ctx context.Context, root string, repository Cache) error {
	_, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	dependencyRoots, err := EnsureSources(ctx, root, module, repository)
	if err != nil {
		return fmt.Errorf("EnsureSources: %w", err)
	}
	own, err := ModuleSources(root, module)
	if err != nil {
		return fmt.Errorf("ModuleSources: %w", err)
	}
	if err := CheckSourceCollisions(append(own, dependencyRoots...)); err != nil {
		return fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	return writeV1Vendor(root, dependencyRoots)
}

func writeV1Vendor(root string, roots SourceRoots) error {
	stage, err := os.MkdirTemp(root, ".easyp-vendor-*")
	if err != nil {
		return fmt.Errorf("MkdirTemp: %w", err)
	}
	defer func() { _ = os.RemoveAll(stage) }()
	allowed := roots.FileAllowed()
	for _, source := range roots {
		err := source.Walk(func(path string) error {
			if allowed != nil && !allowed(path) {
				return nil
			}
			importPath, err := filepath.Rel(source.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			if !filepath.IsLocal(importPath) {
				return fmt.Errorf("invalid import path %q", importPath)
			}
			if err := copySourceFile(roots, path, filepath.Join(stage, importPath)); err != nil {
				return fmt.Errorf("copySourceFile: %w", err)
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

func copySourceFile(roots SourceRoots, source, destination string) (resultErr error) {
	in, err := roots.OpenSourceFile(source)
	if err != nil {
		return fmt.Errorf("OpenSourceFile: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, in.Close()) }()
	statter, ok := in.(interface{ Stat() (os.FileInfo, error) })
	if !ok {
		return fmt.Errorf("source %q has no file metadata", source)
	}
	info, err := statter.Stat()
	if err != nil {
		return fmt.Errorf("Stat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source %q is not regular", source)
	}
	err = os.MkdirAll(filepath.Dir(destination), 0o755)
	if err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("OpenFile: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, out.Close()) }()
	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("Copy: %w", err)
	}
	return nil
}
