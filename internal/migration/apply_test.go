package migration

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_transaction_apply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode os.FileMode
	}{
		{name: "private", mode: 0o600},
		{name: "group_readable", mode: 0o640},
		{name: "read_only", mode: 0o444},
		{name: "executable", mode: 0o751},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeTransactionFile(t, root, "easyp.yaml", "original", tt.mode)
			writeTransactionFile(t, root, "untouched.bak", "older backup", 0o600)
			tx, err := newTransaction(root)
			require.NoError(t, err)
			for _, name := range []string{"easyp.yaml", "easyp.yaml.bak", "easyp.gen.yaml", "untouched.bak"} {
				_, err = tx.capture(name)
				require.NoError(t, err)
			}
			tx.changes = []fileChange{
				{name: "easyp.yaml.bak", content: []byte("original"), mode: tt.mode},
				{name: "easyp.gen.yaml", content: []byte("generated"), mode: tt.mode},
				{name: "easyp.yaml", content: []byte("migrated"), mode: tt.mode},
			}

			require.NoError(t, tx.apply())
			assertTransactionFile(t, root, "easyp.yaml", "migrated", tt.mode)
			assertTransactionFile(t, root, "easyp.yaml.bak", "original", tt.mode)
			assertTransactionFile(t, root, "easyp.gen.yaml", "generated", tt.mode)
			assertTransactionFile(t, root, "untouched.bak", "older backup", 0o600)
			assertNoTransactionTemporary(t, root)
		})
	}
}

func Test_transaction_rejectsChangesBeforeWriting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "source_content",
			mutate: func(t *testing.T, root string) {
				writeTransactionFile(t, root, "input.proto", "changed proto", 0o644)
			},
		},
		{
			name: "source_mode",
			mutate: func(t *testing.T, root string) {
				require.NoError(t, os.Chmod(filepath.Join(root, "input.proto"), 0o600))
			},
		},
		{
			name: "source_inode",
			mutate: func(t *testing.T, root string) {
				writeTransactionFile(t, root, "replacement", "proto", 0o644)
				require.NoError(t, os.Rename(filepath.Join(root, "replacement"), filepath.Join(root, "input.proto")))
			},
		},
		{
			name: "destination_appeared",
			mutate: func(t *testing.T, root string) {
				writeTransactionFile(t, root, "easyp.gen.yaml", "someone else's output", 0o644)
			},
		},
		{
			name: "destination_symlink",
			mutate: func(t *testing.T, root string) {
				require.NoError(t, os.Symlink("input.proto", filepath.Join(root, "easyp.gen.yaml")))
			},
		},
		{
			name: "replacement_content",
			mutate: func(t *testing.T, root string) {
				writeTransactionFile(t, root, "easyp.yaml", "edited original", 0o640)
			},
		},
		{
			name: "replacement_mode",
			mutate: func(t *testing.T, root string) {
				require.NoError(t, os.Chmod(filepath.Join(root, "easyp.yaml"), 0o600))
			},
		},
		{
			name: "backup_appeared",
			mutate: func(t *testing.T, root string) {
				writeTransactionFile(t, root, "easyp.yaml.bak", "unrelated backup", 0o600)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeTransactionFile(t, root, "easyp.yaml", "original", 0o640)
			writeTransactionFile(t, root, "input.proto", "proto", 0o644)
			tx, err := newTransaction(root)
			require.NoError(t, err)
			for _, name := range []string{"input.proto", "easyp.yaml", "easyp.yaml.bak", "easyp.gen.yaml"} {
				_, err = tx.capture(name)
				require.NoError(t, err)
			}
			tx.changes = []fileChange{
				{name: "easyp.yaml.bak", content: []byte("original"), mode: 0o640},
				{name: "easyp.gen.yaml", content: []byte("generated"), mode: 0o644},
				{name: "easyp.yaml", content: []byte("migrated"), mode: 0o640},
			}
			tt.mutate(t, root)
			before, err := os.ReadFile(filepath.Join(root, "easyp.yaml"))
			require.NoError(t, err)

			require.Error(t, tx.apply())
			after, err := os.ReadFile(filepath.Join(root, "easyp.yaml"))
			require.NoError(t, err)
			assert.Equal(t, before, after)
			if tt.name == "backup_appeared" {
				assertTransactionFile(t, root, "easyp.yaml.bak", "unrelated backup", 0o600)
			} else {
				_, err = os.Lstat(filepath.Join(root, "easyp.yaml.bak"))
				require.ErrorIs(t, err, os.ErrNotExist)
			}
			assertNoTransactionTemporary(t, root)
		})
	}
}

