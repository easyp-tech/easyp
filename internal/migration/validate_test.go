package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestBuildNativeBSRLockNoOp(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name        string
		resolveLock bool
	}{
		{name: "without_lock_resolution"},
		{name: "with_lock_resolution", resolveLock: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, files, infos := nativeLockProject(t, nativeBSRLockFixture())
			repository := &mockRepository{}

			plan, err := Build(t.Context(), Options{
				Dir: root, Module: "example.com/acme/api", ResolveLock: tt.resolveLock, Repository: repository,
			})

			require.NoError(t, err)
			assert.True(t, plan.AlreadyV1())
			assert.False(t, plan.NeedsLockResolution())
			assert.Empty(t, plan.Outputs())
			assert.Empty(t, repository.calls, "native validation must not access dependencies")
			assertNativeLockProjectUnchanged(t, root, files, infos)

			lock, err := validateNativeLock(mustRead(t, root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, v1.Lock{Version: 1, Modules: []v1.LockedModule{
				{
					Source: "example.com/acme/dep", Version: "v1.0.0", Commit: testCommit, Hash: testHash,
					BSR: []v1.BSRResolution{{
						Dependency: v1.BSRDependency{
							Module: "buf.build/googleapis/googleapis", Reference: "stable", Commit: strings.Repeat("b", 32),
							Digest: "b5:" + strings.Repeat("c", 64), Config: "proto/buf.yaml",
						},
						Git:        v1.Requirement{Module: "github.com/googleapis/googleapis", Version: "v1.0.0"},
						Resolution: v1.BSRCompatibilitySnapshot,
					}},
				},
				{Source: "github.com/googleapis/googleapis", Version: "v1.0.0", Commit: testCommit, Hash: testHash},
			}}, lock)

			require.NoError(t, plan.CheckUnchanged())
			require.NoError(t, plan.Apply())
			assert.Empty(t, repository.calls, "applying a native no-op must not refresh dependencies")
			assertNativeLockProjectUnchanged(t, root, files, infos)
		})
	}
}

func TestBuildNativeLockRejectsInvalidDocuments(t *testing.T) {
	t.Parallel()
	valid := nativeBSRLockFixture()
	for _, tt := range []struct {
		name    string
		lock    string
		wantErr string
	}{
		{name: "unknown_lock_field", lock: valid + "unsupported: true\n", wantErr: "field unsupported not found"},
		{name: "unknown_module_field", lock: strings.Replace(valid, "    bsr:", "    unsupported: true\n    bsr:", 1), wantErr: "field unsupported not found"},
		{name: "unknown_binding_field", lock: strings.Replace(valid, "        resolution:", "        unsupported: true\n        resolution:", 1), wantErr: "field unsupported not found"},
		{name: "unknown_dependency_field", lock: strings.Replace(valid, "          config:", "          unsupported: true\n          config:", 1), wantErr: "field unsupported not found"},
		{name: "unknown_git_field", lock: strings.Replace(valid, "        git: {module:", "        git: {unsupported: true, module:", 1), wantErr: "field unsupported not found"},
		{name: "malformed_yaml", lock: "version: 1\nmodules: [\n", wantErr: "did not find expected node content"},
		{name: "multiple_documents", lock: valid + "---\nversion: 1\nmodules: []\n", wantErr: "must contain one YAML document"},
		{name: "malformed_trailing_document", lock: valid + "---\nmodules: [\n", wantErr: "did not find expected node content"},
		{name: "invalid_lock_version", lock: strings.Replace(valid, "version: 1", "version: 2", 1), wantErr: "unsupported protobuf.lock version"},
		{name: "invalid_commit", lock: strings.Replace(valid, testCommit, "invalid-commit", 1), wantErr: "invalid commit"},
		{name: "invalid_hash", lock: strings.Replace(valid, testHash, "h1:invalid", 1), wantErr: "invalid content hash"},
		{name: "unpinned_binding", lock: strings.Replace(valid, "git: {module: github.com/googleapis/googleapis, version: v1.0.0}", "git: {module: github.com/googleapis/googleapis}", 1), wantErr: "unpinned BSR Git target"},
		{name: "invalid_binding_version", lock: strings.Replace(valid, "git: {module: github.com/googleapis/googleapis, version: v1.0.0}", "git: {module: github.com/googleapis/googleapis, version: main}", 1), wantErr: "invalid version"},
		{name: "invalid_binding_resolution", lock: strings.Replace(valid, "compatibility_snapshot", "implicit_latest", 1), wantErr: "invalid BSR resolution"},
		{name: "invalid_binding_origin", lock: strings.Replace(valid, "config: proto/buf.yaml", "config: ../buf.yaml", 1), wantErr: "invalid BSR config path"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, files, infos := nativeLockProject(t, tt.lock)
			repository := &mockRepository{}

			plan, err := Build(t.Context(), Options{
				Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repository,
			})

			require.ErrorContains(t, err, "checkNative: validateNativeLock:")
			require.ErrorContains(t, err, tt.wantErr)
			assert.Nil(t, plan)
			assert.Empty(t, repository.calls, "invalid native locks must not access dependencies")
			assertNativeLockProjectUnchanged(t, root, files, infos)
		})
	}
}

