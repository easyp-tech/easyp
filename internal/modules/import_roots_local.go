package modules

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func inspectLocalRootCandidates(ctx context.Context, directory string, module v1.Module, hinted bool) (importRootModule, error) {
	if module.RootsFromMetadata || hinted {
		inspected, err := inspectLocalImportRoots(directory, module)
		return inspected, err
	}
	inspection := &RootInspection{}
	err := sourceview.WalkLocal(ctx, directory, ".", func(logical string, resolved sourceview.Resolution, walkErr error) error {
		path := filepath.Join(directory, filepath.FromSlash(logical))
		if path_helpers.ShouldSkipV1SourceDir(directory, path) {
			return fs.SkipDir
		}
		if walkErr != nil {
			inspection.Problems = append(inspection.Problems, RootPathProblem{Path: logical, Err: walkErr})
			return nil
		}
		if resolved.Info.IsDir() || filepath.Ext(logical) != ".proto" {
			return nil
		}
		physical := filepath.Join(directory, filepath.FromSlash(resolved.Path))
		if path_helpers.ShouldSkipV1SourceDir(directory, filepath.Dir(physical)) {
			return nil
		}
		content, err := sourceview.ReadLocal(ctx, directory, logical)
		if err != nil {
			return fmt.Errorf("ReadLocal: %w", err)
		}
		inspection.Files = append(inspection.Files, RootProtoFile{Path: logical, Identity: physicalSourcePath(physical), Content: content})
		return nil
	})
	if err != nil {
		return importRootModule{}, fmt.Errorf("WalkLocal: %w", err)
	}
	return importRootModule{name: module.Name, roots: module.Roots, inspection: inspection}, nil
}
