package gitmodules

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestGitOperationTimeoutPreservesCauseWithLiveCaller(t *testing.T) {
	t.Setenv("EASYP_GIT_TIMEOUT", "50ms")
	started := time.Now()
	_, err := gitV1(t.Context(), "", "-c", "alias.easyp-wait=!exec sleep 1", "easyp-wait")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NoError(t, t.Context().Err())
	assert.Less(t, time.Since(started), 800*time.Millisecond)
}

func TestGitTimeoutConfigurationFailsBeforeCommand(t *testing.T) {
	for _, value := range []string{"invalid", "0s", "-1s"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("EASYP_GIT_TIMEOUT", value)
			_, err := gitV1(t.Context(), "", "--version")
			require.ErrorContains(t, err, "EASYP_GIT_TIMEOUT")
		})
	}
}

func TestInvalidGitTimeoutLeavesCacheUntouched(t *testing.T) {
	t.Setenv("EASYP_GIT_TIMEOUT", "invalid")
	storage := filepath.Join(t.TempDir(), "cache")
	_, err := New(storage).Fetch(t.Context(), "example.test/team/repo", strings.Repeat("a", 40))
	require.ErrorContains(t, err, "EASYP_GIT_TIMEOUT")
	_, err = os.Stat(storage)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestObjectRepositoryLockUsesOperationTimeout(t *testing.T) {
	t.Setenv("EASYP_GIT_TIMEOUT", "50ms")
	path := filepath.Join(t.TempDir(), "lock")
	unlock, err := lockObjectRepository(t.Context(), path)
	require.NoError(t, err)
	defer unlock()
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = lockObjectRepository(ctx, path)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NoError(t, ctx.Err())
	assert.Less(t, time.Since(started), 400*time.Millisecond)
}

func TestPinnedFetchTimeoutDoesNotFetchFullHistory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SSH fixture uses a POSIX shell")
	}
	t.Setenv("EASYP_GIT_TIMEOUT", "50ms")
	directory := t.TempDir()
	attempts := filepath.Join(directory, "attempts")
	ssh := filepath.Join(directory, "ssh")
	require.NoError(t, os.WriteFile(ssh, []byte("#!/bin/sh\nprintf 'attempt\\n' >> '"+attempts+"'\nexec sleep 1\n"), 0o755))
	t.Setenv("GIT_SSH_COMMAND", ssh)
	t.Setenv("GIT_SSH_VARIANT", "ssh")
	_, err := checkoutCachedCommit(t.Context(), filepath.Join(directory, "checkout"),
		v1.LockedModule{Commit: strings.Repeat("a", 40)},
		v1GitModuleCandidate{remote: "ssh://example.test/team/repo"})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NoError(t, t.Context().Err())
	data, err := os.ReadFile(attempts)
	require.NoError(t, err)
	assert.Equal(t, "attempt\n", string(data))
}

func TestGitTimeoutStopsModuleCandidateSearch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SSH fixture uses a POSIX shell")
	}
	tests := []struct {
		name string
		call func(context.Context, *Cache, string) error
	}{
		{name: "versions", call: func(ctx context.Context, cache *Cache, source string) error {
			_, err := cache.Versions(ctx, source)
			return err
		}},
		{name: "named tag", call: func(ctx context.Context, cache *Cache, source string) error {
			_, err := cache.ResolveTag(ctx, source, "release")
			return err
		}},
		{name: "semantic tag", call: func(ctx context.Context, cache *Cache, source string) error {
			_, err := cache.Fetch(ctx, source, "v1.0.0")
			return err
		}},
		{name: "initial checkout", call: func(ctx context.Context, cache *Cache, source string) error {
			_, err := cache.Fetch(ctx, source, "")
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EASYP_GIT_TIMEOUT", "100ms")
			directory := t.TempDir()
			attempts := filepath.Join(directory, "attempts")
			ssh := filepath.Join(directory, "ssh")
			require.NoError(t, os.WriteFile(ssh, []byte("#!/bin/sh\nprintf 'attempt\\n' >> '"+attempts+"'\nexec sleep 1\n"), 0o755))
			t.Setenv("GIT_SSH_COMMAND", ssh)
			t.Setenv("GIT_SSH_VARIANT", "ssh")
			err := tt.call(t.Context(), New(filepath.Join(directory, "cache")), "ssh://example.test/team/repo/nested")
			require.ErrorIs(t, err, context.DeadlineExceeded)
			assert.NoError(t, t.Context().Err())
			data, err := os.ReadFile(attempts)
			require.NoError(t, err)
			assert.Equal(t, "attempt\n", string(data))
		})
	}
}
