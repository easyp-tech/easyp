package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestReadFileFromImportSearchOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		localPackage  string
		firstPackage  string
		secondPackage string
		wantPackage   PackageName
	}{
		{name: "local_before_dependencies", localPackage: "local", firstPackage: "first", secondPackage: "second", wantPackage: "local"},
		{name: "first_dependency_root", firstPackage: "first", secondPackage: "second", wantPackage: "first"},
		{name: "skip_missing_dependency_root", secondPackage: "second", wantPackage: "second"},
		{name: "well_known_imports_last", wantPackage: "google.protobuf"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			project, first, second := t.TempDir(), t.TempDir(), t.TempDir()
			const importName = "google/protobuf/empty.proto"
			for root, pkg := range map[string]string{project: tt.localPackage, first: tt.firstPackage, second: tt.secondPackage} {
				if pkg == "" {
					continue
				}
				content := fmt.Sprintf("syntax = \"proto3\"; package %s; message Empty {}", pkg)
				writeImportSource(t, root, importName, content)
			}
			app := &Core{logger: logger.NewNop(), importRoots: []string{first, second}}

			file, err := app.readFileFromImport(t.Context(), fs.NewFSWalker(project, "."), importName)

			require.NoError(t, err)
			assert.Equal(t, tt.wantPackage, GetPackageName(file))
		})
	}
}

func TestReadFileFromImportClosesOpenedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantMessage string
	}{
		{name: "valid_proto", content: `syntax = "proto3"; message Item {}`},
		{name: "invalid_proto", content: "invalid proto", wantMessage: "parse item.proto"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := writeImportSource(t, t.TempDir(), "item.proto", tt.content)
			file, err := os.Open(path)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := file.Close()
				if !errors.Is(err, os.ErrClosed) {
					assert.NoError(t, err)
				}
			})
			disk := &mockImportFS{file: file}
			app := &Core{logger: logger.NewNop()}

			_, err = app.readFileFromImport(t.Context(), disk, "item.proto")

			if tt.wantMessage != "" {
				assert.ErrorContains(t, err, tt.wantMessage)
			} else {
				assert.NoError(t, err)
			}
			_, err = file.Read(make([]byte, 1))
			assert.ErrorIs(t, err, os.ErrClosed)
		})
	}
}

func TestReadFileFromImportPreservesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		importName  string
		openErr     error
		wantErr     error
		wantMissing bool
		wantMessage string
	}{
		{
			name:       "permission_error_prevents_fallback",
			importName: "google/protobuf/empty.proto",
			openErr:    fmt.Errorf("denied: %w", os.ErrPermission),
			wantErr:    os.ErrPermission,
		},
		{name: "missing_import", importName: "missing.proto", openErr: os.ErrNotExist, wantMissing: true},
		{name: "path_leaves_root", importName: "../item.proto", openErr: os.ErrPermission, wantMessage: "invalid import path"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := &Core{logger: logger.NewNop()}
			disk := &mockImportFS{err: tt.openErr}

			_, err := app.readFileFromImport(t.Context(), disk, tt.importName)

			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantMissing:
				var missing *OpenImportFileError
				require.ErrorAs(t, err, &missing)
				assert.Equal(t, tt.importName, missing.FileName)
			default:
				require.ErrorContains(t, err, tt.wantMessage)
			}
		})
	}
}

func writeImportSource(t *testing.T, root, name, content string) string {
	t.Helper()

	path := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

type mockImportFS struct {
	FS
	file io.ReadCloser
	err  error
}

func (f *mockImportFS) Open(string) (io.ReadCloser, error) {
	return f.file, f.err
}