func Test_transaction_rollsBackFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*transaction, error)
	}{
		{
			name: "staging",
			setup: func(tx *transaction, failure error) {
				count := 0
				tx.stage = func(root *os.Root, name string, data []byte, mode os.FileMode) error {
					count++
					if count == 2 {
						return failure
					}
					return stageFile(root, name, data, mode)
				}
			},
		},
		{
			name: "second_replacement",
			setup: func(tx *transaction, failure error) {
				tx.rename = func(root *os.Root, oldName, newName string) error {
					if newName == "second.yaml" {
						return failure
					}
					return root.Rename(oldName, newName)
				}
			},
		},
		{
			name: "new_destination",
			setup: func(tx *transaction, failure error) {
				tx.link = func(root *os.Root, oldName, newName string) error {
					if newName == "generated.yaml" {
						return failure
					}
					return root.Link(oldName, newName)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeTransactionFile(t, root, "first.yaml", "first", 0o640)
			writeTransactionFile(t, root, "second.yaml", "second", 0o600)
			writeTransactionFile(t, root, "existing.bak", "preserved", 0o444)
			tx, err := newTransaction(root)
			require.NoError(t, err)
			for _, name := range []string{"first.yaml", "second.yaml", "first.yaml.bak", "generated.yaml", "existing.bak"} {
				_, err = tx.capture(name)
				require.NoError(t, err)
			}
			tx.changes = []fileChange{
				{name: "first.yaml.bak", content: []byte("first"), mode: 0o640},
				{name: "first.yaml", content: []byte("new first"), mode: 0o640},
				{name: "second.yaml", content: []byte("new second"), mode: 0o600},
				{name: "generated.yaml", content: []byte("generated"), mode: 0o644},
			}
			failure := errors.New("injected filesystem failure")
			tt.setup(tx, failure)

			require.ErrorIs(t, tx.apply(), failure)
			assertTransactionFile(t, root, "first.yaml", "first", 0o640)
			assertTransactionFile(t, root, "second.yaml", "second", 0o600)
			assertTransactionFile(t, root, "existing.bak", "preserved", 0o444)
			for _, name := range []string{"first.yaml.bak", "generated.yaml"} {
				_, err = os.Lstat(filepath.Join(root, name))
				require.ErrorIs(t, err, os.ErrNotExist)
			}
			assertNoTransactionTemporary(t, root)
		})
	}
}

func Test_transaction_captureRejectsUnsafeFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		setup    func(*testing.T, string)
	}{
		{name: "escape", filename: "../outside"},
		{name: "absolute", filename: "/outside"},
		{name: "directory", filename: "."},
		{
			name: "symlink", filename: "source.yaml",
			setup: func(t *testing.T, root string) {
				require.NoError(t, os.Symlink("missing", filepath.Join(root, "source.yaml")))
			},
		},
		{
			name: "directory_file", filename: "source.yaml",
			setup: func(t *testing.T, root string) {
				require.NoError(t, os.Mkdir(filepath.Join(root, "source.yaml"), 0o755))
			},
		},
		{
			name: "symlink_parent", filename: "proto/source.proto",
			setup: func(t *testing.T, root string) {
				require.NoError(t, os.Mkdir(filepath.Join(root, "real"), 0o755))
				writeTransactionFile(t, root, "real/source.proto", "proto", 0o644)
				require.NoError(t, os.Symlink("real", filepath.Join(root, "proto")))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, root)
			}
			tx, err := newTransaction(root)
			require.NoError(t, err)
			_, err = tx.capture(tt.filename)
			require.Error(t, err)
		})
	}
}

