package sourceview

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionSharesRootsAndKeepsChecks(t *testing.T) {
	t.Parallel()

	root, outside := t.TempDir(), t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "a", "b"), 0o755))
	for _, name := range []string{"one.proto", "two.proto"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, "a", "b", name), []byte(name), 0o600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.proto"), []byte("secret"), 0o600))
	require.NoError(t, os.Symlink(filepath.Join(outside, "secret.proto"), filepath.Join(root, "a", "b", "escape.proto")))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested", ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nested", "inner.proto"), []byte("inner"), 0o600))
	require.NoError(t, os.Symlink("../nested/inner.proto", filepath.Join(root, "a", "nested_alias.proto")))
	require.NoError(t, os.WriteFile(filepath.Join(root, "orig.proto"), []byte("orig"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "repl.proto"), []byte("replaced"), 0o600))

	session := NewSession()
	t.Cleanup(func() {
		require.NoError(t, session.Close())
	})

	for _, name := range []string{"one.proto", "two.proto"} {
		file, err := session.OpenLocal(t.Context(), root, "a/b/"+name)
		require.NoError(t, err)
		data, readErr := io.ReadAll(file)
		require.NoError(t, readErr)
		require.NoError(t, file.Close())
		assert.Equal(t, name, string(data))
	}
	assert.Len(t, session.roots, 1)

	data, err := session.ReadLocal(t.Context(), root, "a/b/one.proto")
	require.NoError(t, err)
	assert.Equal(t, "one.proto", string(data))

	_, err = session.OpenLocal(t.Context(), root, "a/b/escape.proto")
	require.ErrorIs(t, err, ErrOutsideRoot)

	_, err = session.OpenLocal(t.Context(), root, "a/nested_alias.proto")
	require.ErrorIs(t, err, ErrNestedRepository)

	_, err = session.OpenLocal(t.Context(), root, "a/b/missing.proto")
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = session.OpenLocalSelected(t.Context(), root, "orig.proto", func(Resolution) bool {
		require.NoError(t, os.Rename(filepath.Join(root, "repl.proto"), filepath.Join(root, "orig.proto")))
		return true
	})
	require.ErrorIs(t, err, ErrChanged)
}

func TestSessionBoundsCachedDirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const dirCount = 320
	for i := range dirCount {
		dir := filepath.Join(root, fmt.Sprintf("pkg%03d", i), "v1")
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "service.proto"), []byte(fmt.Sprintf("pkg%03d", i)), 0o600))
	}

	const maxDirs = 16
	session := newSessionWithMaxDirs(maxDirs)
	t.Cleanup(func() {
		require.NoError(t, session.Close())
	})

	for i := range dirCount {
		logical := fmt.Sprintf("pkg%03d/v1/service.proto", i)
		data, err := session.ReadLocal(t.Context(), root, logical)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("pkg%03d", i), string(data))
		assert.LessOrEqual(t, session.cache.cachedCount(), maxDirs)
	}
}

func TestSessionConcurrentOpensAndEvictions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const dirCount = 64
	for i := range dirCount {
		dir := filepath.Join(root, fmt.Sprintf("d%02d", i), "sub")
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "item.proto"), []byte(fmt.Sprintf("d%02d", i)), 0o600))
	}

	const maxDirs = 8
	session := newSessionWithMaxDirs(maxDirs)
	t.Cleanup(func() {
		require.NoError(t, session.Close())
	})

	var wg sync.WaitGroup
	errs := make(chan error, dirCount*2)
	for worker := range 8 {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for step := range dirCount {
				idx := (offset*7 + step) % dirCount
				logical := fmt.Sprintf("d%02d/sub/item.proto", idx)
				data, err := session.ReadLocal(t.Context(), root, logical)
				if err != nil {
					errs <- fmt.Errorf("read %s: %w", logical, err)
					return
				}
				want := fmt.Sprintf("d%02d", idx)
				if string(data) != want {
					errs <- fmt.Errorf("read %s = %q, want %q", logical, data, want)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	assert.LessOrEqual(t, session.cache.cachedCount(), maxDirs)
}

func TestSessionCloseAndNil(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.proto"), []byte("hello"), 0o600))

	var nilSession *Session
	data, err := nilSession.ReadLocal(t.Context(), root, "a.proto")
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
	require.NoError(t, nilSession.Close())

	session := NewSession()
	data, err = session.ReadLocal(t.Context(), root, "a.proto")
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))

	require.NoError(t, session.Close())
	require.NoError(t, session.Close())

	_, err = session.OpenLocal(t.Context(), root, "a.proto")
	require.ErrorIs(t, err, fs.ErrClosed)
}
