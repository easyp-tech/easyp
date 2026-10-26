package modules

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)

// snapshotFileSource exposes names captured while verifying installed contents.
// Names are relative to the installed module directory and include metadata.
// Local sources never implement or use this immutable snapshot capability.
type snapshotFileSource interface {
	SourceFiles(v1.LockedModule) ([]string, error)
}

type sourceFileIndex struct {
	paths        []string
	names        map[string]bool
	physicalRoot string
}

// WalkSnapshotFiles visits the verified regular proto inventory for importRoot.
// It returns false without visiting anything when that root is a local source
// or its cache did not supply an immutable inventory.
func (roots SourceRoots) WalkSnapshotFiles(importRoot string, visit func(string) error) (bool, error) {
	for _, root := range roots {
		if root.index == nil || filepath.Clean(root.Path) != filepath.Clean(importRoot) {
			continue
		}
		for _, path := range root.index.paths {
			if err := visit(path); err != nil {
				return true, err
			}
		}
		return true, nil
	}
	return false, nil
}

func cachedModuleSources(directory string, module v1.Module, entry v1.LockedModule, repository Cache) (SourceRoots, error) {
	roots, err := ModuleSources(directory, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	indexed, ok := repository.(snapshotFileSource)
	if !ok {
		return roots, nil
	}
	files, err := indexed.SourceFiles(entry)
	if err != nil {
		return nil, fmt.Errorf("SourceFiles: %w", err)
	}
	if files == nil {
		return roots, nil
	}
	return indexSnapshotSources(directory, roots, files)
}

func indexSnapshotSources(directory string, roots SourceRoots, files []string) (SourceRoots, error) {
	boundaries := make(map[string]bool)
	var protos []string
	for _, name := range files {
		name = filepath.FromSlash(name)
		if !filepath.IsLocal(name) || filepath.Clean(name) != name || name == "." {
			return nil, fmt.Errorf("invalid snapshot file %q", name)
		}
		path := filepath.Join(directory, name)
		switch filepath.Base(name) {
		case ".git", v1.ModuleFile, "buf.work.yaml", "buf.yaml":
			boundaries[filepath.Dir(path)] = true
		}
		if filepath.Ext(name) == ".proto" {
			protos = append(protos, path)
		}
	}
	slices.SortFunc(protos, compareSourcePaths)
	for index := range roots {
		root := &roots[index]
		catalog := &sourceFileIndex{names: make(map[string]bool), physicalRoot: physicalSourcePath(root.Path)}
		for _, path := range protos {
			if !sourcePathWithin(path, root.Path) || path_helpers.HiddenOrVendorSourcePath(root.Path, filepath.Dir(path)) ||
				(root.fileAllowed != nil && !root.fileAllowed(path)) {
				continue
			}
			nested := false
			for parent := filepath.Dir(path); parent != filepath.Clean(root.Path); parent = filepath.Dir(parent) {
				if boundaries[parent] {
					nested = true
					break
				}
			}
			if nested {
				continue
			}
			name, err := filepath.Rel(root.Path, path)
			if err != nil {
				return nil, fmt.Errorf("Rel: %w", err)
			}
			catalog.paths = append(catalog.paths, path)
			catalog.names[name] = true
		}
		root.index = catalog
	}
	return roots, nil
}

// Directory walks sort each component, so a/child.proto precedes a.proto.
// Preserve that order when an immutable inventory replaces filesystem walks.
func compareSourcePaths(first, second string) int {
	for {
		a, remainingA, nestedA := strings.Cut(first, string(filepath.Separator))
		b, remainingB, nestedB := strings.Cut(second, string(filepath.Separator))
		if compared := strings.Compare(a, b); compared != 0 {
			return compared
		}
		if !nestedA || !nestedB {
			return strings.Compare(remainingA, remainingB)
		}
		first, second = remainingA, remainingB
	}
}
