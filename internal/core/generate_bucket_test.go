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
