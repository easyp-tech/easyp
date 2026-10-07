package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func inspectV1Snapshot(ctx context.Context, view *sourceview.View, directory string, module v1.Module, commit string) (*modules.RootInspection, error) {
	inspection := &modules.RootInspection{}
	err := view.Walk(ctx, ".", func(logical string, resolved sourceview.Resolution, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("Err: %w", err)
		}
		if walkErr != nil {
			inspection.Problems = append(inspection.Problems, modules.RootPathProblem{Path: logical, Err: walkErr})
			return nil
		}
		if resolved.Info.IsDir() {
			selectedAncestor := slices.ContainsFunc(module.Roots, func(root string) bool { return v1FileWithin(root, logical) })
			if logical != "." && !selectedAncestor && snapshotExcludedDirectory(logical, ".", directory, module) {
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
