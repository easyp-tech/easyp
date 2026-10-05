package core

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

func TestSourceImportPathDoesNotDependOnScanRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, tt := range []struct{ name, scanRoot, path string }{
		{name: "repository", scanRoot: root, path: "proto/foo/v1/item.proto"},
		{name: "source root", scanRoot: filepath.Join(root, "proto"), path: "foo/v1/item.proto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			engine := &Core{importRoots: []string{root, filepath.Join(root, "proto"), filepath.Join(root, "dependency")}}
			got := engine.sourceImportPath(disk.NewFSWalker(tt.scanRoot, "."), tt.path)
			require.Equal(t, "foo/v1/item.proto", got)
		})
	}
}
