package gitmodules

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestPinnedFetchSkipsHistoryForUnavailableRepository(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SSH fixture uses a POSIX shell")
	}
	tests := []struct {
		name       string
		diagnostic string
		attempts   int
	}{
		{name: "missing_repository", diagnostic: "ERROR: Repository not found.", attempts: 1},
		{name: "http_missing_repository", diagnostic: "fatal: repository 'https://example.test/team/repo' not found", attempts: 1},
		{name: "authentication_failed", diagnostic: "fatal: Authentication failed for 'https://example.test/team/repo'", attempts: 1},
		{name: "noninteractive_credentials", diagnostic: "fatal: could not read Username for 'https://example.test': terminal prompts disabled", attempts: 1},
		{name: "ssh_credentials", diagnostic: "git@example.test: Permission denied (publickey).", attempts: 1},
		{name: "transient_failure_keeps_fallback", diagnostic: "fatal: connection reset by peer", attempts: 2},
		{name: "unadvertised_commit_keeps_fallback", diagnostic: "error: Server does not allow request for unadvertised object", attempts: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			attempts := filepath.Join(directory, "attempts")
			ssh := filepath.Join(directory, "ssh")
			script := "#!/bin/sh\nprintf 'attempt\\n' >> " + quoteTestShell(attempts) + "\nprintf '%s\\n' " + quoteTestShell(tt.diagnostic) + " >&2\nexit 128\n"
			require.NoError(t, os.WriteFile(ssh, []byte(script), 0o755))
			t.Setenv("GIT_SSH_COMMAND", ssh)
			t.Setenv("GIT_SSH_VARIANT", "ssh")
			_, err := checkoutCachedCommit(t.Context(), filepath.Join(directory, "checkout"),
				v1.LockedModule{Commit: strings.Repeat("a", 40)},
				v1GitModuleCandidate{remote: "ssh://example.test/team/repo"})
			require.Error(t, err)
			data, err := os.ReadFile(attempts)
			require.NoError(t, err)
			assert.Equal(t, tt.attempts, strings.Count(string(data), "attempt\n"))
		})
	}
}

func TestPinnedFetchHistoryFallbackFindsAdvertisedAncestor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SSH fixture uses a POSIX shell")
	}
	directory := t.TempDir()
	remote := filepath.Join(directory, "remote")
	require.NoError(t, os.Mkdir(remote, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "api.proto"), []byte("first"), 0o644))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "first")
	commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "api.proto"), []byte("second"), 0o644))
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "second")
	attempts := filepath.Join(directory, "attempts")
	ssh := filepath.Join(directory, "ssh")
	script := "#!/bin/sh\nif test ! -e " + quoteTestShell(attempts) + "; then\n printf 'attempt\\n' >> " + quoteTestShell(attempts) + "\n printf '%s\\n' 'error: Server does not allow request for unadvertised object' >&2\n exit 128\nfi\nprintf 'attempt\\n' >> " + quoteTestShell(attempts) + "\nexec git-upload-pack " + quoteTestShell(remote) + "\n"
	require.NoError(t, os.WriteFile(ssh, []byte(script), 0o755))
	t.Setenv("GIT_SSH_COMMAND", ssh)
	t.Setenv("GIT_SSH_VARIANT", "ssh")
	checkout := filepath.Join(directory, "checkout")
	found, err := checkoutCachedCommit(t.Context(), checkout, v1.LockedModule{Commit: commit}, v1GitModuleCandidate{remote: "ssh://example.test/team/repo"})
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, commit, strings.TrimSpace(runTestGit(t, checkout, "rev-parse", "HEAD")))
	data, err := os.ReadFile(attempts)
	require.NoError(t, err)
	assert.Equal(t, 2, strings.Count(string(data), "attempt\n"))
}

func TestPinnedMissingNestedRepositoryStillSearchesParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SSH fixture uses a POSIX shell")
	}
	directory := t.TempDir()
	remote := filepath.Join(directory, "remote")
	require.NoError(t, os.MkdirAll(filepath.Join(remote, "nested"), 0o755))
	const source = "ssh://example.test/team/repo/nested"
	require.NoError(t, os.WriteFile(filepath.Join(remote, "nested", "protobuf.mod"), []byte("module "+source+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "nested", "api.proto"), []byte("syntax = \"proto3\"; package api;\n"), 0o644))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
	attempts := filepath.Join(directory, "attempts")
	ssh := filepath.Join(directory, "ssh")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + quoteTestShell(attempts) + "\ncase \"$*\" in\n */repo/nested*) printf '%s\\n' 'ERROR: Repository not found.' >&2; exit 128;;\nesac\nexec git-upload-pack " + quoteTestShell(remote) + "\n"
	require.NoError(t, os.WriteFile(ssh, []byte(script), 0o755))
	t.Setenv("GIT_SSH_COMMAND", ssh)
	t.Setenv("GIT_SSH_VARIANT", "ssh")
	fetched, err := New(filepath.Join(directory, "cache")).Fetch(t.Context(), source, commit)
	require.NoError(t, err)
	assert.Equal(t, source, fetched.Module.Name)
	assert.Equal(t, commit, fetched.Lock.Commit)
	data, err := os.ReadFile(attempts)
	require.NoError(t, err)
	assert.Len(t, strings.Split(strings.TrimSpace(string(data)), "\n"), 2)
	assert.Equal(t, 1, strings.Count(string(data), "/repo/nested"))
}

func quoteTestShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
