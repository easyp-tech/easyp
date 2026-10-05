package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
)

func TestFetchMigration(t *testing.T) {
	t.Parallel()

	config := "generate:\n  inputs:\n    - directory: {path: proto, root: proto}\n"
	tests := []struct {
		name      string
		version   string
		initial   bool
		badHash   bool
		useCommit bool
	}{
		{name: "pinned_commit", useCommit: true},
		{name: "semver", version: "v1.0.0"},
		{name: "initial_resolution", initial: true},
		{name: "mismatch_cleans_checkout", version: "v1.0.0", badHash: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"easyp.yaml":          config,
				"proto/example.proto": "syntax = \"proto3\";\n",
				"proto/README.md":     "retained non-proto content\n",
				"LICENSE":             "license outside import roots\n",
			}
			repository, commit := migrationTestRepository(t, files)
			runTestGit(t, repository, "tag", "v1.0.0")
			// Untracked checkout files never participated in the legacy archive.
			require.NoError(t, os.WriteFile(filepath.Join(repository, "untracked"), []byte("ignored"), 0o600))
			legacyHash := migrationTestHash(t, map[string]string{
				"easyp.yaml":    config,
				"example.proto": files["proto/example.proto"],
				"README.md":     files["proto/README.md"],
				"LICENSE":       files["LICENSE"],
			})
			if tt.initial {
				legacyHash = ""
			}
			if tt.badHash {
				legacyHash = migrationTestHash(t, map[string]string{"different": "tree"})
			}
			version := tt.version
			if tt.useCommit {
				version = commit
			}
			cacheDir := t.TempDir()
			fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, version, legacyHash)
			if tt.badHash {
				require.ErrorContains(t, err, "legacy hash mismatch")
			} else {
				require.NoError(t, err)
				assert.Equal(t, commit, fetched.Lock.Commit)
				assert.Equal(t, repository, fetched.Lock.Source)
				assert.Equal(t, repository, fetched.Module.Name)
				assert.Equal(t, []string{"proto"}, fetched.Module.Roots)
				assert.Equal(t, migrationTestHash(t, files), fetched.Lock.Hash)
				assert.NotEqual(t, legacyHash, fetched.Lock.Hash)
				if version == "" {
					assert.Equal(t, commit, fetched.Lock.Version)
				} else {
					assert.Equal(t, version, fetched.Lock.Version)
				}
			}
			entries, err := os.ReadDir(cacheDir)
			require.NoError(t, err)
			for _, entry := range entries {
				assert.Equal(t, "objects", entry.Name(), "temporary checkout was left behind")
			}
		})
	}
}

func TestHashMigrationLegacyFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		want    map[string]string
		wantErr string
	}{
		{
			name:  "ordered_roots_parent_first",
			files: map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a, a/b]\n", "a/b/message.proto": "message"},
			want:  map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a, a/b]\n", "b/message.proto": "message"},
		},
		{
			name:  "ordered_roots_child_first",
			files: map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a/b, a]\n", "a/b/message.proto": "message"},
			want:  map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a/b, a]\n", "message.proto": "message"},
		},
		{
			name:  "workspace_precedes_buf_module_and_easyp",
			files: map[string]string{"buf.work.yaml": "version: v1\ndirectories: [workspace]\n", "buf.yaml": "version: v2\nmodules: [{path: module}]\n", "easyp.yaml": "generate:\n  inputs: [{directory: {path: local, root: local}}]\n", "workspace/file.proto": "workspace", "module/README": "module", "local/file.proto": "local"},
			want:  map[string]string{"buf.work.yaml": "version: v1\ndirectories: [workspace]\n", "buf.yaml": "version: v2\nmodules: [{path: module}]\n", "easyp.yaml": "generate:\n  inputs: [{directory: {path: local, root: local}}]\n", "file.proto": "workspace", "module/README": "module", "local/file.proto": "local"},
		},
		{
			name:  "buf_module_precedes_easyp",
			files: map[string]string{"buf.yaml": "version: v2\nmodules: [{path: module}]\n", "easyp.yaml": "generate:\n  inputs: [{directory: {path: local, root: local}}]\n", "module/file.proto": "module", "local/file.proto": "local"},
			want:  map[string]string{"buf.yaml": "version: v2\nmodules: [{path: module}]\n", "easyp.yaml": "generate:\n  inputs: [{directory: {path: local, root: local}}]\n", "file.proto": "module", "local/file.proto": "local"},
		},
		{
			name:  "scalar_directory_does_not_strip_path",
			files: map[string]string{"easyp.yaml": "generate:\n  inputs: [{directory: proto}]\n", "proto/file.proto": "message"},
			want:  map[string]string{"easyp.yaml": "generate:\n  inputs: [{directory: proto}]\n", "proto/file.proto": "message"},
		},
		{
			name:  "unconfigured_repository_keeps_all_files",
			files: map[string]string{"proto/file.proto": "message", "README": "documentation"},
			want:  map[string]string{"proto/file.proto": "message", "README": "documentation"},
		},
		{
			name:    "file_collision",
			files:   map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a, b]\n", "a/file.proto": "a", "b/file.proto": "b"},
			wantErr: "collision",
		},
		{
			name:    "directory_file_collision",
			files:   map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a, b]\n", "a/nested/value.proto": "a", "b/nested": "b"},
			wantErr: "collision",
		},
		{
			name:    "retained_empty_directory_collision",
			files:   map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a, b]\n", "a/file.proto": "a", "b/a": "b"},
			wantErr: "collision",
		},
		{
			name:    "native_manifest",
			files:   map[string]string{"protobuf.mod": "module example.com/protos\n", "proto/file.proto": "message"},
			wantErr: "native protobuf.mod",
		},
		{
			name:    "nested_native_manifest",
			files:   map[string]string{"nested/protobuf.mod": "module example.com/protos/nested\n", "nested/file.proto": "message"},
			wantErr: "native protobuf.mod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, _ := migrationTestRepository(t, tt.files)
			files, err := trackedV1Files(t.Context(), repository)
			require.NoError(t, err)
			hash, err := hashMigrationLegacyFiles(repository, files)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, migrationTestHash(t, tt.want), hash)
		})
	}
}

