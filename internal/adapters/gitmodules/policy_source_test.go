package gitmodules

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestPinnedPolicySourceReadsArbitraryConfigNames(t *testing.T) {
	t.Parallel()
	raw := "version: v1\nlinters:\n  default: MINIMAL\n"
	repository := snapshotPolicyFixture(t, map[string]string{"policies/base.rules": raw}, map[string]string{"alias": "policies"})
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	files := cache.PolicyFiles(fetched.Lock, ".")
	got, err := files.Read(t.Context(), "alias/base.rules")
	require.NoError(t, err)
	assert.Equal(t, raw, string(got.Content))
	assert.Equal(t, "alias/base.rules", got.Path)
	assert.Contains(t, got.Canonical, fetched.Lock.Commit)
	_, err = files.Read(t.Context(), "../outside.rules")
	require.Error(t, err)
}

func TestPinnedPolicySourceMissingObjectsIsReadOnlyAndDownloadRepairs(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"base.rules": "version: v1\nlinters:\n  default: MINIMAL\n"}, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	require.NoError(t, os.RemoveAll(filepath.Join(cache.root, "objects")))
	_, err = cache.PolicyFiles(fetched.Lock, ".").Read(t.Context(), "base.rules")
	require.ErrorContains(t, err, "easyp mod download")
	_, statErr := os.Stat(filepath.Join(cache.root, "objects"))
	require.ErrorIs(t, statErr, os.ErrNotExist)
	require.NoError(t, cache.Install(t.Context(), lock))
	require.NoError(t, cache.RepairPolicySources(t.Context(), lock))
	_, err = cache.PolicyFiles(fetched.Lock, ".").Read(t.Context(), "base.rules")
	require.NoError(t, err)
}

func TestPinnedPolicySourceRejectsTamperedObject(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"base.rules": "version: v1\nlinters:\n  default: MINIMAL\n"}, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	hash := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD:base.rules"))
	objectRepository := filepath.Join(cache.root, "objects", v1CacheSourceKey(repository))
	policyTestLooseObjects(t, objectRepository)
	object := filepath.Join(objectRepository, "objects", hash[:2], hash[2:])
	var data bytes.Buffer
	writer := zlib.NewWriter(&data)
	body := "version: v1\nlinters:\n  default: STANDARD\n"
	_, err = fmt.Fprintf(writer, "blob %d\x00%s", len(body), body)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.NoError(t, os.MkdirAll(filepath.Dir(object), 0o755))
	require.NoError(t, os.WriteFile(object, data.Bytes(), 0o644))
	require.NoError(t, cache.VerifyCached(t.Context(), lock), "native projection remains unchanged")
	_, err = cache.PolicyFiles(fetched.Lock, ".").Read(t.Context(), "base.rules")
	require.ErrorContains(t, err, "object hash mismatch")
}

func policyTestLooseObjects(t *testing.T, objectRepository string) {
	t.Helper()
	// Force the reader to use loose objects, rather than a valid duplicate in
	// the pack. Then corrupt the exact object selected by the pinned tree.
	listing := runTestGit(t, objectRepository, "cat-file", "--batch-all-objects", "--batch-check=%(objectname) %(objecttype)")
	for _, line := range strings.Split(strings.TrimSpace(listing), "\n") {
		fields := strings.Fields(line)
		require.Len(t, fields, 2)
		data, err := exec.CommandContext(t.Context(), "git", "-C", objectRepository, "cat-file", fields[1], fields[0]).Output()
		require.NoError(t, err)
		var compressed bytes.Buffer
		writer := zlib.NewWriter(&compressed)
		_, err = fmt.Fprintf(writer, "%s %d\x00", fields[1], len(data))
		require.NoError(t, err)
		_, err = writer.Write(data)
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		filename := filepath.Join(objectRepository, "objects", fields[0][:2], fields[0][2:])
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
		require.NoError(t, os.WriteFile(filename, compressed.Bytes(), 0o644))
	}
	require.NoError(t, os.RemoveAll(filepath.Join(objectRepository, "objects", "pack")))
}

func TestExplicitDownloadRepairsMissingPolicyBlob(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"base.rules": "version: v1\nlinters:\n  default: MINIMAL\n"}, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	objectRepository := filepath.Join(cache.root, "objects", v1CacheSourceKey(repository))
	policyTestLooseObjects(t, objectRepository)
	id := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD:base.rules"))
	require.NoError(t, os.Remove(filepath.Join(objectRepository, "objects", id[:2], id[2:])))
	require.NoError(t, cache.VerifyCached(t.Context(), lock))
	_, err = cache.PolicyFiles(fetched.Lock, ".").Read(t.Context(), "base.rules")
	require.Error(t, err)
	require.NoError(t, cache.RepairPolicySources(t.Context(), lock))
	_, err = cache.PolicyFiles(fetched.Lock, ".").Read(t.Context(), "base.rules")
	require.NoError(t, err)
}

func TestNativeCacheRemainsUsableOfflineWithoutUnselectedPolicyObjects(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"base.rules": "version: v1\nlinters:\n  default: MINIMAL\n"}, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	require.NoError(t, os.RemoveAll(filepath.Join(cache.root, "objects")))
	require.NoError(t, os.Rename(repository, repository+"-offline"))
	require.NoError(t, cache.Install(t.Context(), lock))
	require.NoError(t, cache.VerifyCached(t.Context(), lock))
}
