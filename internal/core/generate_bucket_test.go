package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBucketCancellationPreventsWrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	bucket := NewGenerateBucket()
	output := filepath.Join(root, "out.txt")
	bucket.PutFile(ctx, output, []byte("data"))
	require.ErrorIs(t, bucket.DumpToFs(ctx), context.Canceled)
	_, err := os.Stat(output)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestGenerateBucketValidationUsesPathOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for range 20 {
		bucket := NewGenerateBucket()
		bucket.PutFile(t.Context(), filepath.Join(root, "z.go"), []byte("package {"))
		bucket.PutFile(t.Context(), filepath.Join(root, "a.go"), []byte("package {"))
		require.ErrorContains(t, bucket.DumpToFs(t.Context()), filepath.Join(root, "a.go"))
	}
}

func TestGenerateBucketWriteErrorsUsePathOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"a-file", "m-file"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), nil, 0o644))
	}
	for range 20 {
		bucket := NewGenerateBucket()
		for i := range 50 {
			bucket.PutFile(t.Context(), filepath.Join(root, "ok", string(rune('a'+i%26))+string(rune('a'+i/26))+".txt"), []byte("data"))
		}
		bucket.PutFile(t.Context(), filepath.Join(root, "m-file", "x.txt"), []byte("data"))
		bucket.PutFile(t.Context(), filepath.Join(root, "a-file", "x.txt"), []byte("data"))
		err := bucket.DumpToFs(t.Context())
		require.ErrorContains(t, err, "os.MkdirAll "+filepath.Join(root, "a-file"))
		assert.NotContains(t, err.Error(), "m-file")
	}
	data, err := os.ReadFile(filepath.Join(root, "ok", "aa.txt"))
	require.NoError(t, err)
	assert.Equal(t, "data", string(data))
}
