package gitmodules

import (
	"context"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// snapshotPhysicalProtoFilters preserves selection in both logical root names
// and their resolved targets for Buf source filters.
func snapshotPhysicalProtoFilters(ctx context.Context, view *sourceview.View, filters []v1.ProtoFileFilter) []v1.ProtoFileFilter {
	result := slices.Clone(filters)
	for _, filter := range filters {
		resolved, err := view.Resolve(ctx, filepath.ToSlash(filter.Root))
		if err != nil || len(resolved.Links) == 0 {
			continue
		}
		physical := filter
		physical.Root = filepath.FromSlash(resolved.Path)
		translate := func(names []string) []string {
			translated := make([]string, 0, len(names))
			for _, name := range names {
				relative, err := filepath.Rel(filter.Root, name)
				if err != nil || !filepath.IsLocal(relative) {
					continue
				}
				translated = append(translated, filepath.Join(physical.Root, relative))
			}
			return translated
		}
		physical.Includes = translate(filter.Includes)
		physical.Excludes = translate(filter.Excludes)
		result = append(result, physical)
	}
	return result
}
