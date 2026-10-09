package gitmodules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestVerifyCachedDoesNotWriteVerificationStamp(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, nil, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	directory, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	stamp := cacheVerificationStampPath(directory)
	if err := os.Remove(stamp); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	require.NoError(t, cache.VerifyCached(t.Context(), lock))
	_, err = os.Stat(stamp)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestVerificationStampCannotAcceptUnrelatedSnapshotFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "README.md"), []byte("unrelated"), 0o644))
	hash, err := hashSnapshotV1Files(directory)
	require.NoError(t, err)
	fingerprint, _, err := cacheTreeFingerprint(directory)
	require.NoError(t, err)
	require.NoError(t, writeCacheVerificationStamp(directory, cacheVerificationStamp{Version: cacheVerificationVersion, Hash: hash, Fingerprint: fingerprint}))
	err = verifyInstalledV1Module(directory, v1.LockedModule{Hash: hash}, false)
	require.ErrorContains(t, err, "unrelated snapshot file")
}
