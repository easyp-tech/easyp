package generation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/protosource"
)

// selectedPackageFiles reads package declarations without compiling unrelated
// files. Malformed selected files and required imports are rejected by the real
// compiler later; an unrelated broken declaration is not a generation target.
func selectedPackageFiles(ctx context.Context, selected v1GenerationModule, packages []string) ([]string, map[string]bool, error) {
	roots, err := modules.ModuleSources(selected.directory, selected.module)
	if err != nil {
		return nil, nil, fmt.Errorf("ModuleSources: %w", err)
	}
	requested := make(map[string]bool, len(packages))
	for _, name := range packages {
		requested[name] = true
	}
	matched := make(map[string]bool)
	seen := make(map[string]bool)
	var files []string
	for _, root := range roots {
		err := root.Walk(func(path string) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("Err: %w", err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("ReadFile: %w", err)
			}
			name := protosource.Package(raw)
			if !requested[name] {
				return nil
			}
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			relative = filepath.ToSlash(relative)
			matched[name] = true
			if !seen[relative] {
				seen[relative] = true
				files = append(files, relative)
			}
			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	slices.Sort(files)
	return files, matched, nil
}
