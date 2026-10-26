package gitmodules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestSnapshotVerificationWaitHonorsCancellation(t *testing.T) {
	t.Parallel()
	cache := New(t.TempDir())
	entry := v1.LockedModule{
		Source: "example.test/module", Version: "v1.0.0",
		Commit: "1111111111111111111111111111111111111111",
		Hash:   "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}
	snapshot := cache.snapshotState(v1ModuleCachePath(cache.root, entry))
	require.NoError(t, snapshot.lock(t.Context()))
	defer snapshot.unlock()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	err := cache.Install(ctx, v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestSourceFilesDistinguishesMissingAndEmptyInventory(t *testing.T) {
	t.Parallel()
	cache := New(t.TempDir())
	entry := v1.LockedModule{
		Source: "example.test/module", Version: "v1.0.0",
		Commit: "1111111111111111111111111111111111111111",
	}
	files, err := cache.SourceFiles(entry)
	require.NoError(t, err)
	require.Nil(t, files)
	// Compute the actual empty-directory hash before selecting the cache path.
	entry.Hash, err = hashSnapshotV1Files(t.TempDir())
	require.NoError(t, err)
	directory := v1ModuleCachePath(cache.root, entry)
	require.NoError(t, os.MkdirAll(directory, 0o755))
	require.NoError(t, cache.VerifyCached(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}))
	files, err = cache.SourceFiles(entry)
	require.NoError(t, err)
	require.NotNil(t, files)
	require.Empty(t, files)
}

func TestInstalledSnapshotInventoryIsReusedWithinRun(t *testing.T) {
	t.Parallel()
	remote, commit := diagnosticRepository(t)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), remote, commit)
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cache.Install(t.Context(), lock))
	files, err := cache.SourceFiles(fetched.Lock)
	require.NoError(t, err)
	assert.Equal(t, []string{"api.proto", "easyp.yaml"}, files)
	files[0] = "caller mutation"
	files, err = cache.WithLogger(nil).SourceFiles(fetched.Lock)
	require.NoError(t, err)
	assert.Equal(t, []string{"api.proto", "easyp.yaml"}, files)

	other := fetched.Lock
	other.Roots = []string{"elsewhere"}
	files, err = cache.SourceFiles(other)
	require.NoError(t, err)
	assert.Nil(t, files)

	directory, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "api.proto"), []byte("tampered"), 0o644))
	// Ordinary use trusts an installed snapshot for this command. A fresh handle
	// and explicit verification still detect corruption, without changing files.
	require.NoError(t, cache.Install(t.Context(), lock))
	fresh := &Cache{root: cache.root}
	files, err = fresh.SourceFiles(fetched.Lock)
	require.NoError(t, err)
	assert.Nil(t, files)
	require.ErrorContains(t, fresh.Install(t.Context(), lock), "hash mismatch")
	require.ErrorContains(t, cache.VerifyCached(t.Context(), lock), "hash mismatch")
	files, err = cache.SourceFiles(fetched.Lock)
	require.NoError(t, err)
	assert.Nil(t, files)
	require.ErrorContains(t, cache.Install(t.Context(), lock), "hash mismatch")
}

func TestInstalledSnapshotConcurrentReaders(t *testing.T) {
	t.Parallel()
	remote, commit := diagnosticRepository(t)
	producer := New(t.TempDir())
	fetched, err := producer.Fetch(t.Context(), remote, commit)
	require.NoError(t, err)
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	cache := &Cache{root: producer.root}
	var group sync.WaitGroup
	failures := make(chan error, 16)
	for range 16 {
		group.Go(func() {
			if err := cache.Install(t.Context(), lock); err != nil {
				failures <- err
				return
			}
			files, err := cache.SourceFiles(fetched.Lock)
			if err != nil {
				failures <- err
				return
			}
			if len(files) != 2 || files[0] != "api.proto" || files[1] != "easyp.yaml" {
				failures <- fmt.Errorf("unexpected inventory %v", files)
			}
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
}
