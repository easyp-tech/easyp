package gitmodules

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/gitindex"
)

// migrationFiles records the snapshot and the historical proof it requires.
type migrationFiles struct {
	regularFiles             []string
	omittedAuxiliarySymlinks bool
}

// Migration uses native regular-file selection, but additionally refuses links
// that could change legacy proto import names or configured root boundaries.
func migrationTrackedFiles(ctx context.Context, checkout string, roots ...string) (migrationFiles, error) {
	entries, err := readV1GitIndex(ctx, checkout)
	if err != nil {
		return migrationFiles{}, fmt.Errorf("readV1GitIndex: %w", err)
	}
	files, err := regularTrackedV1Files(checkout, entries)
	if err != nil {
		return migrationFiles{}, fmt.Errorf("regularTrackedV1Files: %w", err)
	}
	tracked := migrationFiles{regularFiles: files}
	for _, entry := range entries {
		switch entry.Mode {
		case gitindex.RegularFile, gitindex.ExecutableFile:
			continue
		case gitindex.Symlink:
			if path.Ext(entry.Path) == ".proto" {
				return migrationFiles{}, fmt.Errorf("unsupported non-regular Git mode in %q", entry.Raw)
			}
			if root, found := migrationRootThroughSymlink(entry.Path, roots); found {
				return migrationFiles{}, fmt.Errorf("unsupported non-regular Git mode in %q used by root %q", entry.Raw, root)
			}
			// Metadata links were rejected during regular-file selection.
			// Auxiliary target bytes never enter the snapshot or its digest.
			tracked.omittedAuxiliarySymlinks = true
		default:
			// Native snapshots omit submodules, but migration cannot prove
			// their legacy contracts from this repository's regular files.
			return migrationFiles{}, fmt.Errorf("unsupported non-regular Git mode in %q", entry.Raw)
		}
	}
	return tracked, nil
}

func migrationRootThroughSymlink(name string, roots []string) (string, bool) {
	for _, root := range roots {
		root = path.Clean(filepath.ToSlash(root))
		if root == name || strings.HasPrefix(root, name+"/") {
			return root, true
		}
	}
	return "", false
}
