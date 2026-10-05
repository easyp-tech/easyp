package migration

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/protosource"
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
	const scopeError = "v1 roots would change generation scope or .proto import names (including hidden/vendor/nested modules); whole roots, exact paths and complete protobuf packages cannot preserve this configuration; split the module or migrate separate generation projects manually"
	// Empty selector lists mean all files at runtime, never no files.
	if len(legacy) == 0 {
		return localSourceSelection{}, fmt.Errorf("%s", scopeError)
	}
	for name, physical := range legacy {
		if current[name] != physical {
			return localSourceSelection{}, fmt.Errorf("%s", scopeError)
		}
	}
	pathSet := make(map[string]bool)
	for _, input := range inputs {
		pathSet[filepath.ToSlash(input.Path)] = true
	}
	paths := slices.Sorted(maps.Keys(pathSet))
	// Invalid optional path syntax does not invalidate a previously supported
	// package proof. Public generate.paths validation remains strict.
	if v1.ValidatePathSelectors(paths) == nil {
		byPath := make(map[string]string)
		matchedPaths := make(map[string]bool)
		for name, physical := range current {
			for _, selector := range paths {
				if v1.PathSelectorMatches(selector, name) {
					byPath[name] = physical
					matchedPaths[selector] = true
				}
			}
		}
		// A dot selector applies to every root. It cannot preserve a whole input
		// in one root combined with a narrower input in another root over time.
		if (!pathSet["."] || allRoots) && len(matchedPaths) == len(paths) && maps.Equal(legacy, byPath) {
			return localSourceSelection{files: byPath, paths: paths}, nil
		}
	}
	packages := make(map[string]bool)
	for _, name := range slices.Sorted(maps.Keys(legacy)) {
		physical := legacy[name]
		declared, err := readSourcePackage(root, physical)
		if err != nil {
			return localSourceSelection{}, fmt.Errorf("readSourcePackage: %w", err)
		}
		if declared == "" {
			return localSourceSelection{}, fmt.Errorf("%s", scopeError)
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
		return localSourceSelection{}, fmt.Errorf("%s", scopeError)
	}
	return localSourceSelection{files: selected, packages: slices.Sorted(maps.Keys(packages))}, nil
}

func readSourcePackage(root, name string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return "", fmt.Errorf("ReadFile: %w", err)
	}
	return protosource.Package(raw), nil
}

func collectProto(root, importRoot, search string, native bool, files map[string]string) error {
	// WalkDir does not follow links, but a linked starting directory/parent must
	// also be rejected before traversal.
	rel, err := filepath.Rel(root, search)
	if err != nil {
		return fmt.Errorf("Rel: %w", err)
	}
	part := root
	for _, segment := range strings.Split(rel, string(filepath.Separator)) {
		part = filepath.Join(part, segment)
		info, err := os.Lstat(part)
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("source %s is not a regular directory; manual migration required", part)
		}
	}
	err = filepath.WalkDir(search, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source symlink %s prevents safe scope analysis", name)
		}
		if entry.IsDir() {
			if native && path_helpers.ShouldSkipV1SourceDir(importRoot, name) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(name) != ".proto" {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("source %s is not a regular file", name)
		}
		importName, err := filepath.Rel(importRoot, name)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		physicalName, err := filepath.Rel(root, name)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		importName = filepath.ToSlash(importName)
		if previous, ok := files[importName]; ok && previous != physicalName {
			return fmt.Errorf("roots collide on import %s (%s and %s); manual migration required", importName, previous, physicalName)
		}
		files[importName] = physicalName
		return nil
	})
	if err != nil {
		return fmt.Errorf("WalkDir: %w", err)
	}
	return nil
}
