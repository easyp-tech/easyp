package migration

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/protosource"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func localRoots(cfg legacyConfig) ([]string, []legacyDirectory, error) {
	var roots []string
	var inputs []legacyDirectory
	for _, input := range cfg.Generate.Inputs {
		if input.Directory == nil {
			continue
		}
		dir := *input.Directory
		if dir.Path == "" {
			dir.Path = "."
		}
		if dir.Root == "" {
			dir.Root = "."
		}
		for _, value := range []string{dir.Path, dir.Root} {
			if strings.Contains(value, "$") {
				return nil, nil, fmt.Errorf("directory placeholder %q prevents safe source analysis; supply a literal path or migrate manually", value)
			}
			if !filepath.IsLocal(value) || strings.ContainsAny(value, "*?[]\\ \t\r\n#") {
				return nil, nil, fmt.Errorf("directory %q is external or sliced; manual migration is required to preserve import paths", value)
			}
		}
		dir.Root = filepath.Clean(dir.Root)
		dir.Path = filepath.Clean(dir.Path)
		if !slices.Contains(roots, dir.Root) {
			roots = append(roots, dir.Root)
		}
		inputs = append(inputs, dir)
	}
	if len(inputs) == 0 && len(cfg.Generate.Inputs) == 0 && len(cfg.Generate.Plugins) > 0 {
		return nil, nil, fmt.Errorf("generation plugins have no legacy inputs; choose explicit inputs before migration to avoid widening scope")
	}
	return roots, inputs, nil
}

type localSourceSelection struct {
	files    map[string]string
	packages []string
	paths    []string
}

const localSelectionScopeError = "v1 roots would change generation scope or .proto import names (including hidden/vendor/nested modules); whole roots, exact paths and complete protobuf packages cannot preserve this configuration; split the module or migrate separate generation projects manually"

// proveLocalSelection compares both the physical files and their import names.
// Paths preserve literal directory inputs, including future files in those
// directories. Complete packages are a fallback for mixed-root selections.
func proveLocalSelection(root string, inputs []legacyDirectory, roots []string) (localSourceSelection, error) {
	legacy, current := make(map[string]string), make(map[string]string)
	for _, input := range inputs {
		importRoot := filepath.Join(root, input.Root)
		search := filepath.Join(importRoot, input.Path)
		if err := collectProto(root, importRoot, search, false, legacy); err != nil {
			return localSourceSelection{}, fmt.Errorf("collectProto: %w", err)
		}
	}
	for _, moduleRoot := range roots {
		importRoot := filepath.Join(root, moduleRoot)
		if err := collectProto(root, importRoot, importRoot, true, current); err != nil {
			return localSourceSelection{}, fmt.Errorf("collectProto: %w", err)
		}
	}
	wholeRoots := make(map[string]bool)
	for _, input := range inputs {
		if input.Path == "." {
			wholeRoots[input.Root] = true
		}
	}
	allRoots := true
	for _, moduleRoot := range roots {
		allRoots = allRoots && wholeRoots[moduleRoot]
	}
	// A directory subset must retain its path even when no outside source
	// exists yet; otherwise a later build copy would silently widen targets.
	if maps.Equal(legacy, current) && (allRoots || len(legacy) == 0) {
		return localSourceSelection{files: current}, nil
	}
	// Empty selector lists mean all files at runtime, never no files.
	if len(legacy) == 0 {
		return localSourceSelection{}, fmt.Errorf("%s", localSelectionScopeError)
	}
	for name, physical := range legacy {
		if current[name] != physical {
			return localSourceSelection{}, fmt.Errorf("%s", localSelectionScopeError)
		}
	}
	if selection, ok := tryLocalPathSelection(inputs, legacy, current); ok {
		return selection, nil
	}
	selection, err := proveLocalPackageSelection(root, legacy, current)
	if err != nil {
		return localSourceSelection{}, fmt.Errorf("proveLocalPackageSelection: %w", err)
	}
	return selection, nil
}

func tryLocalPathSelection(inputs []legacyDirectory, legacy, current map[string]string) (localSourceSelection, bool) {
	pathSet := make(map[string]bool)
	for _, input := range inputs {
		pathSet[filepath.ToSlash(filepath.Join(input.Root, input.Path))] = true
	}
	paths := slices.Sorted(maps.Keys(pathSet))
	// Invalid optional path syntax does not invalidate a previously supported
	// package proof. Public generate.paths validation remains strict.
	if v1.ValidatePathSelectors(paths) != nil {
		return localSourceSelection{}, false
	}
	byPath := make(map[string]string)
	matchedPaths := make(map[string]bool)
	for name, modulePath := range current {
		for _, selector := range paths {
			if v1.PathSelectorMatches(selector, filepath.ToSlash(modulePath)) {
				byPath[name] = modulePath
				matchedPaths[selector] = true
			}
		}
	}
	if len(matchedPaths) == len(paths) && maps.Equal(legacy, byPath) {
		return localSourceSelection{files: byPath, paths: paths}, true
	}
	return localSourceSelection{}, false
}

