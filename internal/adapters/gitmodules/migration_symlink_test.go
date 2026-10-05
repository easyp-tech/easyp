package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMigrationOmitsAuxiliarySymlinks(t *testing.T) {
	t.Parallel()

	for _, target := range []string{"internal", "external", "dangling"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			for _, mode := range []struct {
				name   string
				native bool
				pinned bool
			}{
				{name: "legacy_initial"},
				{name: "legacy_proto_archive_pin", pinned: true},
				{name: "native_initial", native: true},
				{name: "native_pin", native: true, pinned: true},
			} {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()
					files := map[string]string{
						"proto/file.proto": "syntax = \"proto3\";\n",
						"LICENSE":          "retained regular auxiliary file\n",
					}
					if !mode.native {
						files["easyp.yaml"] = "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n"
					}
					repository, _ := migrationTestRepository(t, files)
					if mode.native {
						files["protobuf.mod"] = "module " + repository + "\nroots proto\n"
						migrationTestWriteFiles(t, repository, map[string]string{"protobuf.mod": files["protobuf.mod"]})
					}
					linkTarget := "../proto/file.proto"
					if target != "internal" {
						outside := t.TempDir()
						linkTarget = filepath.Join(outside, "file.proto")
						if target == "external" {
							// The pointer is auxiliary even when its target is a proto.
							// Its bytes must never enter the migration or native digest.
							require.NoError(t, os.WriteFile(linkTarget, []byte("unread external proto"), 0o000))
						}
					}
					commit := migrationTestCommitSymlink(t, repository, "example-workspace/.bazelrc", linkTarget)
					var version, legacyHash string
					if mode.pinned {
						version = commit
						if !mode.native {
							legacyHash = migrationTestHash(t, map[string]string{"file.proto": files["proto/file.proto"]})
						}
					}
					cacheDir := t.TempDir()
					fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, version, legacyHash)
					require.NoError(t, err)
					assert.Equal(t, commit, fetched.Lock.Commit)
					assert.Equal(t, commit, fetched.Lock.Version)
					assert.Equal(t, []string{"proto"}, fetched.Module.Roots)
					assert.Equal(t, migrationTestHash(t, files), fetched.Lock.Hash)
					ordinary, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, commit)
					require.NoError(t, err)
					assert.Equal(t, ordinary.Lock.Hash, fetched.Lock.Hash)
					migrationTestAssertNoCheckout(t, cacheDir)
				})
			}
		})
	}
}

func TestFetchMigrationAuxiliarySymlinksRequireProtoArchiveHash(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		regularPin bool
	}{
		{name: "regular_only_tree_is_not_historical_proof", regularPin: true},
		{name: "wrong_proto_archive_hash"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{"file.proto": "syntax = \"proto3\";", "README": "tracked regular auxiliary content"}
			repository, _ := migrationTestRepository(t, files)
			commit := migrationTestCommitSymlink(t, repository, ".bazelrc", "file.proto")
			legacyHash := migrationTestHash(t, map[string]string{"file.proto": "different proto content"})
			if tt.regularPin {
				legacyHash = migrationTestHash(t, files)
			}
			cacheDir := t.TempDir()
			fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, commit, legacyHash)
			require.ErrorContains(t, err, "legacy hash mismatch")
			assert.Empty(t, fetched.Lock.Source)
			migrationTestAssertNoCheckout(t, cacheDir)
		})
	}
}

func TestFetchMigrationRejectsSourceAndMetadataSymlinks(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		files  map[string]string
		link   string
		target string
		root   string
	}{
		{name: "proto_link", files: map[string]string{"file.proto": "proto"}, link: "unsafe.proto", target: "file.proto"},
		{name: "root_directory", files: map[string]string{"easyp.yaml": "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n", "real/file.proto": "proto"}, link: "proto", target: "real"},
		{name: "root_ancestor", files: map[string]string{"easyp.yaml": "generate:\n  inputs: [{directory: {path: api/proto, root: api/proto}}]\n", "real/proto/file.proto": "proto"}, link: "api", target: "real"},
		{name: "native_root_directory", files: map[string]string{"real/file.proto": "proto"}, link: "proto", target: "real", root: "proto"},
		{name: "native_root_ancestor", files: map[string]string{"real/proto/file.proto": "proto"}, link: "api", target: "real", root: "api/proto"},
		{name: "root_metadata", files: map[string]string{"config.yaml": "generate: {}\n", "file.proto": "proto"}, link: "easyp.yaml", target: "config.yaml"},
		{name: "nested_metadata", files: map[string]string{"manifest": "direct ()\n", "file.proto": "proto"}, link: "nested/protobuf.mod", target: "../manifest"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range []struct {
				name   string
				pinned bool
			}{{name: "initial"}, {name: "pinned", pinned: true}} {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()
					repository, _ := migrationTestRepository(t, tt.files)
					if tt.root != "" {
						migrationTestWriteFiles(t, repository, map[string]string{"protobuf.mod": "module " + repository + "\nroots " + tt.root + "\n"})
					}
					commit := migrationTestCommitSymlink(t, repository, tt.link, tt.target)
					var version, legacyHash string
					if mode.pinned {
						version, legacyHash = commit, "h1:untrusted"
					}
					cacheDir := t.TempDir()
					_, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, version, legacyHash)
					require.ErrorContains(t, err, "non-regular")
					migrationTestAssertNoCheckout(t, cacheDir)
				})
			}
		})
	}
}

