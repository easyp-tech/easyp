package api

import (
	"bytes"
	"context"
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
		return err
	}
	original, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	if err != nil {
		return fmt.Errorf("read protobuf.mod: %w", err)
	}
	module, err := v1.ParseModule(strings.NewReader(string(original)))
	if err != nil {
		return err
	}
	lock, err := resolveV1Lock(ctx, root, module)
	if err != nil {
		return err
	}
	cacheRoot, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return err
	}
	updated, err := augmentV1ManifestRequirements(original, root, module, lock, filepath.Join(cacheRoot, "v1", "git"))
	if err != nil {
		return err
	}
	return writeV1ResolvedFiles(root, original, updated, lock)
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

func resolveV1Lock(ctx *cli.Context, root string, module v1.Module) (v1.Lock, error) {
	if len(module.Replaces) > 0 {
		return v1.Lock{}, fmt.Errorf("module %s: remove local replacements before writing a reproducible lock", module.Name)
	}
	cacheRoot, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return v1.Lock{}, err
	}
	callCtx := ctx.Context
	if callCtx == nil {
		callCtx = context.Background()
	}
	gitCacheRoot := filepath.Join(cacheRoot, "v1", "git")
	lock, err := buildV1Lock(callCtx, module, gitCacheRoot)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	if err := downloadV1Lock(callCtx, lock, gitCacheRoot); err != nil {
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
		return err
	}

	lockPath := filepath.Join(root, "protobuf.lock")
	tmp, err := os.CreateTemp(root, ".protobuf.lock-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append([]byte("# protobuf.lock - GENERATED FILE, DO NOT EDIT MANUALLY\n"), raw...)); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), lockPath); err != nil {
		return fmt.Errorf("write protobuf.lock: %w", err)
	}
	return nil
}

func (m Mod) downloadV1IfPresent(ctx *cli.Context) (bool, error) {
	root, err := os.Getwd()
	if err != nil {
		return true, err
	}
	manifest, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	if !v1.IsModuleManifest(manifest) {
		return false, nil // v0 protobuf.mod uses direct/replace blocks
	}
	module, err := v1.ParseModule(strings.NewReader(string(manifest)))
	if err != nil {
		return true, err
	}
	lock, err := readV1Lock(filepath.Join(root, "protobuf.lock"))
	if err != nil {
		return true, fmt.Errorf("read protobuf.lock; run easyp mod tidy: %w", err)
	}
	if err := validateV1Requirements(remoteV1Requirements(module), lock); err != nil {
		return true, err
	}
	cacheBase, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return true, err
	}
	callCtx := ctx.Context
	if callCtx == nil {
		callCtx = context.Background()
	}
	gitCacheRoot := filepath.Join(cacheBase, "v1", "git")
	if err := downloadV1Lock(callCtx, lock, gitCacheRoot); err != nil {
		return true, err
	}
	_, err = rootsFromV1Lock(lock, gitCacheRoot)
	return true, err
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
					if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
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