func proveLocalPackageSelection(root string, legacy, current map[string]string) (localSourceSelection, error) {
	packages := make(map[string]bool)
	for _, name := range slices.Sorted(maps.Keys(legacy)) {
		physical := legacy[name]
		declared, err := readSourcePackage(root, physical)
		if err != nil {
			return localSourceSelection{}, fmt.Errorf("readSourcePackage: %w", err)
		}
		if declared == "" {
			return localSourceSelection{}, fmt.Errorf("%s", localSelectionScopeError)
		}
		if err := v1.ValidatePackageSelectors([]string{declared}); err != nil {
			return localSourceSelection{}, fmt.Errorf("ValidatePackageSelectors: %w", err)
		}
		packages[declared] = true
	}
	selected := make(map[string]string)
	for _, name := range slices.Sorted(maps.Keys(current)) {
		physical := current[name]
		declared, err := readSourcePackage(root, physical)
		if err != nil {
			return localSourceSelection{}, fmt.Errorf("readSourcePackage: %w", err)
		}
		if packages[declared] {
			selected[name] = physical
		}
	}
	if !maps.Equal(legacy, selected) {
		return localSourceSelection{}, fmt.Errorf("%s", localSelectionScopeError)
	}
	return localSourceSelection{files: selected, packages: slices.Sorted(maps.Keys(packages))}, nil
}

func readSourcePackage(root, name string) (string, error) {
	raw, err := sourceview.ReadLocal(context.Background(), root, name)
	if err != nil {
		return "", fmt.Errorf("ReadFile: %w", err)
	}
	return protosource.Package(raw), nil
}

func collectProto(root, importRoot, search string, native bool, files map[string]string) error {
	rel, err := filepath.Rel(root, search)
	if err != nil {
		return fmt.Errorf("Rel: %w", err)
	}
	if !filepath.IsLocal(rel) {
		return fmt.Errorf("source %q leaves migration root", search)
	}
	rootRel, err := filepath.Rel(root, importRoot)
	if err != nil {
		return fmt.Errorf("Rel: %w", err)
	}
	rootResolution, err := sourceview.ResolveLocal(context.Background(), root, rootRel)
	if err != nil {
		return fmt.Errorf("ResolveLocal: %w", err)
	}
	physicalImportRoot := filepath.Join(root, filepath.FromSlash(rootResolution.Path))
	return sourceview.WalkLocal(context.Background(), root, rel, func(logical string, resolved sourceview.Resolution, walkErr error) error {
		name := filepath.Join(root, filepath.FromSlash(logical))
		if errors.Is(walkErr, sourceview.ErrCycle) && resolved.Info != nil && resolved.Info.IsDir() {
			if native && path_helpers.HiddenOrVendorSourcePath(importRoot, name) {
				return nil
			}
			return walkErr
		}
		if native && path_helpers.ShouldSkipV1SourceDir(importRoot, name) {
			return fs.SkipDir
		}
		if walkErr != nil {
			if filepath.Ext(name) != ".proto" && logical != filepath.ToSlash(rel) {
				return nil
			}
			return walkErr
		}
		physical := filepath.Join(root, filepath.FromSlash(resolved.Path))
		if resolved.Info.IsDir() {
			if native && path_helpers.ShouldSkipV1SourceDir(physicalImportRoot, physical) {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(name) != ".proto" {
			return nil
		}
		if !resolved.Info.Mode().IsRegular() {
			return fmt.Errorf("source %s is not a regular file", name)
		}
		if native {
			for dir := filepath.Dir(physical); dir != root && sourceWithin(root, dir); dir = filepath.Dir(dir) {
				if path_helpers.ShouldSkipV1SourceDir(physicalImportRoot, dir) {
					return nil
				}
			}
		}
		importName, err := filepath.Rel(importRoot, name)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		logicalName, err := filepath.Rel(root, name)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		importName = filepath.ToSlash(importName)
		if previous, ok := files[importName]; ok && previous != logicalName {
			return fmt.Errorf("roots collide on import %s (%s and %s); manual migration required", importName, previous, logicalName)
		}
		files[importName] = logicalName
		return nil
	})
}

func sourceWithin(root, name string) bool {
	rel, err := filepath.Rel(root, name)
	return err == nil && filepath.IsLocal(rel)
}
