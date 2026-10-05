package gitmodules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestPinnedObjectCacheWorksWithoutRemote(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	remote := filepath.Join(root, "remote")
	require.NoError(t, os.Mkdir(remote, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte("module "+remote+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "api.proto"), []byte("syntax = \"proto3\"; package api; message Item {}\n"), 0o644))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
	cache := New(filepath.Join(root, "cache"))
	first, err := cache.Fetch(t.Context(), remote, commit)
	require.NoError(t, err)
	require.NoError(t, os.Rename(remote, remote+"-offline"))
	second, err := cache.Fetch(t.Context(), remote, commit)
	require.NoError(t, err)
	require.Equal(t, first.Lock, second.Lock)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{first.Lock}}))
	dir, _, err := cache.Cached(first.Lock)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "api.proto"), []byte("tampered"), 0o644))
	require.ErrorContains(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{first.Lock}}), "hash mismatch")
}

func TestObjectRepositoryLockHonorsCancellation(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "lock")
	unlock, err := lockObjectRepository(t.Context(), path)
	require.NoError(t, err)
	defer unlock()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()
	_, err = lockObjectRepository(ctx, path)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
