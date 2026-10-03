package migration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/modules"
)

func TestEmptyHistoricalLockGetsNativeLock(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFixture(t, root, "easyp.yaml", "lint: {}\n")
	writeFixture(t, root, "easyp.lock", "")
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/app"})
	require.NoError(t, err)
	require.NoError(t, plan.Apply())
	lock, err := modules.ReadLock(filepath.Join(root, "protobuf.lock"))
	require.NoError(t, err)
	require.Equal(t, 1, lock.Version)
	require.Empty(t, lock.Modules)
}
