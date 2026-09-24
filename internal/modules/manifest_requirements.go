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
	imports, err := v1RootImports(root, module.Roots)
	if err != nil {
		return nil, fmt.Errorf("v1RootImports: %w", err)
	}
	versions, err := manifestRequirementVersions(lock, repository)
	if err != nil {
		return nil, fmt.Errorf("manifestRequirementVersions: %w", err)
	}
	var additions []v1ManifestRequirement
	for _, entry := range lock.Modules {
		if existing[entry.Source] {
			delete(indirect, entry.Source)
			continue
		}
		installDir, dependency, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("Cached: %w", err)
		}
		direct := false
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
					direct = true
					break
				}
			}
			if direct {
				break
			}
		}
		if line, ok := indirect[entry.Source]; ok {
			line = line.withVersion(versions[entry.Source])
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
		sourceRoot := filepath.Join(moduleDir, root)
		err := WalkProtoFiles(sourceRoot, func(path string) error {
			paths, err := ReadProtoImports(path)
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
