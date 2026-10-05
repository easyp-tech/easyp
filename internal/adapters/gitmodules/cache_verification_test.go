package gitmodules

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestCacheVerificationStampInvalidatesOnContentChange(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	name := filepath.Join(root, "api.proto")
	require.NoError(t, os.WriteFile(name, []byte("same"), 0o644))
	hash, err := hashV1Files(root, []string{"api.proto"})
	require.NoError(t, err)
	entry := v1.LockedModule{Source: "example.test/cache", Version: "v1.0.0", Commit: "1111111111111111111111111111111111111111", Hash: hash}

	require.NoError(t, verifyInstalledV1Module(root, entry))
	assert.FileExists(t, cacheVerificationStampPath(root))

	info, err := os.Stat(name)
	require.NoError(t, err)
	originalModTime := info.ModTime()
	require.NoError(t, os.WriteFile(name, []byte("evil"), 0o644))
	require.NoError(t, os.Chtimes(name, time.Now().Add(-time.Hour), originalModTime))

	err = verifyInstalledV1Module(root, entry)
	require.ErrorContains(t, err, "hash mismatch")
}

func TestCacheTreeFingerprintStableWithoutChanges(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "proto"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "proto", "api.proto"), []byte("content"), 0o644))

	first, supported, err := cacheTreeFingerprint(root)
	require.NoError(t, err)
	second, secondSupported, err := cacheTreeFingerprint(root)
	require.NoError(t, err)

	assert.Equal(t, supported, secondSupported)
	assert.Equal(t, first, second)
}
