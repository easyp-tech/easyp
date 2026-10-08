package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestResolvedWriterOwnsCapturedBytes(t *testing.T) {
	t.Parallel()
	root, tx := resolvedTransactionFixture(t)
	observed, err := tx.capture("consumer.proto", nil)
	require.NoError(t, err)

	// A caller may reuse its observation buffer without changing the files
	// whose captured bytes still justify the transaction's pending writes.
	observed.data[0] ^= 0xff
	err = tx.apply()

	require.NoError(t, err)
	for _, name := range []string{"consumer.proto", v1.ModuleFile, v1.LockFile} {
		data, err := os.ReadFile(filepath.Join(root, name))
		require.NoError(t, err)
		assert.Equal(t, "new "+name, string(data))
	}
	info, err := os.Stat(filepath.Join(root, "consumer.proto"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
}
