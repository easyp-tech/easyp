package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func TestGetRenamer(t *testing.T) {
	tests := map[string]struct {
		moduleConfig   models.ModuleConfig
		passedFile     string
		expectedResult string
	}{
		"directories are empty": {
			moduleConfig: models.ModuleConfig{
				Directories: nil,
			},
			passedFile:     "proto/file.proto",
			expectedResult: "proto/file.proto",
		},
		"directories contain one dir": {
			moduleConfig: models.ModuleConfig{
				Directories: []string{"proto/protovalidate"},
			},
			passedFile:     "proto/protovalidate/buf/validate/validate.proto",
			expectedResult: "buf/validate/validate.proto",
		},
		"directories contain several dirs": {
			moduleConfig: models.ModuleConfig{
				Directories: []string{"proto/protovalidate", "proto/protovalidate-testing"},
			},
			passedFile:     "proto/protovalidate/buf/validate/validate.proto",
			expectedResult: "buf/validate/validate.proto",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			renamer := getRenamer(tc.moduleConfig)
			result := renamer(tc.passedFile)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestBuildInstallTree_RegularFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	renamer := getRenamer(models.ModuleConfig{
		Directories: []string{"proto"},
	})

	filePath := filepath.Join(srcDir, "proto", "grpc", "federation", "federation.proto")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
	require.NoError(t, os.WriteFile(filePath, []byte("syntax = \"proto3\";"), 0644))

	err := buildInstallTree(srcDir, dstDir, renamer)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dstDir, "grpc", "federation", "federation.proto"))
	require.NoError(t, err)
	require.Equal(t, "syntax = \"proto3\";", string(data))
}

func TestBuildInstallTree_SymlinkRewritten(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	renamer := getRenamer(models.ModuleConfig{
		Directories: []string{"proto"},
	})

	targetRelFromSrc := filepath.Join("proto", "grpc", "federation", "federation.proto")
	targetPath := filepath.Join(srcDir, targetRelFromSrc)
	require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
	require.NoError(t, os.WriteFile(targetPath, []byte("syntax = \"proto3\";"), 0644))

	linkDir := filepath.Join(srcDir, "compiler", "testdata")
	require.NoError(t, os.MkdirAll(linkDir, 0755))

	linkPath := filepath.Join(linkDir, "federation.proto")
	require.NoError(t, os.Symlink("../../proto/grpc/federation/federation.proto", linkPath))

	err := buildInstallTree(srcDir, dstDir, renamer)
	require.NoError(t, err)

	linkDst := filepath.Join(dstDir, "compiler", "testdata", "federation.proto")
	info, err := os.Lstat(linkDst)
	require.NoError(t, err)
	require.True(t, info.Mode()&os.ModeSymlink != 0, "expected symlink")

	resolved, err := filepath.EvalSymlinks(linkDst)
	require.NoError(t, err)

	targetInDst := filepath.Join(dstDir, "grpc", "federation", "federation.proto")
	resolvedTarget, err := filepath.EvalSymlinks(targetInDst)
	require.NoError(t, err)
	resolvedLink, err := filepath.EvalSymlinks(resolved)
	require.NoError(t, err)
	require.Equal(t, resolvedTarget, resolvedLink)
}

func TestBuildInstallTree_SymlinkOutsideRewrittenDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	renamer := getRenamer(models.ModuleConfig{
		Directories: []string{"proto"},
	})

	targetPath := filepath.Join(srcDir, "other", "file.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
	require.NoError(t, os.WriteFile(targetPath, []byte("hello"), 0644))

	linkPath := filepath.Join(srcDir, "link.txt")
	require.NoError(t, os.Symlink("other/file.txt", linkPath))

	err := buildInstallTree(srcDir, dstDir, renamer)
	require.NoError(t, err)

	linkDst := filepath.Join(dstDir, "link.txt")
	info, err := os.Lstat(linkDst)
	require.NoError(t, err)
	require.True(t, info.Mode()&os.ModeSymlink != 0)

	resolved, err := filepath.EvalSymlinks(linkDst)
	require.NoError(t, err)
	targetInDst := filepath.Join(dstDir, "other", "file.txt")
	resolvedTarget, err := filepath.EvalSymlinks(targetInDst)
	require.NoError(t, err)
	resolvedLink, err := filepath.EvalSymlinks(resolved)
	require.NoError(t, err)
	require.Equal(t, resolvedTarget, resolvedLink)
}

func TestBuildInstallTree_GrpcFederationLayout(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	renamer := getRenamer(models.ModuleConfig{
		Directories: []string{"proto"},
	})

	protoFile := filepath.Join(srcDir, "proto", "grpc", "federation", "federation.proto")
	require.NoError(t, os.MkdirAll(filepath.Dir(protoFile), 0755))
	require.NoError(t, os.WriteFile(protoFile, []byte("syntax = \"proto3\";"), 0644))

	testdataDir := filepath.Join(srcDir, "compiler", "testdata")
	require.NoError(t, os.MkdirAll(testdataDir, 0755))
	require.NoError(t, os.Symlink("../../proto/grpc/federation/federation.proto", filepath.Join(testdataDir, "federation.proto")))

	err := buildInstallTree(srcDir, dstDir, renamer)
	require.NoError(t, err)

	linkTarget, err := os.Readlink(filepath.Join(dstDir, "compiler", "testdata", "federation.proto"))
	require.NoError(t, err)
	require.NotEmpty(t, linkTarget)

	_, err = os.Stat(filepath.Join(dstDir, "grpc", "federation", "federation.proto"))
	require.NoError(t, err)

	linkDst := filepath.Join(dstDir, "compiler", "testdata", "federation.proto")
	resolved, err := filepath.EvalSymlinks(linkDst)
	require.NoError(t, err)
	data, err := os.ReadFile(resolved)
	require.NoError(t, err)
	require.Equal(t, "syntax = \"proto3\";", string(data))
}

func TestBuildInstallTree_AbsoluteSymlinkRejected(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	renamer := getRenamer(models.ModuleConfig{
		Directories: []string{"proto"},
	})

	absTarget := "/etc/passwd"
	if runtime.GOOS == "windows" {
		absTarget = `C:\Windows\System32\drivers\etc\hosts`
	}

	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "proto"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "proto", "file.proto"), []byte("data"), 0644))
	require.NoError(t, os.Symlink(absTarget, filepath.Join(srcDir, "proto", "bad.proto")))

	err := buildInstallTree(srcDir, dstDir, renamer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "absolute symlink target not allowed")
}

func TestBuildInstallTree_EscapingSymlinkRejected(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	renamer := getRenamer(models.ModuleConfig{
		Directories: []string{"proto"},
	})

	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "proto"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "proto", "file.proto"), []byte("data"), 0644))
	require.NoError(t, os.Symlink("../../../etc/passwd", filepath.Join(srcDir, "proto", "escape.proto")))

	err := buildInstallTree(srcDir, dstDir, renamer)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "symlink target escapes source tree") ||
		strings.Contains(err.Error(), "absolute symlink"))
}
