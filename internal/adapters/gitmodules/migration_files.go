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
	regularFiles []string
	trackedFiles []string
	symlinks     []string
	hasSymlinks  bool
}

// migrationTrackedFiles records immutable Git modes for historical proofs.
// The current logical snapshot independently validates relevant alias targets.
func migrationTrackedFiles(ctx context.Context, checkout string) (migrationFiles, error) {
	entries, err := readV1GitIndex(ctx, checkout)
	if err != nil {
		return migrationFiles{}, fmt.Errorf("readV1GitIndex: %w", err)
	}
	tracked := migrationFiles{}
	for _, entry := range entries {
		tracked.trackedFiles = append(tracked.trackedFiles, entry.Path)
		switch entry.Mode {
		case gitindex.RegularFile, gitindex.ExecutableFile:
			tracked.regularFiles = append(tracked.regularFiles, entry.Path)
			continue
		case gitindex.Symlink:
			tracked.symlinks = append(tracked.symlinks, entry.Path)
			// A regular-only subset cannot prove an old whole installed tree.
			tracked.hasSymlinks = true
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
