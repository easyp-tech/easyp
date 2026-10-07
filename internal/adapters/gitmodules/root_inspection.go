package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func inspectV1Snapshot(ctx context.Context, view *sourceview.View, directory string, module v1.Module, commit string) (*modules.RootInspection, error) {
	var boundaries []inspectionSourceBoundary
	for _, root := range module.Roots {
		logical := filepath.ToSlash(root)
		resolved, err := view.Resolve(ctx, logical)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && len(resolved.Links) == 0 {
				continue
			}
			return nil, fmt.Errorf("Resolve: %w", err)
		}
		boundaries = append(boundaries, inspectionSourceBoundary{logical: logical, physical: resolved.Path})
	}
	inspection := &modules.RootInspection{}
	err := view.Walk(ctx, ".", func(logical string, resolved sourceview.Resolution, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("Err: %w", err)
		}
		if walkErr != nil {
			// Match strict snapshots: unentered submodules are opaque, while
			// aliases attempting to enter them remain source path failures.
			if errors.Is(walkErr, gitsnapshot.ErrGitlink) && len(resolved.Links) == 0 {
				return nil
			}
			inspection.Problems = append(inspection.Problems, modules.RootPathProblem{Path: logical, Err: walkErr})
			return nil
		}
		if resolved.Info.IsDir() {
			if inspectionSourceDirectoryExcluded(directory, logical, resolved.Path, boundaries) {
				return fs.SkipDir
			}
			return nil
		}
		if path.Ext(logical) != ".proto" || len(selectV1ProtoFiles([]string{logical}, module.ProtoFilters)) == 0 {
			return nil
		}
		file, err := view.Open(ctx, logical)
		if err != nil {
			return fmt.Errorf("Open: %w", err)
		}
		content, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			readErr = fmt.Errorf("ReadAll: %w", readErr)
		}
		if closeErr != nil {
			closeErr = fmt.Errorf("Close: %w", closeErr)
		}
		if err := errors.Join(readErr, closeErr); err != nil {
			return err
		}
		inspection.Files = append(inspection.Files, modules.RootProtoFile{
			Path: logical, Identity: module.Name + "@" + commit + ":" + resolved.Path, Content: content,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Walk: %w", err)
	}
	return inspection, nil
}

type inspectionSourceBoundary struct {
	logical  string
	physical string
}

func inspectionSourceDirectoryExcluded(directory, logical, physical string, boundaries []inspectionSourceBoundary) bool {
	var owners []inspectionSourceBoundary
	for _, boundary := range boundaries {
		if v1FileWithin(boundary.logical, logical) {
			return false
		}
		if v1FileWithin(logical, boundary.logical) {
			owners = append(owners, boundary)
		}
	}
	if len(owners) == 0 {
		owners = []inspectionSourceBoundary{{logical: ".", physical: "."}}
	}
	skip := func(root, name string) bool {
		rootPath := filepath.Join(directory, filepath.FromSlash(root))
		namePath := filepath.Join(directory, filepath.FromSlash(name))
		return path_helpers.HiddenOrVendorSourcePath(rootPath, namePath) || path_helpers.ShouldSkipV1SourceDir(rootPath, namePath)
	}
	for _, owner := range owners {
		if !skip(owner.logical, logical) && !skip(owner.physical, physical) {
			return false
		}
	}
	return true
}
