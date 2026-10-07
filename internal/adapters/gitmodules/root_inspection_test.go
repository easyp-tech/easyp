package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestRootInspectionGitlinkBoundaries(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		alias     string
		target    string
		wantCause error
	}{
		{name: "opaque_direct_gitlink"},
		{name: "file_alias_into_gitlink", alias: "alias.proto", target: "third_party/external/file.proto", wantCause: gitsnapshot.ErrGitlink},
		{name: "directory_alias_into_gitlink", alias: "alias", target: "third_party/external", wantCause: gitsnapshot.ErrGitlink},
		{name: "unsafe_alias", alias: "outside.proto", target: "../outside.proto", wantCause: sourceview.ErrOutsideRoot},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(repository, "file.proto"), []byte(`syntax = "proto3";`), 0o644))
			if tt.alias != "" {
				require.NoError(t, os.Symlink(tt.target, filepath.Join(repository, tt.alias)))
			}
			runTestGit(t, repository, "init", "-q")
			runTestGit(t, repository, "add", ".")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
			commit := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
			runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "160000,"+commit+",third_party/external")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "opaque submodule")
			commit = strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
			cache := New(t.TempDir())

			fetched, err := cache.FetchForRootResolution(t.Context(), repository, commit)

			require.NoError(t, err)
			require.NotNil(t, fetched.Inspection)
			if tt.wantCause == nil {
				assert.Empty(t, fetched.Inspection.Problems)
			} else {
				require.Len(t, fetched.Inspection.Problems, 1)
				assert.Equal(t, tt.alias, fetched.Inspection.Problems[0].Path)
				assert.ErrorIs(t, fetched.Inspection.Problems[0].Err, tt.wantCause)
			}

			_, err = cache.FetchWithRoots(t.Context(), repository, commit, []string{"."})

			if tt.wantCause == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantCause)
			}
			// An explicit root that enters the submodule still fails.
			_, err = cache.FetchWithRoots(t.Context(), repository, commit, []string{"third_party/external"})
			require.ErrorIs(t, err, gitsnapshot.ErrGitlink)
		})
	}
}
