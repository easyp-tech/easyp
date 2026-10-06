package gitmodules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
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
	legacy := make(map[string][]byte)
	for _, name := range files {
		if path.Ext(name) == ".proto" {
			content, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("dependency %s source selection omits legacy file %q; manual migration is required: %w", module.Name, name, err)
				}
				return fmt.Errorf("ReadFile: %w", err)
			}
			importName := renameMigrationLegacyFile(name, roots)
			if _, exists := legacy[importName]; exists {
				return fmt.Errorf("dependency %s source selection collides at %q; manual migration is required", module.Name, importName)
			}
			legacy[importName] = content
		}
	}
	selected, err := migrationSelectedFiles(checkout, module)
	if err != nil {
		return fmt.Errorf("migrationSelectedFiles: %w", err)
	}
	return validateMigrationSourceContents(module.Name, legacy, selected)
}

func migrationUsesLogicalAliases(checkout v1ModuleCheckout, tracked migrationFiles) bool {
	for _, name := range tracked.symlinks {
		if moduleconfig.IsGitDependencyConfigFile(path.Base(name)) {
			return true
		}
		if _, found := migrationRootThroughSymlink(name, checkout.module.Roots); found {
			return true
		}
		info, err := os.Stat(filepath.Join(checkout.snapshot, filepath.FromSlash(name)))
		if err != nil {
			continue
		}
		if info.Mode().IsRegular() && path.Ext(name) == ".proto" && snapshotSelectedAlias(name, checkout.module) {
			return true
		}
		if info.IsDir() {
			for _, root := range checkout.module.Roots {
				if v1FileWithin(name, root) {
					return true
				}
			}
		}
	}
	return false
}

// New acquisitions may introduce usable logical roots that old installers could
// not represent. Validate their current import ownership, and require every raw
// legacy proto to remain covered by a selected logical source. Auxiliary aliases
// never exempt unrelated outside-root, hidden, or nested-module source files.
func validateMigrationLogicalOwnership(ctx context.Context, checkout v1ModuleCheckout, tracked migrationFiles) error {
	_, err := migrationSelectedFiles(checkout.snapshot, checkout.module)
	if err != nil {
		return fmt.Errorf("migrationSelectedFiles: %w", err)
	}
	view, err := sourceV1SnapshotView(ctx, checkout.dir, checkout.commit)
	if err != nil {
		return fmt.Errorf("sourceV1SnapshotView: %w", err)
	}
	sources, err := modules.ModuleSources(checkout.snapshot, checkout.module)
	if err != nil {
		return fmt.Errorf("ModuleSources: %w", err)
	}
	covered := make(map[string]bool)
	for _, source := range sources {
		err = modules.WalkProtoFiles(source.Path, func(file string) error {
			logical, err := filepath.Rel(checkout.snapshot, file)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			resolved, err := view.Resolve(ctx, filepath.ToSlash(logical))
			if err != nil {
				return fmt.Errorf("Resolve: %w", err)
			}
			covered[resolved.Path] = true
			return nil
		})
		if err != nil {
			return fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	for _, name := range tracked.regularFiles {
		if strings.HasSuffix(name, ".proto") && !covered[name] {
			return fmt.Errorf("dependency %s source selection omits legacy file %q; manual migration is required", checkout.module.Name, name)
		}
	}
	return nil
}

func migrationSelectedFiles(checkout string, module v1.Module) (map[string][]byte, error) {
	sources, err := modules.ModuleSources(checkout, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	selected := make(map[string][]byte)
	owners := make(map[string]string)
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
			if previous, exists := owners[importName]; exists && previous != physicalName {
				return fmt.Errorf("dependency %s source selection collides at %q; manual migration is required", module.Name, importName)
			}
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("ReadFile: %w", err)
			}
			owners[importName], selected[importName] = physicalName, content
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	return selected, nil
}

func validateMigrationSourceContents(module string, legacy, selected map[string][]byte) error {
	// Stable ordering makes the first actionable mismatch reproducible.
	names := make([]string, 0, len(legacy)+len(selected))
	for name := range legacy {
		names = append(names, name)
	}
	for name := range selected {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range slices.Compact(names) {
		old, oldExists := legacy[name]
		current, currentExists := selected[name]
		if oldExists != currentExists || !bytes.Equal(old, current) {
			return fmt.Errorf("dependency %s source selection differs at import %q (legacy present %t, v1 present %t, same contents %t); manual migration is required", module, name, oldExists, currentExists, bytes.Equal(old, current))
		}
	}
	return nil
}
