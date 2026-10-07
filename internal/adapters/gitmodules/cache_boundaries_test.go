package gitmodules

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceCacheDirectoriesOwnsOnlyGitStorage(t *testing.T) {
	t.Parallel()
	storage := t.TempDir()
	cache := New(storage)

	assert.Equal(t, []string{filepath.Join(storage, "v1/git")}, cache.SourceCacheDirectories())
	directories := cache.SourceCacheDirectories()
	directories[0] = "changed"
	assert.Equal(t, []string{filepath.Join(storage, "v1/git")}, cache.SourceCacheDirectories())
}
