package migration

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// RootsRepository supplies verified legacy paths and pinned proto bytes for an
// explicit Git selection. Non-nil empty roots retain authoritative/intrinsic roots.
type RootsRepository interface {
	FetchMigrationWithRoots(context.Context, string, string, string, []string) (modules.Fetched, error)
}

type gitModuleSelection struct {
	name, root string
	subdirs    []string
	index      int
}

func (s gitModuleSelection) needsProof() bool {
	return s.root != "." || !slices.Contains(s.subdirs, ".")
}

func (s gitModuleSelection) roots() []string {
	if s.root == "." {
		return []string{}
	}
	return []string{s.root}
}

func planGitSelections(cfg legacyConfig, local string, hasLocal bool, packages, paths []string, deps *requirements) ([]v1.GenerateModule, map[string]gitModuleSelection, error) {
	selected := make(map[string]gitModuleSelection)
	var order []string
	for index, input := range cfg.Generate.Inputs {
		if input.GitRepo == nil {
			continue
		}
		git := *input.GitRepo
		root, err := literalGitPath(git.Root)
		if err != nil {
			return nil, nil, fmt.Errorf("generate.inputs[%d].git_repo.root: %w", index, err)
		}
		subdir, err := literalGitPath(git.SubDirectory)
		if err != nil {
			return nil, nil, fmt.Errorf("generate.inputs[%d].git_repo.sub_directory: %w", index, err)
		}
		if err := deps.add(git.URL, false); err != nil {
			return nil, nil, fmt.Errorf("add: %w", err)
		}
		name, _, _ := strings.Cut(git.URL, "@")
		previous, exists := selected[name]
		if exists && previous.root != root {
			return nil, nil, fmt.Errorf("generate.inputs[%d].git_repo %s root %q conflicts with input %d root %q; one module cannot preserve both import namespaces", index, name, root, previous.index, previous.root)
		}
		if !exists {
			previous = gitModuleSelection{name: name, root: root, index: index}
			order = append(order, name)
		}
		if !slices.Contains(previous.subdirs, subdir) {
			previous.subdirs = append(previous.subdirs, subdir)
		}
		selected[name] = previous
	}
	var entries []v1.GenerateModule
	if len(order) > 0 && hasLocal {
		entries = append(entries, v1.GenerateModule{Module: local, Packages: slices.Clone(packages), Paths: slices.Clone(paths)})
	}
	for _, name := range order {
		selection := selected[name]
		entry := v1.GenerateModule{Module: name}
		if !slices.Contains(selection.subdirs, ".") {
			entry.Paths = minimalGitPaths(selection.subdirs)
		}
		entries = append(entries, entry)
	}
	return entries, selected, nil
}

func literalGitPath(value string) (string, error) {
	if value == "" {
		return ".", nil
	}
	if strings.Contains(value, "$") {
		return "", fmt.Errorf("git path placeholder %q prevents safe source analysis; supply a bounded literal path", value)
	}
	for _, component := range strings.Split(value, "/") {
		if component == ".." {
			return "", fmt.Errorf("git path %q leaves its installed module; manual migration is required", value)
		}
	}
	clean := path.Clean(value)
	if err := v1.ValidatePathSelectors([]string{clean}); err != nil {
		return "", fmt.Errorf("ValidatePathSelectors: %w", err)
	}
	return clean, nil
}

func minimalGitPaths(paths []string) []string {
	result := slices.Clone(paths)
	slices.Sort(result)
	result = slices.Compact(result)
	return slices.DeleteFunc(result, func(name string) bool {
		return slices.ContainsFunc(paths, func(parent string) bool { return name != parent && v1.PathSelectorMatches(parent, name) })
	})
}

type gitSelectionProof struct {
	selections map[string]gitModuleSelection
	fetched    map[string]modules.Fetched
	entries    []v1.GenerateModule
	localName  string
	bindings   map[string]migrationProtoSource
}

func proveGitSelections(selections map[string]gitModuleSelection, fetched map[string]modules.Fetched, entries []v1.GenerateModule, local string) (*gitSelectionProof, error) {
	proof := &gitSelectionProof{selections: selections, fetched: cloneMigrationFetched(fetched), entries: slices.Clone(entries), localName: local}
	for index, entry := range proof.entries {
		selection, ok := selections[entry.Module]
		if !ok {
			continue
		}
		dependency, ok := proof.fetched[entry.Module]
		if !ok {
			return nil, fmt.Errorf("generate.inputs[%d].git_repo %s has no verified historical source", selection.index, selection.name)
		}
		if dependency.Inspection == nil || dependency.Inspection.LegacyFiles == nil {
			if selection.needsProof() {
				return nil, fmt.Errorf("generate.inputs[%d].git_repo %s requires verified legacy source mappings from the roots-aware migration repository", selection.index, selection.name)
			}
			continue
		}
		if err := validateGitSelectionRoots(selection, dependency); err != nil {
			return nil, fmt.Errorf("validateGitSelectionRoots: %w", err)
		}
		paths, err := proveGitModuleSelection(selection, dependency)
		if err != nil {
			return nil, fmt.Errorf("generate.inputs[%d].git_repo %s root %q sub_directory %v: %w", selection.index, selection.name, selection.root, selection.subdirs, err)
		}
		proof.entries[index].Paths = paths
	}
	return proof, nil
}

