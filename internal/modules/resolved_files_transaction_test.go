package modules

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestResolvedTransactionRestoresEveryAppliedDestination(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		at   int
	}{
		{name: "after_source_commit", at: 2},
		{name: "after_source_and_lock_commit", at: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, tx := resolvedTransactionFixture(t)
			cause := errors.New("rename failed")
			calls := 0
			tx.rename = func(root *os.Root, before, after string) error {
				calls++
				if calls == tt.at {
					return cause
				}
				return root.Rename(before, after)
			}

			err := tx.apply()

			require.ErrorIs(t, err, cause)
			assert.Equal(t, tt.at*2-1, calls, "failure must occur after committed destinations and restore them")
			assertResolvedFixture(t, root)
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			assert.Len(t, entries, 3)
		})
	}
}

func TestResolvedTransactionStagingFailureKeepsDestinations(t *testing.T) {
	t.Parallel()
	root, tx := resolvedTransactionFixture(t)
	cause := errors.New("stage write failed")
	tx.stage = func(*os.Root, string, []byte, os.FileMode) error { return cause }

	err := tx.apply()

	require.ErrorIs(t, err, cause)
	assertResolvedFixture(t, root)
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestResolvedTransactionRejectsChangedValidationInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(*testing.T, string)
	}{
		{name: "source_bytes", change: func(t *testing.T, root string) {
			require.NoError(t, os.WriteFile(filepath.Join(root, "consumer.proto"), []byte("concurrent source"), 0o640))
		}},
		{name: "unchanged_source_bytes", change: func(t *testing.T, root string) {
			require.NoError(t, os.WriteFile(filepath.Join(root, "untouched.proto"), []byte("concurrent unchanged source"), 0o640))
		}},
		{name: "source_inode", change: func(t *testing.T, root string) {
			require.NoError(t, os.Rename(filepath.Join(root, "consumer.proto"), filepath.Join(root, "previous.proto")))
			require.NoError(t, os.WriteFile(filepath.Join(root, "consumer.proto"), []byte("original source"), 0o640))
		}},
		{name: "metadata_bytes", change: func(t *testing.T, root string) {
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), []byte("concurrent manifest"), 0o600))
		}},
		{name: "replaced_source_alias", change: func(t *testing.T, root string) {
			require.NoError(t, os.Remove(filepath.Join(root, "alias.proto")))
			require.NoError(t, os.Symlink("consumer.proto", filepath.Join(root, "alias.proto")))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, tx := resolvedTransactionFixture(t)
			require.NoError(t, os.WriteFile(filepath.Join(root, "untouched.proto"), []byte("unchanged source"), 0o640))
			require.NoError(t, os.Symlink("consumer.proto", filepath.Join(root, "alias.proto")))
			_, err := tx.capture("untouched.proto", nil)
			require.NoError(t, err)
			_, err = tx.capture("alias.proto", nil)
			require.NoError(t, err)
			tx.beforeCommit = func() error {
				tt.change(t, root)
				return nil
			}

			err = tx.apply()

			require.ErrorIs(t, err, sourceview.ErrChanged)
			current, err := os.ReadFile(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, "original lock", string(current))
		})
	}
}

func TestResolvedTransactionRetainsRecoveryOnRollbackFailure(t *testing.T) {
	t.Parallel()
	root, tx := resolvedTransactionFixture(t)
	cause := errors.New("filesystem refused rename")
	calls := 0
	tx.rename = func(root *os.Root, before, after string) error {
		calls++
		if calls > 1 {
			return cause
		}
		return root.Rename(before, after)
	}

	err := tx.apply()

	require.ErrorIs(t, err, cause)
	require.ErrorContains(t, err, "recovery copies remain")
	directories, err := filepath.Glob(filepath.Join(root, ".easyp-resolved-*"))
	require.NoError(t, err)
	require.Len(t, directories, 1)
	recovery, err := os.ReadFile(filepath.Join(directories[0], "original-0"))
	require.NoError(t, err)
	assert.Equal(t, "original source", string(recovery))
}

func resolvedTransactionFixture(t *testing.T) (string, *resolvedFilesTransaction) {
	t.Helper()
	root := t.TempDir()
	for name, fixture := range map[string]struct {
		data string
		mode os.FileMode
	}{
		"consumer.proto": {data: "original source", mode: 0o640},
		v1.ModuleFile:    {data: "original manifest", mode: 0o600},
		v1.LockFile:      {data: "original lock", mode: 0o644},
	} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(fixture.data), fixture.mode))
	}
	tx, err := newResolvedFilesTransaction(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, tx.close()) })
	for _, name := range []string{"consumer.proto", v1.ModuleFile, v1.LockFile} {
		_, err := tx.capture(name, nil)
		require.NoError(t, err)
		err = tx.plan(name, []byte("new "+name), 0o644)
		require.NoError(t, err)
	}
	return root, tx
}

func assertResolvedFixture(t *testing.T, root string) {
	t.Helper()
	for name, fixture := range map[string]struct {
		data string
		mode os.FileMode
	}{
		"consumer.proto": {data: "original source", mode: 0o640},
		v1.ModuleFile:    {data: "original manifest", mode: 0o600},
		v1.LockFile:      {data: "original lock", mode: 0o644},
	} {
		data, err := os.ReadFile(filepath.Join(root, name))
		require.NoError(t, err)
		assert.Equal(t, fixture.data, string(data), name)
		info, err := os.Stat(filepath.Join(root, name))
		require.NoError(t, err)
		assert.Equal(t, fixture.mode, info.Mode().Perm(), name)
	}
}

func TestResolvedWriterPreservesMetadataAliasesAndModes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	original := []byte("module example.test/consumer\n")
	require.NoError(t, os.Mkdir(filepath.Join(root, "metadata"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "metadata/mod"), original, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "metadata/lock"), []byte("version: 1\nmodules: []\n"), 0o640))
	require.NoError(t, os.Symlink("metadata/mod", filepath.Join(root, v1.ModuleFile)))
	require.NoError(t, os.Symlink("metadata/lock", filepath.Join(root, v1.LockFile)))
	updated := []byte("module example.test/consumer // keep\n")

	err := writeV1ResolvedFiles(root, original, updated, v1.Lock{Version: 1})

	require.NoError(t, err)
	for _, name := range []string{v1.ModuleFile, v1.LockFile} {
		info, err := os.Lstat(filepath.Join(root, name))
		require.NoError(t, err)
		assert.NotZero(t, info.Mode()&os.ModeSymlink, name)
	}
	mod, err := os.ReadFile(filepath.Join(root, "metadata/mod"))
	require.NoError(t, err)
	assert.Equal(t, updated, mod)
	modInfo, err := os.Stat(filepath.Join(root, "metadata/mod"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), modInfo.Mode().Perm())
	lockInfo, err := os.Stat(filepath.Join(root, "metadata/lock"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), lockInfo.Mode().Perm())
}

func TestResolvedWriterRejectsChangedOriginalManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	original := []byte("module example.test/original\n")
	concurrent := []byte("module example.test/concurrent\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), concurrent, 0o600))

	err := writeV1ResolvedFiles(root, original, original, v1.Lock{Version: 1})

	require.ErrorIs(t, err, sourceview.ErrChanged)
	assert.NoFileExists(t, filepath.Join(root, v1.LockFile))
	current, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
	require.NoError(t, err)
	assert.Equal(t, concurrent, current)
}
