package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestSnapshotRejectsHostFilenameCollisions(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	upper := filepath.Join(repository, "Case.proto")
	require.NoError(t, os.WriteFile(upper, []byte("UPPER"), 0o644))
	_, caseErr := os.Stat(filepath.Join(repository, "case.proto"))
	caseInsensitive := caseErr == nil
	runTestGit(t, repository, "init", "-q")
	first := strings.TrimSpace(runTestGit(t, repository, "hash-object", "-w", "--", "Case.proto"))
	other := filepath.Join(repository, "other")
	require.NoError(t, os.WriteFile(other, []byte("LOWER"), 0o644))
	second := strings.TrimSpace(runTestGit(t, repository, "hash-object", "-w", "--", "other"))
	runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "100644,"+first+",Case.proto")
	runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "100644,"+second+",case.proto")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "case-sensitive Git names")

	fetched, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, "")

	if caseInsensitive {
		require.ErrorContains(t, err, "snapshot path collision")
		assert.Empty(t, fetched.Lock.Source)
	} else {
		require.NoError(t, err)
		assert.NotEmpty(t, fetched.Lock.Hash)
	}
}

func TestSnapshotCollisionsRespectRetainedPaths(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, upper, lower, excluded string }{
		{name: "retained directory collision", upper: "Case/a.proto", lower: "case/b.proto"},
		{name: "excluded file", upper: "Case.proto", lower: "case.proto", excluded: "Case.proto"},
		{name: "excluded namespace", upper: "Case/a.proto", lower: "case/b.proto", excluded: "Case"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(repository, "probe"), []byte("probe"), 0o644))
			_, caseErr := os.Stat(filepath.Join(repository, "Probe"))
			caseInsensitive := caseErr == nil
			runTestGit(t, repository, "init", "-q")
			for _, file := range []struct{ name, data string }{{name: tt.upper, data: "UPPER"}, {name: tt.lower, data: "LOWER"}} {
				require.NoError(t, os.WriteFile(filepath.Join(repository, "blob"), []byte(file.data), 0o644))
				hash := strings.TrimSpace(runTestGit(t, repository, "hash-object", "-w", "--", "blob"))
				runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "100644,"+hash+","+file.name)
			}
			if tt.excluded != "" {
				require.NoError(t, os.WriteFile(filepath.Join(repository, "buf.yaml"), []byte("version: v1\nbuild:\n  excludes: ["+tt.excluded+"]\n"), 0o644))
				runTestGit(t, repository, "add", "buf.yaml")
			}
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "retained logical paths")
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), repository, "")
			if caseInsensitive && tt.excluded == "" {
				require.ErrorContains(t, err, "snapshot path collision")
				return
			}
			require.NoError(t, err)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			installed, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			data, err := os.ReadFile(filepath.Join(installed, tt.lower))
			require.NoError(t, err)
			assert.Equal(t, "LOWER", string(data))
			if tt.excluded != "" {
				names, err := snapshotV1Files(installed)
				require.NoError(t, err)
				assert.NotContains(t, names, tt.upper)
				assert.Contains(t, names, tt.lower)
			}
		})
	}
}

func TestSnapshotIgnoredMetadataDoesNotRetainCollisionNamespace(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name             string
		retainedMetadata bool
	}{
		{name: "without retained metadata"},
		{name: "with retained metadata", retainedMetadata: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(repository, "protobuf.mod"), []byte("module "+repository+"\nroots case\n"), 0o644))
			require.NoError(t, os.MkdirAll(filepath.Join(repository, "case"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(repository, "case/file.proto"), []byte("LOWER"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(repository, "pointer"), []byte("missing"), 0o644))
			runTestGit(t, repository, "init", "-q")
			runTestGit(t, repository, "add", "protobuf.mod", "case/file.proto")
			if tt.retainedMetadata {
				require.NoError(t, os.WriteFile(filepath.Join(repository, "case/buf.yaml"), []byte("version: v1\n"), 0o644))
				runTestGit(t, repository, "add", "case/buf.yaml")
			}
			hash := strings.TrimSpace(runTestGit(t, repository, "hash-object", "-w", "--", "pointer"))
			runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "120000,"+hash+",Case/buf.yaml")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "ignored failure namespace")
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), repository, "")
			require.NoError(t, err)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			installed, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			data, err := os.ReadFile(filepath.Join(installed, "case/file.proto"))
			require.NoError(t, err)
			assert.Equal(t, "LOWER", string(data))
			names, err := snapshotV1Files(installed)
			require.NoError(t, err)
			assert.NotContains(t, names, "Case/buf.yaml")
			if tt.retainedMetadata {
				assert.Contains(t, names, "case/buf.yaml")
			}
		})
	}
}

func TestSnapshotUnusedAlternativeMajorMetadataDoesNotOverrideNative(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, alias string
		required    bool
	}{
		{name: "unused Buf", alias: "buf.yaml"},
		{name: "unused EasyP", alias: "easyp.yaml"},
		{name: "required native candidate", alias: "protobuf.mod", required: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			source := repository + "/v2"
			require.NoError(t, os.WriteFile(filepath.Join(repository, "protobuf.mod"), []byte("module "+source+"\nroots proto\n"), 0o644))
			for name, content := range map[string]string{"proto/file.proto": "SOURCE", "V2/buf.yaml": "version: v1\n", "pointer": "missing"} {
				filename := filepath.Join(repository, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
				require.NoError(t, os.WriteFile(filename, []byte(content), 0o644))
			}
			runTestGit(t, repository, "init", "-q")
			runTestGit(t, repository, "add", "protobuf.mod", "proto/file.proto", "V2/buf.yaml")
			hash := strings.TrimSpace(runTestGit(t, repository, "hash-object", "-w", "--", "pointer"))
			runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "120000,"+hash+",v2/"+tt.alias)
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "native precedence over unused major metadata")
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), source, "")
			if tt.required {
				require.ErrorIs(t, err, os.ErrNotExist)
				assert.Empty(t, fetched.Lock.Source)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, source, fetched.Module.Name)
			assert.Equal(t, []string{"proto"}, fetched.Module.Roots)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			installed, module, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			assert.Equal(t, source, module.Name)
			names, err := snapshotV1Files(installed)
			require.NoError(t, err)
			assert.Contains(t, names, "V2/buf.yaml")
			assert.NotContains(t, names, "v2/"+tt.alias)
		})
	}
}