func TestMigrationTrackedFilesHandlesMaterializedSymlinks(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		link    string
		target  string
		root    string
		wantErr bool
	}{
		{name: "auxiliary", link: "example-workspace/.bazelrc", target: "../real/file.proto", root: "."},
		{name: "proto", link: "link.proto", target: "real/file.proto", root: ".", wantErr: true},
		{name: "root", link: "proto", target: "real", root: "proto", wantErr: true},
		{name: "root_ancestor", link: "api", target: "real", root: "api/proto", wantErr: true},
		{name: "metadata", link: "buf.lock", target: "real/file.proto", root: ".", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, _ := migrationTestRepository(t, map[string]string{"real/file.proto": "proto", "README": "regular auxiliary"})
			migrationTestCommitSymlink(t, repository, tt.link, tt.target)
			checkout := t.TempDir()
			runTestGit(t, checkout, "clone", "--quiet", "--no-checkout", "--", repository, checkout)
			runTestGit(t, checkout, "-c", "core.symlinks=false", "checkout", "--quiet", "HEAD")
			info, err := os.Lstat(filepath.Join(checkout, filepath.FromSlash(tt.link)))
			require.NoError(t, err)
			require.True(t, info.Mode().IsRegular())
			files, auxiliarySymlinks, err := migrationTrackedFiles(t.Context(), checkout, tt.root)
			if tt.wantErr {
				require.ErrorContains(t, err, "non-regular")
				return
			}
			require.NoError(t, err)
			assert.True(t, auxiliarySymlinks)
			assert.ElementsMatch(t, []string{"real/file.proto", "README"}, files)
		})
	}
}

func TestFetchMigrationAuxiliarySymlinksPreserveArchiveAttributes(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name      string
		attribute string
		source    string
	}{
		{name: "export_ignore", attribute: "proto/affected.proto export-ignore\n", source: "syntax = \"proto3\";"},
		{name: "export_subst", attribute: "proto/affected.proto export-subst\n", source: "// $Format:%H$\nsyntax = \"proto3\";"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range []struct {
				name   string
				pinned bool
			}{{name: "initial"}, {name: "pinned", pinned: true}} {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()
					files := map[string]string{
						"easyp.yaml":           "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
						".gitattributes":       tt.attribute,
						"proto/kept.proto":     "syntax = \"proto3\";",
						"proto/affected.proto": tt.source,
					}
					repository, _ := migrationTestRepository(t, files)
					commit := migrationTestCommitSymlink(t, repository, ".bazelrc", "missing")
					var version, legacyHash string
					if mode.pinned {
						installed := map[string]string{"kept.proto": files["proto/kept.proto"]}
						if tt.name == "export_subst" {
							installed["affected.proto"] = strings.ReplaceAll(tt.source, "$Format:%H$", commit)
						}
						version, legacyHash = commit, migrationTestHash(t, installed)
					}
					cacheDir := t.TempDir()
					_, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, version, legacyHash)
					require.ErrorContains(t, err, "source selection")
					require.ErrorContains(t, err, "manual migration")
					migrationTestAssertNoCheckout(t, cacheDir)
				})
			}
		})
	}
}

func migrationTestCommitSymlink(t *testing.T, repository, name, target string) string {
	t.Helper()
	link := filepath.Join(repository, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o755))
	require.NoError(t, os.Symlink(target, link))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "symlink fixture")
	return strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
}

func migrationTestAssertNoCheckout(t *testing.T, cacheDir string) {
	t.Helper()
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.Equal(t, "objects", entry.Name(), "temporary checkout was left behind")
	}
}
