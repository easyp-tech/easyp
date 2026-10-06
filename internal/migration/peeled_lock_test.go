package migration

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/modules"
)

func TestMigrationPeeledLockPreservesHistoricalHashAndBackup(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	const dependency = "example.com/acme/dependency"
	legacyLock := dependency + " v0.4.0^{} " + testHash + "\n"
	writeFixture(t, root, "easyp.yaml", "deps: ["+dependency+"]\n")
	writeFixture(t, root, "easyp.lock", legacyLock)
	preview, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	assert.True(t, preview.NeedsLockResolution())
	require.Error(t, preview.Apply())
	repo := &mockRepository{
		fetched:     map[string]modules.Fetched{dependency + "@v0.4.0": migrationFetched(dependency, "v0.4.0", testCommit)},
		wantOldHash: testHash,
	}
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
	require.NoError(t, err)
	require.NoError(t, plan.Apply())
	assert.Equal(t, []string{dependency + "@v0.4.0"}, repo.calls)
	assert.Equal(t, legacyLock, string(mustRead(t, root, "easyp.lock")))
	lock, err := validateNativeLock(mustRead(t, root, "protobuf.lock"))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 1)
	assert.Equal(t, "v0.4.0", lock.Modules[0].Version)
	assert.Equal(t, testCommit, lock.Modules[0].Commit)
	assert.Equal(t, testNewHash, lock.Modules[0].Hash)
}

func TestMigrationPeeledLockRejectsUnsupportedPins(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, version string }{
		{name: "branch", version: "main^{}"},
		{name: "short semver", version: "v1.2^{}"},
		{name: "commit expression", version: testCommit + "^{}"},
		{name: "double expression", version: "v1.2.3^{}^{}"},
		{name: "peeled pseudo-shaped tag", version: "v0.0.0-20260101123456-" + testCommit + "^{}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseLegacyLock([]byte("example.com/acme/dependency " + tt.version + " " + testHash + "\n"))
			require.ErrorContains(t, err, "unsupported ref")
		})
	}
}

func TestMigrationPeeledLockPreservesVerificationError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	const dependency = "example.com/acme/dependency"
	legacy := "deps: [" + dependency + "]\n"
	legacyLock := dependency + " v0.4.0^{} " + testHash + "\n"
	writeFixture(t, root, "easyp.yaml", legacy)
	writeFixture(t, root, "easyp.lock", legacyLock)
	mismatch := errors.New("legacy hash mismatch")
	repo := &mockRepository{err: mismatch}
	_, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
	require.ErrorIs(t, err, mismatch)
	assert.Equal(t, []string{dependency + "@v0.4.0"}, repo.calls)
	assert.Equal(t, legacyLock, string(mustRead(t, root, "easyp.lock")))
	assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
	for _, name := range []string{"protobuf.mod", "protobuf.lock", "easyp.gen.yaml", "easyp.yaml.v0.bak"} {
		assert.NoFileExists(t, filepath.Join(root, name))
	}
}
