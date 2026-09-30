package gitmodules

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// The normal dependency adapter reads legacy requirements but drops local
// replacements. Refuse them before trusting that adapted dependency graph,
// including nested declarations whose effect cannot be established safely.
func validateMigrationLegacyReplacements(checkout string, files []string) error {
	for _, name := range files {
		if path.Base(name) != v1.ModuleFile {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		if v1.IsModuleManifest(raw) {
			continue
		}
		manifest, err := modfile.Parse(raw)
		if err != nil {
			return fmt.Errorf("Parse: %w", err)
		}
		if len(manifest.Replace) > 0 {
			return fmt.Errorf("dependency legacy manifest %q has local replacements; full lock migration requires manual migration", name)
		}
	}
	return nil
}

func hasNativeMigrationModule(checkout string, files []string, source string) (bool, error) {
	for _, name := range files {
		if path.Base(name) != v1.ModuleFile {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
		if err != nil {
			return false, fmt.Errorf("ReadFile: %w", err)
		}
		if !v1.IsModuleManifest(raw) {
			continue
		}
		module, err := v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return false, fmt.Errorf("ParseModule: %w", err)
		}
		if module.Name == source {
			return true, nil
		}
	}
	return false, nil
}

// validateMigrationSelection checks both default Git generation inputs and the
// import namespace available to consumers. Equal tree hashes alone do not prove
// equivalence: v1 roots can omit legacy files or introduce additional aliases.
func validateMigrationSelection(checkout string, files []string, module v1.Module) error {
	roots, err := readMigrationLegacyRoots(checkout, files)
	if err != nil {
		return fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}
	legacy := make(map[string]string)
	for _, name := range files {
		if path.Ext(name) == ".proto" {
			legacy[renameMigrationLegacyFile(name, roots)] = name
		}
	}
	sources, err := modules.ModuleSources(checkout, module)
	if err != nil {
		return fmt.Errorf("ModuleSources: %w", err)
	}
	selected := make(map[string]string)
	for _, source := range sources {
		err := modules.WalkProtoFiles(source.Path, func(file string) error {
			importName, err := filepath.Rel(source.Path, file)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			physicalName, err := filepath.Rel(checkout, file)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			importName, physicalName = filepath.ToSlash(importName), filepath.ToSlash(physicalName)
			if previous, exists := selected[importName]; exists && previous != physicalName {
				return fmt.Errorf("dependency %s source selection collides at %q; manual migration is required", module.Name, importName)
			}
			selected[importName] = physicalName
			return nil
		})
		if err != nil {
			return fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	// Stable ordering makes the first actionable mismatch reproducible.
	names := make([]string, 0, len(legacy)+len(selected))
	for name := range legacy {
		names = append(names, name)
	}
	for name := range selected {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if legacy[name] != selected[name] {
			return fmt.Errorf("dependency %s source selection differs at import %q (legacy file %q, v1 file %q); manual migration is required", module.Name, name, legacy[name], selected[name])
		}
	}
	return nil
}