func nativeBSRLockFixture() string {
	return fmt.Sprintf(`version: 1
modules:
  - source: example.com/acme/dep
    version: v1.0.0
    commit: %s
    hash: %s
    bsr:
      - dependency:
          module: buf.build/googleapis/googleapis
          reference: stable
          commit: %s
          digest: b5:%s
          config: proto/buf.yaml
        git: {module: github.com/googleapis/googleapis, version: v1.0.0}
        resolution: compatibility_snapshot
  - source: github.com/googleapis/googleapis
    version: v1.0.0
    commit: %s
    hash: %s
`, testCommit, testHash, strings.Repeat("b", 32), strings.Repeat("c", 64), testCommit, testHash)
}

func nativeLockProject(t *testing.T, lock string) (string, map[string]string, map[string]os.FileInfo) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		v1.PolicyFile:   "# Keep native policy bytes.\nversion: v1\n",
		v1.GenerateFile: "# Keep native generation bytes.\nversion: v1\nplugins: []\n",
		v1.ModuleFile:   "// Keep native manifest bytes.\nmodule example.com/acme/api\nrequire example.com/acme/dep v1.0.0\n",
		v1.LockFile:     lock,
		"api.proto":     "syntax = \"proto3\";\npackage acme.api;\nmessage Request {}\n",
		"README.md":     "Unrelated project content must also stay unchanged.\n",
	}
	infos := make(map[string]os.FileInfo, len(files)+1)
	for name, content := range files {
		writeFixture(t, root, name, content)
		info, err := os.Stat(filepath.Join(root, name))
		require.NoError(t, err)
		infos[name] = info
	}
	info, err := os.Stat(root)
	require.NoError(t, err)
	infos["."] = info
	return root, files, infos
}

func assertNativeLockProjectUnchanged(t *testing.T, root string, files map[string]string, infos map[string]os.FileInfo) {
	t.Helper()
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	var actual, expected []string
	for _, entry := range entries {
		actual = append(actual, entry.Name())
	}
	for name, content := range files {
		expected = append(expected, name)
		assert.Equal(t, content, string(mustRead(t, root, name)), name)
	}
	assert.ElementsMatch(t, expected, actual, "native migration must not create backups or staging files")
	for name, before := range infos {
		after, err := os.Stat(filepath.Join(root, name))
		require.NoError(t, err)
		assert.True(t, os.SameFile(before, after), "%s was replaced", name)
		assert.Equal(t, before.Mode(), after.Mode(), "%s mode changed", name)
		assert.Equal(t, before.ModTime(), after.ModTime(), "%s was written", name)
	}
}
