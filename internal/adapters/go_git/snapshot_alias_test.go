package go_git

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
)

func TestSnapshotRevisionMaterializesCommittedAliases(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, content := range map[string]string{
		"real/item.proto":        "committed proto bytes\n",
		"real/manifest.txt":      "module example.test/app\nroots api\n",
		"real/policy-source.txt": "version: v1\n",
	} {
		filename := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
		require.NoError(t, os.WriteFile(filename, []byte(content), 0o644))
	}
	for name, target := range map[string]string{
		"alias.proto":  "real/item.proto",
		"api":          "real",
		"protobuf.mod": "real/manifest.txt",
		"easyp.yaml":   "real/policy-source.txt",
		"README-link":  "../outside",
	} {
		require.NoError(t, os.Symlink(target, filepath.Join(root, name)))
	}
	snapshotGit(t, root, "init", "-q", "-b", "baseline")
	snapshotGit(t, root, "add", ".")
	snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "aliases")
	require.NoError(t, os.WriteFile(filepath.Join(root, "real/item.proto"), []byte("working tree changed\n"), 0o644))

	baseline, err := SnapshotRevision(t.Context(), root, "baseline")

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, baseline.Close()) })
	for name, expected := range map[string]string{
		"alias.proto":    "committed proto bytes\n",
		"api/item.proto": "committed proto bytes\n",
		"protobuf.mod":   "module example.test/app\nroots api\n",
		"easyp.yaml":     "version: v1\n",
	} {
		filename := filepath.Join(baseline.Root, name)
		actual, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, expected, string(actual))
		info, err := os.Lstat(filename)
		require.NoError(t, err)
		assert.True(t, info.Mode().IsRegular(), name)
	}
	assert.NoFileExists(t, filepath.Join(baseline.Root, "README-link"))
}

func TestSnapshotRevisionRejectsRequiredUnsafeAliases(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		links    map[string]string
		manifest bool
	}{
		{name: "external proto", links: map[string]string{"alias.proto": "../outside.proto"}},
		{name: "dangling metadata", links: map[string]string{"protobuf.mod": "missing"}},
		{name: "proto cycle", links: map[string]string{"alias.proto": "other.proto", "other.proto": "alias.proto"}},
		{name: "declared directory cycle", links: map[string]string{"proto": "."}, manifest: true},
		{name: "declared dangling root", links: map[string]string{"proto": "missing"}, manifest: true},
		{name: "declared external root", links: map[string]string{"proto": "../outside"}, manifest: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, "item.proto"), []byte("committed\n"), 0o644))
			if tt.manifest {
				require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte("module example.test/app\nroots proto\n"), 0o644))
			}
			for name, target := range tt.links {
				require.NoError(t, os.Symlink(target, filepath.Join(root, name)))
			}
			snapshotGit(t, root, "init", "-q", "-b", "baseline")
			snapshotGit(t, root, "add", ".")
			snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "unsafe aliases")

			baseline, err := SnapshotRevision(t.Context(), root, "baseline")

			require.Error(t, err)
			assert.Nil(t, baseline)
		})
	}
}

func TestSnapshotRevisionIgnoresUnselectedDirectoryCycles(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, path, marker string }{
		{name: "hidden", path: "proto/.aux/cycle"},
		{name: "vendor", path: "proto/easyp_vendor/cycle"},
		{name: "nested module", path: "proto/other/cycle", marker: "proto/other/protobuf.mod"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(root, tt.path)), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte("module example.test/app\nroots proto\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(root, "proto/item.proto"), []byte("syntax = \"proto3\"; message Item {}\n"), 0o644))
			if tt.marker != "" {
				require.NoError(t, os.WriteFile(filepath.Join(root, tt.marker), []byte("module example.test/other\nroots absent\n"), 0o644))
			}
			require.NoError(t, os.Symlink("..", filepath.Join(root, tt.path)))
			snapshotGit(t, root, "init", "-q", "-b", "baseline")
			snapshotGit(t, root, "add", ".")
			snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "unused directory cycles")
			baseline, err := SnapshotRevision(t.Context(), root, "baseline")
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, baseline.Close()) })
			assert.FileExists(t, filepath.Join(baseline.Root, "proto/item.proto"))
		})
	}
}

func TestSnapshotRevisionRejectsSelectedGitlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "proto"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte("module example.test/app\nroots proto/submodule\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "proto/item.proto"), []byte("syntax = \"proto3\"; message Item {}\n"), 0o644))
	snapshotGit(t, root, "init", "-q", "-b", "baseline")
	snapshotGit(t, root, "add", ".")
	snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "sources")
	repo, err := gogit.PlainOpen(root)
	require.NoError(t, err)
	head, err := repo.Head()
	require.NoError(t, err)
	snapshotGit(t, root, "update-index", "--add", "--cacheinfo", "160000,"+head.Hash().String()+",proto/submodule")
	snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "selected submodule")
	baseline, err := SnapshotRevision(t.Context(), root, "baseline")
	require.ErrorIs(t, err, gitsnapshot.ErrGitlink)
	assert.Nil(t, baseline)
}
