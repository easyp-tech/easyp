package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestRepairIncompleteSharedStoreRestoresHealthyEarlierPin(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"one/base.rules": "version: v1\nlinters:\n  default: MINIMAL\n"}, nil)
	first := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	require.NoError(t, os.MkdirAll(filepath.Join(repository, "two"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repository, "two/base.rules"), []byte("version: v1\nlinters:\n  default: STANDARD\n"), 0o644))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "later module")
	second := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	cache := &Cache{root: t.TempDir()}
	lock := v1.Lock{Modules: []v1.LockedModule{{Source: repository + "/one", Commit: first}, {Source: repository + "/two", Commit: second}}}
	require.NoError(t, cache.RepairPolicySources(t.Context(), lock))
	objectRepository := filepath.Join(cache.root, "objects", v1CacheSourceKey(repository))
	policyTestLooseObjects(t, objectRepository)
	id := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD:two/base.rules"))
	require.NoError(t, os.Remove(filepath.Join(objectRepository, "objects", id[:2], id[2:])))
	require.NoError(t, cache.RepairPolicySources(t.Context(), lock))
	for _, tc := range []struct {
		name         string
		pin          v1.LockedModule
		prefix, want string
	}{{"first", lock.Modules[0], "one", "MINIMAL"}, {"second", lock.Modules[1], "two", "STANDARD"}} {
		file, err := cache.PolicyFiles(tc.pin, tc.prefix).Read(t.Context(), "base.rules")
		require.NoError(t, err)
		require.Contains(t, string(file.Content), tc.want)
	}
}
