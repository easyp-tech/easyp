package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)

// SourceRoot associates a physical import root with its module identity.
type SourceRoot struct {
	Path        string
	Module      string
	fileAllowed func(string) bool
}

// SourceRoots preserves import precedence and module ownership.
type SourceRoots []SourceRoot

// Paths returns a new slice in the same order as the source roots.
func (roots SourceRoots) Paths() []string {
	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.Path)
	}
	return paths
}

// Walk visits only proto files selected by the module's source metadata.
func (root SourceRoot) Walk(visit func(string) error) error {
	return WalkProtoFiles(root.Path, func(path string) error {
		if root.fileAllowed != nil && (!root.fileAllowed(path) || !root.fileAllowed(physicalSourcePath(path))) {
			return nil
		}
		return visit(path)
	})
}

func (root SourceRoot) allows(path, importRoot string) bool {
	return root.allowsPath(filepath.Clean(path), filepath.Clean(importRoot)) &&
		root.allowsPath(physicalSourcePath(path), physicalSourcePath(importRoot))
}

func (root SourceRoot) allowsPath(path, importRoot string) bool {
	if root.fileAllowed != nil && !root.fileAllowed(path) {
		return false
	}
	// A containing consumer root does not own files inside a nested module.
	// Import lookup must use the same boundaries as source discovery.
	for directory := filepath.Dir(path); sourcePathWithin(directory, importRoot) && filepath.Clean(directory) != filepath.Clean(importRoot); directory = filepath.Dir(directory) {
		if path_helpers.ShouldSkipV1SourceDir(importRoot, directory) {
			return false
		}
	}
	return true
}

// FileAllowed returns an immutable source filter for physical import lookup.
// Paths outside these roots remain available to the consumer's own inputs.
// Overlapping roots select the union of their allowed files. Both the import's
// spelling and its physical target must pass their respective owning selections.
func (roots SourceRoots) FileAllowed() func(string) bool {
	if !slices.ContainsFunc(roots, func(root SourceRoot) bool { return root.fileAllowed != nil }) {
		return nil
	}
	type selection struct {
		path string
		root SourceRoot
	}
	selections := make([]selection, 0, len(roots))
	for _, root := range roots {
		selections = append(selections, selection{path: filepath.Clean(root.Path), root: root})
		canonical, err := filepath.EvalSymlinks(root.Path)
		if err == nil && canonical != root.Path {
			selections = append(selections, selection{path: canonical, root: root})
		}
	}
	allowedPath := func(path string) bool {
		inside := false
		for _, root := range selections {
			if !sourcePathWithin(path, root.path) {
				continue
			}
			inside = true
			if root.root.allowsPath(path, root.path) {
				return true
			}
		}
		return !inside
	}
	return func(path string) bool {
		return allowedPath(filepath.Clean(path)) && allowedPath(physicalSourcePath(path))
	}
}

func sourcePathWithin(path, directory string) bool {
	relative, err := filepath.Rel(directory, path)
	return err == nil && filepath.IsLocal(relative)
}

func physicalSourcePath(path string) string {
	canonical, err := filepath.EvalSymlinks(path)
	if err == nil {
		return canonical
	}
	return filepath.Clean(path)
}

func moduleFileSelection(directory string, filters []v1.ProtoFileFilter) func(string) bool {
	canonical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		canonical = directory
	}
	// Metadata is immutable for the lifetime of an effective graph.
	filters = slices.Clone(filters)
	for i, filter := range filters {
		filters[i].Includes = slices.Clone(filter.Includes)
		filters[i].Excludes = slices.Clone(filter.Excludes)
	}
	return func(path string) bool {
		base := directory
		if !sourcePathWithin(path, base) {
			base = canonical
		}
		relative, err := filepath.Rel(base, path)
		if err != nil || !filepath.IsLocal(relative) {
			return true
		}
		name := filepath.ToSlash(relative)
		inside := false
		within := func(directory string) bool {
			directory = filepath.ToSlash(directory)
			return directory == "." || name == directory || strings.HasPrefix(name, directory+"/")
		}
		for _, filter := range filters {
			if !within(filter.Root) {
				continue
			}
			inside = true
			included := len(filter.Includes) == 0 || slices.ContainsFunc(filter.Includes, within)
			if included && !slices.ContainsFunc(filter.Excludes, within) {
				return true
			}
		}
		return !inside
	}
}

// FileModules maps import paths to their owning modules for managed selectors.
// Callers check import collisions before using this mapping.
func (roots SourceRoots) FileModules() (map[string]string, error) {
	modules := make(map[string]string)
	allowed := roots.FileAllowed()
	for _, root := range roots {
		err := root.Walk(func(path string) error {
			if allowed != nil && !allowed(path) {
				return nil
			}
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return err
			}
			modules[filepath.ToSlash(relative)] = root.Module
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return modules, nil
}

// ModuleSources validates module roots and resolves them relative to its directory.
func ModuleSources(directory string, module v1.Module) (SourceRoots, error) {
	roots := make(SourceRoots, 0, len(module.Roots))
	var allowed func(string) bool
	if len(module.ProtoFilters) > 0 {
		allowed = moduleFileSelection(directory, module.ProtoFilters)
	}
	for _, root := range module.Roots {
		path := filepath.Join(directory, root)
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("Stat: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("module %s has invalid root %q: not a directory", module.Name, root)
		}
		roots = append(roots, SourceRoot{Path: path, Module: module.Name, fileAllowed: allowed})
	}
	return roots, nil
}
