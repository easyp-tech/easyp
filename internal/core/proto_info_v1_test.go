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
	tests := []struct {
		name       string
		importPath string
	}{
		{name: "root import", importPath: "dep.proto"},
		{name: "nested import", importPath: "dep/v1/dep.proto"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			project, dependency := t.TempDir(), t.TempDir()
			path := filepath.Join(dependency, tt.importPath)
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			require.NoError(t, os.WriteFile(path, []byte("syntax = \"proto3\"; package dep.v1; message Dep {}\n"), 0o644))
			roots := []string{dependency}
			app := New(Options{Logger: logger.NewNop(), ImportRoots: roots})
			roots[0] = t.TempDir()

			parsed, err := app.readFileFromImport(t.Context(), fs.NewFSWalker(project, "."), tt.importPath)

			require.NoError(t, err)
			require.NotNil(t, parsed)
		})
	}
}
