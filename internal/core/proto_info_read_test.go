package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestReadFileFromImportSearchOrder(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		local       bool
		firstRoot   bool
		secondRoot  bool
		wantPackage PackageName
	}{
		{name: "local before dependencies", local: true, firstRoot: true, secondRoot: true, wantPackage: "local"},
		{name: "first dependency root", firstRoot: true, secondRoot: true, wantPackage: "first"},
		{name: "skip missing dependency root", secondRoot: true, wantPackage: "second"},
		{name: "well-known imports last", wantPackage: "google.protobuf"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			project, first, second := t.TempDir(), t.TempDir(), t.TempDir()
			const importName = "google/protobuf/empty.proto"
			for _, source := range []struct {
				root    string
				pkg     string
				present bool
			}{
				{root: project, pkg: "local", present: tc.local},
				{root: first, pkg: "first", present: tc.firstRoot},
				{root: second, pkg: "second", present: tc.secondRoot},
			} {
				if !source.present {
					continue
				}
				path := filepath.Join(source.root, importName)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				content := fmt.Sprintf("syntax = \"proto3\"; package %s; message Empty {}", source.pkg)
				require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
			}
			app := &Core{logger: logger.NewNop(), importRoots: []string{first, second}}
			file, err := app.readFileFromImport(t.Context(), fs.NewFSWalker(project, "."), importName)
			require.NoError(t, err)
			require.Equal(t, tc.wantPackage, GetPackageName(file))
		})
	}
}

func TestReadFileFromImportClosesOpenedFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		content string
		wantErr bool
	}{
		{name: "valid proto", content: `syntax = "proto3"; message Item {}`},
		{name: "invalid proto", content: "invalid proto", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := &trackedImportFile{Reader: strings.NewReader(tc.content)}
			disk := importTestFS{open: func(string) (io.ReadCloser, error) { return file, nil }}
			app := &Core{logger: logger.NewNop()}
			_, err := app.readFileFromImport(t.Context(), disk, "item.proto")
			if tc.wantErr {
				require.ErrorContains(t, err, "item.proto")
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, 1, file.closes)
		})
	}
}

func TestReadFileFromImportPreservesErrors(t *testing.T) {
	t.Parallel()

	app := &Core{logger: logger.NewNop()}
	disk := importTestFS{open: func(string) (io.ReadCloser, error) {
		return nil, fmt.Errorf("denied: %w", os.ErrPermission)
	}}
	_, err := app.readFileFromImport(t.Context(), disk, "google/protobuf/empty.proto")
	require.ErrorIs(t, err, os.ErrPermission)

	_, err = app.readFileFromImport(t.Context(), fs.NewFSWalker(t.TempDir(), "."), "missing.proto")
	var missing *OpenImportFileError
	require.ErrorAs(t, err, &missing)
	require.Equal(t, "missing.proto", missing.FileName)
}

type importTestFS struct {
	FS
	open func(string) (io.ReadCloser, error)
}

func (f importTestFS) Open(name string) (io.ReadCloser, error) {
	return f.open(name)
}

type trackedImportFile struct {
	*strings.Reader
	closes int
}

func (f *trackedImportFile) Close() error {
	f.closes++
	return nil
}