func validateGitSelectionRoots(selection gitModuleSelection, dependency modules.Fetched) error {
	canonical := func(roots []string) []string {
		roots = slices.Clone(roots)
		slices.Sort(roots)
		return slices.Compact(roots)
	}
	roots := canonical(dependency.Module.Roots)
	if selection.root != "." && !slices.Equal(roots, []string{selection.root}) {
		return fmt.Errorf("generate.inputs[%d].git_repo %s requested root %q, but its verified module roots are %v; a root hint cannot be ignored or override authoritative metadata", selection.index, selection.name, selection.root, roots)
	}
	if !dependency.Module.RootsFromMetadata && (selection.root != "." || !slices.Equal(roots, []string{"."})) && !slices.Equal(roots, canonical(dependency.Lock.Roots)) {
		return fmt.Errorf("dependency %s verified fallback roots %v must be recorded in protobuf.lock for cold/frozen replay", selection.name, roots)
	}
	return nil
}

func cloneMigrationFetched(fetched map[string]modules.Fetched) map[string]modules.Fetched {
	result := maps.Clone(fetched)
	for name, dependency := range result {
		dependency.Module.Roots = slices.Clone(dependency.Module.Roots)
		dependency.Lock.Roots = slices.Clone(dependency.Lock.Roots)
		if dependency.Inspection != nil {
			inspection := *dependency.Inspection
			inspection.Files = slices.Clone(inspection.Files)
			inspection.Problems = slices.Clone(inspection.Problems)
			inspection.LegacyFiles = maps.Clone(inspection.LegacyFiles)
			for index := range inspection.Files {
				inspection.Files[index].Content = bytes.Clone(inspection.Files[index].Content)
			}
			dependency.Inspection = &inspection
		}
		result[name] = dependency
	}
	return result
}

func proveGitModuleSelection(selection gitModuleSelection, dependency modules.Fetched) ([]string, error) {
	files := make(map[string]modules.RootProtoFile)
	for _, file := range dependency.Inspection.Files {
		files[file.Path] = file
	}
	legacy := make(map[string]modules.RootProtoFile)
	var paths []string
	for _, subdir := range selection.subdirs {
		matched := false
		for _, installed := range slices.Sorted(maps.Keys(dependency.Inspection.LegacyFiles)) {
			if !v1.PathSelectorMatches(subdir, installed) {
				continue
			}
			matched = true
			logical := dependency.Inspection.LegacyFiles[installed]
			file, ok := files[logical]
			if !ok {
				return nil, fmt.Errorf("legacy source selection %q resolves to %q outside the v1 source scope; manual migration is required", installed, logical)
			}
			name := installed
			if selection.root != "." {
				name = strings.TrimPrefix(name, selection.root+"/")
			}
			if previous, exists := legacy[name]; exists && previous.Identity != file.Identity {
				return nil, fmt.Errorf("legacy source selection collides at filename %q (%s and %s)", name, previous.Path, file.Path)
			}
			legacy[name] = file
			if subdir == "." {
				continue
			}
			relative := strings.TrimPrefix(installed, subdir+"/")
			physicalSelector := logical
			if installed != subdir {
				physicalSelector = strings.TrimSuffix(logical, "/"+relative)
				if physicalSelector == logical {
					return nil, fmt.Errorf("legacy sub_directory %q cannot be translated through %q -> %q", subdir, installed, logical)
				}
			}
			paths = append(paths, physicalSelector)
		}
		if !matched && subdir != "." {
			return nil, fmt.Errorf("legacy sub_directory %q did not match any verified .proto source; an empty selector would widen generation scope", subdir)
		}
	}
	if slices.Contains(selection.subdirs, ".") {
		paths = nil
	} else {
		paths = minimalGitPaths(paths)
	}
	if err := v1.ValidatePathSelectors(paths); err != nil {
		return nil, fmt.Errorf("ValidatePathSelectors: %w", err)
	}
	current, err := migrationRootNamespace(dependency)
	if err != nil {
		return nil, fmt.Errorf("migrationRootNamespace: %w", err)
	}
	current = maps.Clone(current)
	if len(paths) > 0 {
		maps.DeleteFunc(current, func(_ string, file modules.RootProtoFile) bool {
			return !slices.ContainsFunc(paths, func(selector string) bool { return v1.PathSelectorMatches(selector, file.Path) })
		})
	}
	if err := compareGitSourceMaps(selection.name, legacy, current); err != nil {
		return nil, fmt.Errorf("compareGitSourceMaps: %w", err)
	}
	return paths, nil
}

func migrationRootNamespace(dependency modules.Fetched) (map[string]modules.RootProtoFile, error) {
	result := make(map[string]modules.RootProtoFile)
	for _, root := range dependency.Module.Roots {
		for _, file := range dependency.Inspection.Files {
			if !v1.PathSelectorMatches(root, file.Path) || file.Path == root {
				continue
			}
			name := file.Path
			if root != "." {
				name = strings.TrimPrefix(name, root+"/")
			}
			if previous, exists := result[name]; exists && (previous.Path != file.Path || previous.Identity != file.Identity) {
				return nil, fmt.Errorf("dependency %s filename %q collides between %q and %q", dependency.Module.Name, name, previous.Path, file.Path)
			}
			result[name] = file
		}
	}
	return result, nil
}

func compareGitSourceMaps(module string, legacy, current map[string]modules.RootProtoFile) error {
	names := append(slices.Sorted(maps.Keys(legacy)), slices.Sorted(maps.Keys(current))...)
	slices.Sort(names)
	for _, name := range slices.Compact(names) {
		old, oldExists := legacy[name]
		file, exists := current[name]
		if oldExists != exists || old.Identity != file.Identity || old.Path != file.Path || !bytes.Equal(old.Content, file.Content) {
			return fmt.Errorf("dependency %s source selection differs at filename %q (legacy %q, v1 %q); generation targets, import names and source bindings must match; manual migration is required", module, name, old.Path, file.Path)
		}
	}
	return nil
}
