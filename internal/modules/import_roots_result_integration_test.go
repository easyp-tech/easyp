package modules_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestTidyOwnsPinnedContentsAfterProviderBufferReuse(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, intrinsicServiceFiles(), nil)
	consumer := `syntax = "proto3"; import "svc.proto";`
	root := importRootsConsumer(t, repository, "", consumer)
	cache := &tidyReusedInspectionCache{Cache: gitmodules.New(t.TempDir())}

	err := modules.Tidy(t.Context(), root, cache)

	require.True(t, cache.reused, "the provider must reuse its returned inspection buffer")
	require.NoError(t, err)
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 1)
	assert.Equal(t, commit, lock.Modules[0].Commit)
	require.NoError(t, cache.VerifyCached(t.Context(), lock))
}

type tidyReusedInspectionCache struct {
	*gitmodules.Cache
	buffer []byte
	reused bool
}

func (cache *tidyReusedInspectionCache) FetchWithRoots(ctx context.Context, source, version string, roots []string) (modules.Fetched, error) {
	fetched, err := cache.Cache.FetchWithRoots(ctx, source, version, roots)
	if err != nil {
		return modules.Fetched{}, err
	}
	if fetched.Inspection != nil {
		for _, file := range fetched.Inspection.Files {
			if file.Path == "api/svc.proto" {
				cache.buffer = file.Content
			}
		}
	}
	return fetched, nil
}

func (cache *tidyReusedInspectionCache) Install(ctx context.Context, lock v1.Lock) error {
	if len(cache.buffer) > 0 && !cache.reused {
		cache.buffer[0] = 'X'
		cache.reused = true
	}
	return cache.Cache.Install(ctx, lock)
}
