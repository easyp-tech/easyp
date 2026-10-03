package gitmodules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func diagnosticRepository(t *testing.T) (string, string) {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "repository with spaces")
	require.NoError(t, os.MkdirAll(remote, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "easyp.yaml"), []byte("generate:\n  inputs:\n    - directory: .\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "api.proto"), []byte("syntax = \"proto3\"; package api; message Item {}\n"), 0o644))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	return remote, strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
}

func TestCacheIntegrityDiagnosticIdentifiesRecoveryPath(t *testing.T) {
	t.Parallel()
	remote, commit := diagnosticRepository(t)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), remote, commit)
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	installed, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	name := filepath.Join(installed, "api.proto")
	require.NoError(t, os.WriteFile(name, []byte("tampered"), 0o644))
	err = cache.Install(t.Context(), lock)
	require.ErrorContains(t, err, "hash mismatch")
	assert.Contains(t, err.Error(), installed)
	assert.Contains(t, err.Error(), "protobuf.lock unchanged")
	assert.Contains(t, err.Error(), "easyp mod download")
	contents, readErr := os.ReadFile(name)
	require.NoError(t, readErr)
	assert.Equal(t, "tampered", string(contents), "a diagnostic must not automatically repair evidence")
}

func TestPinnedFailureDiagnostics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		offline bool
		want    string
	}{
		{name: "reachable_missing_commit", want: "restore access to the pinned revision"},
		{name: "inaccessible_repository", offline: true, want: "could not fetch locked commit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote, commit := diagnosticRepository(t)
			if tt.offline {
				require.NoError(t, os.Rename(remote, remote+"-offline"))
			} else {
				commit = strings.Repeat("1", 40)
			}
			_, err := New(t.TempDir()).Fetch(t.Context(), remote, commit)
			require.ErrorContains(t, err, tt.want)
			assert.Contains(t, err.Error(), remote)
			assert.Contains(t, err.Error(), commit)
			if tt.offline {
				assert.NotContains(t, err.Error(), "locked commit "+commit+" is unavailable")
			}
		})
	}
}

func TestTagAccessErrorIsNotMissingVersion(t *testing.T) {
	t.Parallel()
	remote, _ := diagnosticRepository(t)
	_, err := New(t.TempDir()).Fetch(t.Context(), remote, "v1.8.0")
	require.ErrorContains(t, err, "was not found")
	require.NoError(t, os.Rename(remote, remote+"-offline"))
	_, err = New(t.TempDir()).Fetch(t.Context(), remote, "v1.0.0")
	require.ErrorContains(t, err, "could not query Git tags")
	assert.NotContains(t, err.Error(), "was not found")
}

func TestGitCancellationKeepsContextCause(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := gitV1(ctx, root, "-c", "alias.wait=!printf ready > started; sleep 0.25", "wait")
		result <- err
	}()
	require.Eventually(t, func() bool { _, err := os.Stat(filepath.Join(root, "started")); return err == nil }, 5*time.Second, time.Millisecond)
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled git did not exit")
	}
}
