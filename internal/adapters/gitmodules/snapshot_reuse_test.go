package gitmodules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestFetchRetainsSnapshotForOfflineInstall(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, version string
		commit        bool
	}{
		{name: "head"},
		{name: "tag", version: "v1.0.0"},
		{name: "commit", commit: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote, commit := diagnosticRepository(t)
			version := tt.version
			if tt.commit {
				version = commit
			}
			cache := New(t.TempDir())
			fetched, err := cache.Fetch(t.Context(), remote, version)
			require.NoError(t, err)
			directory, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			before, err := os.Stat(filepath.Join(directory, "api.proto"))
			require.NoError(t, err)
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
			require.NoError(t, os.Rename(remote, remote+"-offline"))
			require.NoError(t, cache.Install(t.Context(), lock))
			assert.NoFileExists(t, cacheVerificationStampPath(directory), "publication must not attest a fingerprint collected after hashing")
			after, err := os.Stat(filepath.Join(directory, "api.proto"))
			require.NoError(t, err)
			assert.True(t, os.SameFile(before, after), "Install must retain the already materialized file")
			require.NoError(t, (&Cache{root: cache.root}).VerifyCached(t.Context(), lock))

			// Retained local objects also reconstruct the same pinned snapshot
			// after removing the installation, without the original repository.
			require.NoError(t, os.RemoveAll(directory))
			fresh := &Cache{root: cache.root}
			require.NoError(t, fresh.Install(t.Context(), lock))
			require.NoError(t, fresh.VerifyCached(t.Context(), lock))
		})
	}
}
