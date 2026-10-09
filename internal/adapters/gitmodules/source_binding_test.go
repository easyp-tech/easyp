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

func TestRepairSharedRepositoryRetainsEveryPinnedCommit(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{
		"one/base.rules": "version: v1\nlinters:\n  default: MINIMAL\n",
		"two/base.rules": "version: v1\nlinters:\n  default: MINIMAL\n",
	}, nil)
	first := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	runTestGit(t, repository, "checkout", "--orphan", "unrelated", "-q")
	require.NoError(t, os.WriteFile(filepath.Join(repository, "two/base.rules"), []byte("version: v1\nlinters:\n  default: STANDARD\n"), 0o644))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "second module")
	second := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	cache := &Cache{root: t.TempDir()}
	lock := v1.Lock{Modules: []v1.LockedModule{
		{Source: repository + "/one", Commit: first},
		{Source: repository + "/two", Commit: second},
	}}
	require.NoError(t, cache.RepairPolicySources(t.Context(), lock))
	require.NoError(t, os.RemoveAll(filepath.Join(cache.root, "objects")))
	require.NoError(t, cache.RepairPolicySources(t.Context(), lock))
	file, err := cache.PolicyFiles(lock.Modules[0], "one").Read(t.Context(), "base.rules")
	require.NoError(t, err)
	require.Contains(t, string(file.Content), "MINIMAL")
	file, err = cache.PolicyFiles(lock.Modules[1], "two").Read(t.Context(), "base.rules")
	require.NoError(t, err)
	require.Contains(t, string(file.Content), "STANDARD")
}

func TestRepairFetchHonorsObjectRepositoryLock(t *testing.T) {
	t.Parallel()
	repository := filepath.Join(t.TempDir(), "objects")
	unlock, err := lockObjectRepository(t.Context(), repository+".lock")
	require.NoError(t, err)
	defer unlock()
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	err = refetchPinnedSource(ctx, repository, "unavailable.test/repository", strings.Repeat("a", 40))
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
