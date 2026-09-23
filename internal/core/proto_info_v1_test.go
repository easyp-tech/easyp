package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestReadFileFromImportUsesV1Roots(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	dependency := t.TempDir()
	path := filepath.Join(dependency, "dep.proto")
	require.NoError(t, os.WriteFile(path, []byte("syntax = \"proto3\"; package dep.v1; message Dep {}\n"), 0o644))
	app := &Core{logger: logger.NewNop()}
	app.SetImportRoots([]string{dependency})
	proto, err := app.readFileFromImport(t.Context(), fs.NewFSWalker(project, "."), "dep.proto")
	require.NoError(t, err)
	require.NotNil(t, proto)
}
