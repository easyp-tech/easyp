package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

// Vendor copies verified locked protobuf sources into one import root.
func (m Mod) Vendor(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	_, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
	}
	if len(module.Replaces) != 0 {
		return fmt.Errorf("module %s: remove local replacements before vendoring a reproducible lock", module.Name)
	}
	cacheBase, err := gitCachePath(getLogger(ctx))
	if err != nil {
		return fmt.Errorf("gitCachePath: %w", err)
	}
	dependencyRoots, err := lockedV1DependencySources(ctx.Context, root, module, cacheBase)
	if err != nil {
		return fmt.Errorf("lockedV1DependencySources: %w", err)
	}
	if err := checkV1ImportPathCollisions(root, module.Roots, dependencyRoots.paths()); err != nil {
		return fmt.Errorf("checkV1ImportPathCollisions: %w", err)
	}
	return writeV1Vendor(root, dependencyRoots)
}

func writeV1Vendor(root string, roots v1SourceRoots) error {
	stage, err := os.MkdirTemp(root, ".easyp-vendor-*")
	if err != nil {
		return fmt.Errorf("MkdirTemp: %w", err)
	}
	defer func() { _ = os.RemoveAll(stage) }()
	for _, source := range roots {
		err := walkV1ProtoFiles(source.path, func(path string) error {
			importPath, err := filepath.Rel(source.path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			if !filepath.IsLocal(importPath) {
				return fmt.Errorf("invalid import path %q", importPath)
			}
			if err := copyV1RegularFile(path, filepath.Join(stage, importPath)); err != nil {
				return fmt.Errorf("copyV1RegularFile: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("vendor %s: %w", source.module, err)
		}
	}
	return replaceV1Vendor(root, stage)
}

func replaceV1Vendor(root, stage string) error {
	target := filepath.Join(root, defaultVendorDir)
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
