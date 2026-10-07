package modules

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// augmentV1ManifestRequirements records selected transitive modules. A module
// imported by a root .proto file is direct even if it arrived transitively.
func augmentV1ManifestRequirements(original []byte, root string, module v1.Module, lock v1.Lock, repository Cache) ([]byte, error) {
	imports, err := v1RootImports(root, module.Roots)
	if err != nil {
		return nil, fmt.Errorf("v1RootImports: %w", err)
	}
	owners := make(map[string]bool)
	for _, entry := range lock.Modules {
		installDir, dependency, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("Cached: %w", err)
		}
		for _, depRoot := range dependency.Roots {
			base := filepath.Join(installDir, depRoot)
			for importPath := range imports {
				if !filepath.IsLocal(importPath) {
					continue
				}
				info, err := os.Stat(filepath.Join(base, filepath.FromSlash(importPath)))
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return nil, fmt.Errorf("Stat: %w", err)
				}
				if info.Mode().IsRegular() {
					owners[entry.Source] = true
				}
			}
		}
	}
	return augmentV1ManifestRequirementsWithOwners(original, module, lock, repository, owners)
}

// augmentV1ManifestRequirementsWithOwners accepts ownership proved from the
// proposed consumer view, so renamed imports can promote transitive modules.
func augmentV1ManifestRequirementsWithOwners(original []byte, module v1.Module, lock v1.Lock, repository Cache, owners map[string]bool) ([]byte, error) {
	lines := strings.Split(string(original), "\n")
	indirect := make(map[string]v1RequirementLine)
	for _, line := range parseV1RequirementLines(lines) {
		if line.indirect() {
			indirect[line.module] = line
		}
	}
	existing := make(map[string]bool, len(module.Requires))
	for _, requirement := range module.Requires {
		existing[requirement.Module] = true
	}
	versions, err := manifestRequirementVersions(lock, repository)
	if err != nil {
		return nil, fmt.Errorf("manifestRequirementVersions: %w", err)
	}
	var additions []v1ManifestRequirement
	for _, entry := range lock.Modules {
		if _, derived := indirect[entry.Source]; existing[entry.Source] && !derived {
			continue
		}
		direct := owners[entry.Source]
		if line, ok := indirect[entry.Source]; ok {
			if !existing[entry.Source] {
				line = line.withVersion(versions[entry.Source])
			}
			if direct {
				line = line.direct()
			}
			lines[line.index] = line.String()
			delete(indirect, entry.Source)
			continue
		}
		additions = append(additions, v1ManifestRequirement{Requirement: v1.Requirement{Module: entry.Source, Version: versions[entry.Source]}, indirect: !direct})
	}
	if len(indirect) > 0 {
		stale := make(map[int]bool, len(indirect))
		for _, line := range indirect {
			stale[line.index] = true
		}
		kept := make([]string, 0, len(lines)-len(stale))
		for index, line := range lines {
			if !stale[index] {
				kept = append(kept, line)
			}
		}
		lines = kept
	}
	updated := appendV1Requirements([]byte(strings.Join(lines, "\n")), additions)
	if _, err := v1.ParseModule(bytes.NewReader(updated)); err != nil {
		return nil, fmt.Errorf("ParseModule: %w", err)
	}
	return updated, nil
}

func v1RootImports(moduleDir string, roots []string) (map[string]bool, error) {
	imports := make(map[string]bool)
	for _, root := range roots {
		sourceRoot := SourceRoot{Path: filepath.Join(moduleDir, root), directory: moduleDir}
		err := sourceRoot.Walk(func(path string) error {
			raw, err := (SourceRoots{sourceRoot}).ReadSourceFile(path)
			if err != nil {
				return fmt.Errorf("ReadSourceFile: %w", err)
			}
			paths, err := ParseProtoImports(path, raw)
			if err != nil {
				return fmt.Errorf("ReadProtoImports: %w", err)
			}
			for _, importPath := range paths {
				imports[importPath] = true
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	return imports, nil
}
