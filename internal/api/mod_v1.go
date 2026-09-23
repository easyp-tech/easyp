package api

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Tidy resolves v1 requirements and records exact Git commits and content
// hashes in protobuf.lock.
func (m Mod) Tidy(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	original, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
	}
	existing, err := readV1Lock(filepath.Join(root, v1.LockFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("readV1Lock: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, module, existing)
	if err != nil {
		return fmt.Errorf("resolveV1LockWithPins: %w", err)
	}
	cacheRoot, err := gitCachePath(getLogger(ctx))
	if err != nil {
		return fmt.Errorf("gitCachePath: %w", err)
	}
	updated, err := augmentV1ManifestRequirements(original, root, module, lock, cacheRoot)
	if err != nil {
		return fmt.Errorf("augmentV1ManifestRequirements: %w", err)
	}
	return writeV1ResolvedFiles(root, original, updated, lock)
}

func readV1Manifest(root string) ([]byte, v1.Module, error) {
	original, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
	if err != nil {
		return nil, v1.Module{}, fmt.Errorf("ReadFile: %w", err)
	}
	module, err := v1.ParseModule(bytes.NewReader(original))
	if err != nil {
		return nil, v1.Module{}, fmt.Errorf("ParseModule: %w", err)
	}
	return original, module, nil
}

func writeV1ResolvedFiles(root string, original, updated []byte, lock v1.Lock) error {
	changed := !bytes.Equal(original, updated)
	if changed {
		if err := writeV1Manifest(root, updated); err != nil {
			return err
		}
	}
	if err := writeV1Lock(root, lock); err != nil {
		if changed {
			if rollbackErr := writeV1Manifest(root, original); rollbackErr != nil {
				return errors.Join(err, fmt.Errorf("restore protobuf.mod: %w", rollbackErr))
			}
		}
		return err
	}
	return nil
}

func resolveV1LockWithPins(ctx *cli.Context, root string, module v1.Module, existing v1.Lock) (v1.Lock, error) {
	if len(module.Replaces) > 0 {
		return v1.Lock{}, fmt.Errorf("module %s: remove local replacements before writing a reproducible lock", module.Name)
	}
	cacheRoot, err := gitCachePath(getLogger(ctx))
	if err != nil {
		return v1.Lock{}, err
	}
	gitCacheRoot := cacheRoot
	pins := make(map[string]v1.LockedModule, len(existing.Modules))
	for _, entry := range existing.Modules {
		pins[entry.Source] = entry
	}
	lock, err := buildV1LockWithPins(ctx.Context, module, gitCacheRoot, pins)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	if err := downloadV1Lock(ctx.Context, lock, gitCacheRoot); err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	dependencyRoots, err := rootsFromV1Lock(lock, gitCacheRoot)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	if err := checkV1ImportPathCollisions(root, module.Roots, dependencyRoots); err != nil {
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

func writeV1Lock(root string, lock v1.Lock) error {
	raw, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}

	raw = append([]byte("# protobuf.lock - GENERATED FILE, DO NOT EDIT MANUALLY\n"), raw...)
	if err := writeAtomicFile(filepath.Join(root, v1.LockFile), raw, 0o600); err != nil {
		return fmt.Errorf("writeAtomicFile: %w", err)
	}
	return nil
}

// Download installs the exact v1 dependencies recorded in protobuf.lock.
func (m Mod) Download(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	_, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
	}
	lock, err := readV1Lock(filepath.Join(root, v1.LockFile))
	if err != nil {
		return fmt.Errorf("read protobuf.lock; run easyp mod tidy: %w", err)
	}
	if err := validateV1Requirements(remoteV1Requirements(module), lock); err != nil {
		return err
	}
	cacheBase, err := gitCachePath(getLogger(ctx))
	if err != nil {
		return err
	}
	gitCacheRoot := cacheBase
	if err := downloadV1Lock(ctx.Context, lock, gitCacheRoot); err != nil {
		return err
	}
	_, err = rootsFromV1Lock(lock, gitCacheRoot)
	return err
}

func unresolvedV1Imports(moduleDir string, roots []string) ([]string, error) {
	return unresolvedV1ImportsWithRoots(moduleDir, roots, nil)
}

func unresolvedV1ImportsWithRoots(moduleDir string, roots, dependencyRoots []string) ([]string, error) {
	missing := map[string]struct{}{}
	allRoots := make([]string, 0, len(roots)+len(dependencyRoots))
	for _, root := range roots {
		allRoots = append(allRoots, filepath.Join(moduleDir, root))
	}
	allRoots = append(allRoots, dependencyRoots...)
	for _, root := range roots {
		sourceRoot := filepath.Join(moduleDir, root)
		err := walkV1ProtoFiles(sourceRoot, func(path string) error {
			imports, err := readV1ProtoImports(path)
			if err != nil {
				return fmt.Errorf("readV1ProtoImports: %w", err)
			}
			for _, importPath := range imports {
				if strings.HasPrefix(importPath, "google/protobuf/") {
					continue
				}
				found := false
				for _, candidateRoot := range allRoots {
					candidate := filepath.Join(candidateRoot, importPath)
					info, err := os.Stat(candidate)
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					if err != nil {
						return fmt.Errorf("stat import %s: %w", importPath, err)
					}
					if info.Mode().IsRegular() {
						found = true
						break
					}
				}
				if !found {
					missing[importPath] = struct{}{}
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	paths := make([]string, 0, len(missing))
	for path := range missing {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths, nil
}