func Test_transaction_rejectsReplacedRoot(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	require.NoError(t, os.Mkdir(root, 0o755))
	writeTransactionFile(t, root, "easyp.yaml", "original", 0o600)
	tx, err := newTransaction(root)
	require.NoError(t, err)
	_, err = tx.capture("easyp.yaml")
	require.NoError(t, err)
	tx.changes = []fileChange{{name: "easyp.yaml", content: []byte("migrated"), mode: 0o600}}
	require.NoError(t, os.Rename(root, filepath.Join(parent, "old")))
	require.NoError(t, os.Mkdir(root, 0o755))
	writeTransactionFile(t, root, "easyp.yaml", "other project", 0o600)

	require.Error(t, tx.apply())
	assertTransactionFile(t, root, "easyp.yaml", "other project", 0o600)
	assertTransactionFile(t, filepath.Join(parent, "old"), "easyp.yaml", "original", 0o600)
	assertNoTransactionTemporary(t, root)
}

func Test_newTransaction_rejectsSymlinkRoot(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	require.NoError(t, os.Mkdir(root, 0o755))
	link := filepath.Join(parent, "link")
	require.NoError(t, os.Symlink(root, link))
	_, err := newTransaction(link)
	require.Error(t, err)
}

func Test_transaction_rechecksAfterStaging(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "input.proto", "original proto", 0o644)
	tx, err := newTransaction(root)
	require.NoError(t, err)
	for _, name := range []string{"input.proto", "easyp.yaml"} {
		_, err = tx.capture(name)
		require.NoError(t, err)
	}
	tx.changes = []fileChange{{name: "easyp.yaml", content: []byte("migrated"), mode: 0o644}}
	tx.stage = func(handle *os.Root, name string, data []byte, mode os.FileMode) error {
		writeTransactionFile(t, root, "input.proto", "edited proto", 0o644)
		return stageFile(handle, name, data, mode)
	}

	require.ErrorContains(t, tx.apply(), "changed since planning")
	_, err = os.Lstat(filepath.Join(root, "easyp.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
	assertNoTransactionTemporary(t, root)
}

func Test_transaction_beforeApplyRejectsAfterStaging(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "easyp.yaml", "original", 0o640)
	tx, err := newTransaction(root)
	require.NoError(t, err)
	for _, name := range []string{"easyp.yaml", "easyp.gen.yaml"} {
		_, err = tx.capture(name)
		require.NoError(t, err)
	}
	tx.changes = []fileChange{
		{name: "easyp.yaml", content: []byte("migrated"), mode: 0o640},
		{name: "easyp.gen.yaml", content: []byte("generated"), mode: 0o644},
	}
	failure := errors.New("source inventory changed")
	called := false
	tx.beforeApply = func() error {
		called = true
		entries, readErr := os.ReadDir(root)
		require.NoError(t, readErr)
		var stageDirectory string
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".easyp-migrate-") {
				stageDirectory = entry.Name()
			}
		}
		require.NotEmpty(t, stageDirectory)
		assertTransactionFile(t, root, filepath.Join(stageDirectory, "output-0"), "migrated", 0o640)
		assertTransactionFile(t, root, filepath.Join(stageDirectory, "original-0"), "original", 0o640)
		assertTransactionFile(t, root, filepath.Join(stageDirectory, "output-1"), "generated", 0o644)
		return failure
	}

	require.ErrorIs(t, tx.apply(), failure)
	assert.True(t, called)
	assertTransactionFile(t, root, "easyp.yaml", "original", 0o640)
	_, err = os.Lstat(filepath.Join(root, "easyp.gen.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
	assertNoTransactionTemporary(t, root)
}

func Test_transaction_neverOverwritesLateDestination(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "easyp.yaml", "original", 0o640)
	tx, err := newTransaction(root)
	require.NoError(t, err)
	for _, name := range []string{"easyp.yaml", "easyp.gen.yaml"} {
		_, err = tx.capture(name)
		require.NoError(t, err)
	}
	tx.changes = []fileChange{
		{name: "easyp.yaml", content: []byte("migrated"), mode: 0o640},
		{name: "easyp.gen.yaml", content: []byte("generated"), mode: 0o644},
	}
	tx.link = func(handle *os.Root, oldName, newName string) error {
		writeTransactionFile(t, root, newName, "concurrent output", 0o600)
		return handle.Link(oldName, newName)
	}

	require.ErrorIs(t, tx.apply(), os.ErrExist)
	assertTransactionFile(t, root, "easyp.yaml", "original", 0o640)
	assertTransactionFile(t, root, "easyp.gen.yaml", "concurrent output", 0o600)
	assertNoTransactionTemporary(t, root)
}

