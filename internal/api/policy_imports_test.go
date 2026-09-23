package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestPolicyImportRootsIncludeModuleAndLocalDependency(t *testing.T) {
	root := t.TempDir()
	dep := t.TempDir()
	replacement, err := filepath.Rel(root, dep)
	require.NoError(t, err)
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/root\nroots proto\nrequire example.com/dep v1.0.0\nreplace example.com/dep => "+replacement+"\n")
	writeV1GenerateFixture(t, root, "proto/root.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, dep, "protobuf.mod", "module example.com/dep\nroots src\n")
	writeV1GenerateFixture(t, dep, "src/dep.proto", "syntax = \"proto3\";\n")

	moduleDir, err := findV1PolicyModuleDir(root, filepath.Join(root, "proto"))
	require.NoError(t, err)
	require.Equal(t, root, moduleDir)
	roots, err := resolveV1PolicyImportRoots(t.Context(), logger.NewNop(), moduleDir)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{filepath.Join(root, "proto"), filepath.Join(dep, "src")}, roots)
}

func TestPolicyImportRootsAllowNoManifest(t *testing.T) {
	root := t.TempDir()
	moduleDir, err := findV1PolicyModuleDir(root, root)
	require.NoError(t, err)
	require.Empty(t, moduleDir)
}
