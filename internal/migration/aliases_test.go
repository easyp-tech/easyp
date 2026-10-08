package migration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/easyp-tech/easyp/internal/sourceview"
	"github.com/stretchr/testify/require"
)

func TestBuildLinkedMetadataPreservesTarget(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "policy-source.yaml", "lint: {}\n", 0o640)
	require.NoError(t, os.Symlink("policy-source.yaml", filepath.Join(root, "easyp.yaml")))
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/api"})
	require.NoError(t, err)
	require.NoError(t, plan.Apply())
	assertTransactionFile(t, root, "policy-source.yaml", "lint: {}\n", 0o640)
	assertTransactionFile(t, root, "easyp.yaml.v0.bak", "lint: {}\n", 0o640)
	info, err := os.Lstat(filepath.Join(root, "easyp.yaml"))
	require.NoError(t, err)
	require.True(t, info.Mode().IsRegular())
}

func TestLinkedMetadataRollbackRestoresPointer(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "policy-source.yaml", "lint: {}\n", 0o640)
	require.NoError(t, os.Symlink("policy-source.yaml", filepath.Join(root, "easyp.yaml")))
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/api"})
	require.NoError(t, err)
	plan.tx.rename = func(root *os.Root, from, to string) error {
		if to == "protobuf.mod" {
			return errors.New("injected failure")
		}
		return root.Rename(from, to)
	}
	plan.tx.link = func(root *os.Root, from, to string) error {
		if to == "protobuf.mod" {
			return errors.New("injected failure")
		}
		return root.Link(from, to)
	}
	require.ErrorContains(t, plan.Apply(), "injected failure")
	pointer, err := os.Readlink(filepath.Join(root, "easyp.yaml"))
	require.NoError(t, err)
	require.Equal(t, "policy-source.yaml", pointer)
	assertTransactionFile(t, root, "policy-source.yaml", "lint: {}\n", 0o640)
}

func TestLocalSelectionPreservesDirectoryAliasNames(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "sources"), 0o755))
	writeTransactionFile(t, root, "sources/model.proto", "syntax = \"proto3\"; package model;", 0o644)
	require.NoError(t, os.Symlink("sources", filepath.Join(root, "proto")))
	require.NoError(t, os.Symlink("missing", filepath.Join(root, "sources", "unused.txt")))
	selection, err := proveLocalSelection(root, []legacyDirectory{{Root: "proto", Path: "."}}, []string{"proto"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"model.proto": "proto/model.proto"}, selection.files)
}

func TestLinkedInputRechecksPointerAndTarget(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{name: "pointer", mutate: func(t *testing.T, root string) {
			writeTransactionFile(t, root, "other-policy.yaml", "lint: {}\n", 0o640)
			require.NoError(t, os.Remove(filepath.Join(root, "easyp.yaml")))
			require.NoError(t, os.Symlink("other-policy.yaml", filepath.Join(root, "easyp.yaml")))
		}},
		{name: "target inode", mutate: func(t *testing.T, root string) {
			writeTransactionFile(t, root, "other-policy.yaml", "lint: {}\n", 0o640)
			require.NoError(t, os.Rename(filepath.Join(root, "other-policy.yaml"), filepath.Join(root, "policy-source.yaml")))
		}},
		{name: "target mode", mutate: func(t *testing.T, root string) {
			require.NoError(t, os.Chmod(filepath.Join(root, "policy-source.yaml"), 0o600))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeTransactionFile(t, root, "policy-source.yaml", "lint: {}\n", 0o640)
			require.NoError(t, os.Symlink("policy-source.yaml", filepath.Join(root, "easyp.yaml")))
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/api"})
			require.NoError(t, err)
			tt.mutate(t, root)
			require.Error(t, plan.Apply())
			require.NoFileExists(t, filepath.Join(root, "protobuf.mod"))
		})
	}
}

func TestNativeLinkedMetadataNoopPreservesTopology(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "policy-source.yaml", "version: v1\nlinters:\n  default: MINIMAL\n", 0o640)
	writeTransactionFile(t, root, "easyp.gen.yaml", "version: v1\n", 0o644)
	writeTransactionFile(t, root, "manifest-source", "module example.com/api\nroots .\n", 0o600)
	require.NoError(t, os.Symlink("policy-source.yaml", filepath.Join(root, "easyp.yaml")))
	require.NoError(t, os.Symlink("manifest-source", filepath.Join(root, "protobuf.mod")))
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/api"})
	require.NoError(t, err)
	require.True(t, plan.AlreadyV1())
	require.NoError(t, plan.Apply())
	pointer, err := os.Readlink(filepath.Join(root, "easyp.yaml"))
	require.NoError(t, err)
	require.Equal(t, "policy-source.yaml", pointer)
	pointer, err = os.Readlink(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Equal(t, "manifest-source", pointer)
}

func TestBuildAbsoluteInternalSourceAlias(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "easyp.yaml", "generate:\n  inputs:\n    - directory:\n        root: proto\n        path: .\n", 0o644)
	require.NoError(t, os.Mkdir(filepath.Join(root, "proto"), 0o755))
	writeTransactionFile(t, root, "model-source.proto", "syntax = \"proto3\"; package model.v1; message Model {}", 0o644)
	require.NoError(t, os.Symlink(filepath.Join(root, "model-source.proto"), filepath.Join(root, "proto/model.proto")))
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/api"})
	require.NoError(t, err)
	require.NoError(t, plan.Apply())
	require.Equal(t, map[string]string{"model.proto": "proto/model.proto"}, plan.local.selection.files)
}

func TestTransactionRejectsOutputOverlappingLinkedInputHop(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "source.proto", "original source", 0o644)
	require.NoError(t, os.Symlink("source.proto", filepath.Join(root, "protobuf.mod")))
	require.NoError(t, os.Symlink("protobuf.mod", filepath.Join(root, "alias.proto")))
	tx, err := newTransaction(root)
	require.NoError(t, err)
	_, err = tx.captureInput("alias.proto")
	require.NoError(t, err)
	_, err = tx.captureInput("protobuf.mod")
	require.NoError(t, err)
	tx.changes = []fileChange{{name: "protobuf.mod", content: []byte("replacement"), mode: 0o644}}
	require.ErrorContains(t, tx.apply(), "overlaps linked input hop")
	pointer, err := os.Readlink(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Equal(t, "source.proto", pointer)
}

func TestBuildSourceRootDirectoryAliasCycleFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTransactionFile(t, root, "easyp.yaml", "generate:\n  inputs:\n    - directory:\n        root: proto\n        path: .\n", 0o644)
	writeTransactionFile(t, root, "file.proto", "syntax = \"proto3\"; package model.v1; message Model {}", 0o644)
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o755))
	require.NoError(t, os.Symlink(".", filepath.Join(root, "proto")))
	_, err := Build(context.Background(), Options{Dir: root, Module: "example.com/api"})
	require.ErrorIs(t, err, sourceview.ErrCycle)
}