func Test_transaction_preservesRecoveryCopiesWhenRollbackFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "easyp.yaml", "original", 0o640)
	tx, err := newTransaction(root)
	require.NoError(t, err)
	for _, name := range []string{"easyp.yaml", "easyp.gen.yaml"} {
		_, err = tx.capture(name)
		require.NoError(t, err)
	}
	tx.changes = []fileChange{
		{name: "easyp.yaml", content: []byte("migrated"), mode: 0o640},
		{name: "easyp.gen.yaml", content: []byte("generated"), mode: 0o644},
	}
	applyFailure := errors.New("install failure")
	rollbackFailure := errors.New("rollback failure")
	tx.link = func(_ *os.Root, _, _ string) error { return applyFailure }
	tx.rename = func(handle *os.Root, oldName, newName string) error {
		if strings.Contains(oldName, "original-") {
			return rollbackFailure
		}
		return handle.Rename(oldName, newName)
	}

	err = tx.apply()
	require.ErrorIs(t, err, applyFailure)
	require.ErrorIs(t, err, rollbackFailure)
	require.ErrorContains(t, err, "recovery copies remain")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	var recoveryDirectory string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".easyp-migrate-") {
			recoveryDirectory = entry.Name()
		}
	}
	require.NotEmpty(t, recoveryDirectory)
	assertTransactionFile(t, root, filepath.Join(recoveryDirectory, "original-0"), "original", 0o640)
}

func Test_transaction_validatesPlannedDestinations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		changes []fileChange
	}{
		{name: "uncaptured", changes: []fileChange{{name: "other.yaml", mode: 0o644}}},
		{name: "duplicate", changes: []fileChange{{name: "easyp.yaml", mode: 0o644}, {name: "easyp.yaml", mode: 0o644}}},
		{name: "escape", changes: []fileChange{{name: "../easyp.yaml", mode: 0o644}}},
		{name: "nonregular_mode", changes: []fileChange{{name: "easyp.yaml", mode: os.ModeSymlink | 0o644}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			tx, err := newTransaction(root)
			require.NoError(t, err)
			_, err = tx.capture("easyp.yaml")
			require.NoError(t, err)
			tx.changes = tt.changes

			require.Error(t, tx.apply())
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

func Test_transaction_canonicalizesParentAlias(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	realParent := filepath.Join(parent, "real")
	require.NoError(t, os.Mkdir(realParent, 0o755))
	root := filepath.Join(realParent, "root")
	require.NoError(t, os.Mkdir(root, 0o755))
	alias := filepath.Join(parent, "alias")
	require.NoError(t, os.Symlink(realParent, alias))
	tx, err := newTransaction(filepath.Join(alias, "root"))
	require.NoError(t, err)
	_, err = tx.capture("easyp.yaml")
	require.NoError(t, err)
	tx.changes = []fileChange{{name: "easyp.yaml", content: []byte("migrated"), mode: 0o644}}

	require.NoError(t, tx.apply())
	assertTransactionFile(t, root, "easyp.yaml", "migrated", 0o644)
	assertNoTransactionTemporary(t, root)
}

func Test_transaction_capturesNestedSource(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "proto"), 0o755))
	writeTransactionFile(t, root, "proto/source.proto", "proto", 0o644)
	tx, err := newTransaction(root)
	require.NoError(t, err)
	initial, err := tx.capture("proto/source.proto")
	require.NoError(t, err)
	assert.Equal(t, "proto", string(initial.data))
	require.NoError(t, tx.apply())
	writeTransactionFile(t, root, "proto/source.proto", "edited proto", 0o644)
	_, err = tx.capture("proto/source.proto")
	require.ErrorContains(t, err, "changed since planning")
	assertNoTransactionTemporary(t, root)
}

func writeTransactionFile(t *testing.T, root, name, content string, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.WriteFile(path, []byte(content), mode))
	require.NoError(t, os.Chmod(path, mode))
}

func assertTransactionFile(t *testing.T, root, name, content string, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(root, name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, mode, info.Mode())
}

func assertNoTransactionTemporary(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), ".easyp-migrate-"), "leftover: %s", entry.Name())
	}
}