func TestReadMigrationLegacyRootsRejectsAmbiguity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{name: "second_document", filename: "easyp.yaml", content: "generate: {}\n---\ngenerate: {}\n"},
		{name: "alias", filename: "easyp.yaml", content: "generate:\n  inputs: [{directory: &root proto}, {directory: *root}]\n"},
		{name: "null_root", filename: "easyp.yaml", content: "generate:\n  inputs: [{directory: {path: proto, root: null}}]\n"},
		{name: "unknown_root_field", filename: "easyp.yaml", content: "generate:\n  inputs: [{directory: {path: proto, roots: [proto]}}]\n"},
		{name: "placeholder", filename: "easyp.yaml", content: "generate:\n  inputs: [{directory: {path: proto, root: '${ROOT}'}}]\n"},
		{name: "duplicate_key", filename: "easyp.yaml", content: "generate: {}\ngenerate: {}\n"},
		{name: "null_document", filename: "easyp.yaml", content: "null\n"},
		{name: "input_without_source", filename: "easyp.yaml", content: "generate:\n  inputs: [{}]\n"},
		{name: "unsafe_root", filename: "buf.work.yaml", content: "version: v1\ndirectories: [../proto]\n"},
		{name: "buf_null_root", filename: "buf.work.yaml", content: "version: v1\ndirectories: [null]\n"},
		{name: "buf_missing_directories", filename: "buf.work.yaml", content: "version: v1\ndirectory: proto\n"},
		{name: "buf_v2_missing_path", filename: "buf.yaml", content: "version: v2\nmodules: [{name: example.com/protos}]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, _ := migrationTestRepository(t, map[string]string{tt.filename: tt.content})
			_, err := hashMigrationLegacyFiles(repository, []string{tt.filename})
			require.Error(t, err)
		})
	}
}

func TestFetchMigrationRejectsUnsafeNodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		submodule bool
	}{
		{name: "symlink"},
		{name: "submodule", submodule: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := migrationTestRepository(t, map[string]string{"file.proto": "message"})
			if tt.submodule {
				runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "160000,"+commit+",unsafe")
			} else {
				require.NoError(t, os.Symlink("file.proto", filepath.Join(repository, "unsafe.proto")))
				runTestGit(t, repository, "add", "unsafe.proto")
			}
			runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "unsafe")
			runTestGit(t, repository, "tag", "v1.0.0")
			cacheDir := t.TempDir()
			_, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, "v1.0.0", "h1:untrusted")
			require.ErrorContains(t, err, "non-regular")
			entries, err := os.ReadDir(cacheDir)
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

func TestMigrationTrackedFilesRejectsMaterializedProtoSymlink(t *testing.T) {
	t.Parallel()

	repository, _ := migrationTestRepository(t, map[string]string{"link.proto": "target.proto"})
	blob := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD:link.proto"))
	// A checkout with core.symlinks=false materializes a Git symlink as a
	// regular file. The index mode must still prevent legacy verification.
	runTestGit(t, repository, "update-index", "--cacheinfo", "120000,"+blob+",link.proto")
	_, _, err := migrationTrackedFiles(t.Context(), repository)
	require.ErrorContains(t, err, "non-regular Git mode")
}

func TestFetchMigrationPinnedCommitIgnoresMovedTag(t *testing.T) {
	t.Parallel()

	files := map[string]string{"old.proto": "old"}
	repository, commit := migrationTestRepository(t, files)
	runTestGit(t, repository, "tag", "v1.0.0")
	require.NoError(t, os.WriteFile(filepath.Join(repository, "new.proto"), []byte("new"), 0o600))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "new")
	runTestGit(t, repository, "tag", "-f", "v1.0.0")
	cacheDir := t.TempDir()
	fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, commit, migrationTestHash(t, files))
	require.NoError(t, err)
	assert.Equal(t, commit, fetched.Lock.Commit)
	assert.Equal(t, migrationTestHash(t, files), fetched.Lock.Hash)
}

func TestFetchMigrationRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	_, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), "unused", "main", "h1:untrusted")
	require.ErrorContains(t, err, "full Git commit or SemVer tag")
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestFetchMigrationRejectsLegacyHashWithoutPin(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	_, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), "unused", "", "h1:untrusted")
	require.ErrorContains(t, err, "full Git commit or SemVer tag")
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestFetchMigrationRejectsChangedSourceSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name:  "proto_outside_roots",
			files: map[string]string{"easyp.yaml": "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n", "proto/a.proto": "a", "extras/b.proto": "b"},
		},
		{
			name:  "overlapping_roots_add_import_alias",
			files: map[string]string{"buf.work.yaml": "version: v1\ndirectories: [a, a/b]\n", "a/b/file.proto": "file"},
		},
		{
			name:  "buf_v1beta1_roots_change_import_names",
			files: map[string]string{"buf.yaml": "version: v1beta1\nbuild: {roots: [proto]}\n", "proto/file.proto": "file"},
		},
		{
			name:  "hidden_directory_is_no_longer_selected",
			files: map[string]string{"file.proto": "file", ".hidden/extra.proto": "extra"},
		},
		{
			name:  "vendored_directory_is_no_longer_selected",
			files: map[string]string{"file.proto": "file", "easyp_vendor/extra.proto": "extra"},
		},
		{
			name:  "nested_buf_module_is_no_longer_selected",
			files: map[string]string{"file.proto": "file", "nested/buf.yaml": "version: v1\n", "nested/extra.proto": "extra"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range []struct {
				name    string
				withPin bool
			}{{name: "initial_resolution"}, {name: "historical_pin", withPin: true}} {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()
					repository, commit := migrationTestRepository(t, tt.files)
					files, err := trackedV1Files(t.Context(), repository)
					require.NoError(t, err)
					oldHash, err := hashMigrationLegacyFiles(repository, files)
					require.NoError(t, err)
					version := ""
					if mode.withPin {
						version = commit
					} else {
						oldHash = ""
					}
					cacheDir := t.TempDir()
					_, err = (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, version, oldHash)
					require.ErrorContains(t, err, "source selection")
					require.ErrorContains(t, err, "manual migration")
					entries, err := os.ReadDir(cacheDir)
					require.NoError(t, err)
					for _, entry := range entries {
						assert.Equal(t, "objects", entry.Name(), "temporary checkout was left behind")
					}
				})
			}
		})
	}
}

func TestFetchMigrationAcceptsNativeInitialResolution(t *testing.T) {
	t.Parallel()

	repository, _ := migrationTestRepository(t, map[string]string{"proto/file.proto": "file", "extra.proto": "extra"})
	require.NoError(t, os.WriteFile(filepath.Join(repository, "protobuf.mod"), []byte("module "+repository+"\nroots proto\n"), 0o600))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "native")
	cacheDir := t.TempDir()
	fetched, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, "", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"proto"}, fetched.Module.Roots)
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestFetchMigrationRejectsLegacyDependencyReplacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest string
		withPin  bool
	}{
		{name: "root_initial", manifest: "protobuf.mod"},
		{name: "root_pinned", manifest: "protobuf.mod", withPin: true},
		{name: "nested_initial", manifest: "nested/protobuf.mod"},
		{name: "nested_pinned", manifest: "nested/protobuf.mod", withPin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"file.proto": "file",
				tt.manifest:  "direct (\nexample.com/common@v1.0.0\n)\nreplace example.com/common@v1.0.0 => ../common\n",
			}
			repository, commit := migrationTestRepository(t, files)
			var version, oldHash string
			if tt.withPin {
				version, oldHash = commit, migrationTestHash(t, files)
			}
			cacheDir := t.TempDir()
			_, err := (&Cache{root: cacheDir}).FetchMigration(t.Context(), repository, version, oldHash)
			require.ErrorContains(t, err, "local replacements")
			require.ErrorContains(t, err, tt.manifest)
			require.ErrorContains(t, err, "manual migration")
			entries, err := os.ReadDir(cacheDir)
			require.NoError(t, err)
			for _, entry := range entries {
				assert.Equal(t, "objects", entry.Name(), "temporary checkout was left behind")
			}
		})
	}
}

func migrationTestRepository(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	directory := t.TempDir()
	migrationTestWriteFiles(t, directory, files)
	runTestGit(t, directory, "init", "-q")
	runTestGit(t, directory, "add", ".")
	runTestGit(t, directory, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	return directory, strings.TrimSpace(runTestGit(t, directory, "rev-parse", "HEAD"))
}

func migrationTestHash(t *testing.T, files map[string]string) string {
	t.Helper()
	directory := t.TempDir()
	migrationTestWriteFiles(t, directory, files)
	hash, err := dirhash.HashDir(directory, "", dirhash.DefaultHash)
	require.NoError(t, err)
	return hash
}

func migrationTestWriteFiles(t *testing.T, directory string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(directory, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
}
