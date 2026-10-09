package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestPinnedPolicySourceBoundsAliasedModule(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{
		"v2/base.rules": "version: v1\nlinters:\n  default: MINIMAL\n",
		"outside.rules": "version: v1\nlinters:\n  default: STANDARD\n",
	}, map[string]string{"module": "v2", "v2/escape.rules": "../outside.rules"})
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	files := cache.PolicyFiles(fetched.Lock, "module")
	_, err = files.Read(t.Context(), "base.rules")
	require.NoError(t, err)
	_, err = files.Read(t.Context(), "escape.rules")
	require.ErrorContains(t, err, "outside root")
}

func TestPinnedPolicySourceSHA256(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	raw := "version: v1\nlinters:\n  default: MINIMAL\n"
	require.NoError(t, os.WriteFile(filepath.Join(repository, "protobuf.mod"), []byte("module "+repository+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repository, "api.proto"), []byte("syntax = \"proto3\";\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repository, "base.rules"), []byte(raw), 0o644))
	runTestGit(t, repository, "init", "--object-format=sha256", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, commit)
	require.NoError(t, err)
	require.Len(t, fetched.Lock.Commit, 64)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	file, err := cache.PolicyFiles(fetched.Lock, ".").Read(t.Context(), "base.rules")
	require.NoError(t, err)
	require.Equal(t, raw, string(file.Content))
	upper := fetched.Lock
	upper.Commit = strings.ToUpper(upper.Commit)
	upper.Version = strings.ToUpper(upper.Version)
	file, err = cache.PolicyFiles(upper, ".").Read(t.Context(), "base.rules")
	require.NoError(t, err)
	require.Equal(t, raw, string(file.Content))
}
