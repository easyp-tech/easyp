package gitmodules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	policyresolver "github.com/easyp-tech/easyp/internal/policy"
)

func TestFreshSnapshotMaterializesLogicalAliases(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	files := map[string]string{
		"meta/module":        "module " + repository + "\nroots proto\n",
		"sources/file.proto": "syntax = \"proto3\";\npackage alias;\n",
	}
	for name, data := range files {
		file := filepath.Join(repository, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
		require.NoError(t, os.WriteFile(file, []byte(data), 0o644))
	}
	require.NoError(t, os.Symlink("meta/module", filepath.Join(repository, "protobuf.mod")))
	require.NoError(t, os.Symlink("sources", filepath.Join(repository, "proto")))
	require.NoError(t, os.Symlink("missing", filepath.Join(repository, "auxiliary")))
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	cache := &Cache{root: t.TempDir()}
	source := repository
	fetched, err := cache.Fetch(t.Context(), source, commit)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(fetched.Lock.Hash, "h1:"), fetched.Lock.Hash)
	assert.Equal(t, []string{"proto"}, fetched.Module.Roots)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	installed, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(installed, "proto/file.proto"))
	require.NoError(t, err)
	assert.Equal(t, files["sources/file.proto"], string(data))
	info, err := os.Lstat(filepath.Join(installed, "protobuf.mod"))
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
	assert.Equal(t, v1CacheSourceKey(commit+":"+fetched.Lock.Hash), filepath.Base(installed))
}

func TestFreshRegularSnapshotUsesH1(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repository, "file.proto"), []byte("syntax = \"proto3\";\n"), 0o644))
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	fetched, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(fetched.Lock.Hash, "h1:"), fetched.Lock.Hash)
}

func TestSnapshotIgnoresCheckoutTransformations(t *testing.T) {
	repository := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repository, "file.proto"), []byte("syntax = \"proto3\";\n"), 0o644))
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	global := filepath.Join(t.TempDir(), "gitconfig")
	require.NoError(t, os.WriteFile(global, []byte("[core]\n autocrlf = true\n"), 0o600))
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	cache := &Cache{root: t.TempDir()}
	fresh, err := cache.Fetch(t.Context(), repository, commit)
	require.NoError(t, err)
	assert.Equal(t, migrationTestHash(t, map[string]string{"file.proto": "syntax = \"proto3\";\n"}), fresh.Lock.Hash)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fresh.Lock}}))
}

func TestSnapshotSelectedAliasesAreBounded(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, link, target, metadata, wantErr string }{
		{name: "internal file", link: "proto/alias.proto", target: "../data/file.proto"},
		{name: "component order", link: "proto/alias.proto", target: "../jump/../file.proto"},
		{name: "selected missing", link: "proto/alias.proto", target: "missing", wantErr: "file does not exist"},
		{name: "selected external", link: "proto/alias.proto", target: "../../external", wantErr: "outside root"},
		{name: "selected absolute", link: "proto/alias.proto", target: "/tmp/external", wantErr: "outside root"},
		{name: "selected file cycle", link: "proto/alias.proto", target: "alias.proto", wantErr: "cycle"},
		{name: "root missing", link: "alias", target: "missing", metadata: "roots alias\n", wantErr: "file does not exist"},
		{name: "root external", link: "alias", target: "../external", metadata: "roots alias\n", wantErr: "outside root"},
		{name: "root cycle", link: "alias", target: ".", metadata: "roots alias\n", wantErr: "cycle"},
		{name: "outside roots omitted", link: "alias.proto", target: "missing"},
		{name: "hidden metadata omitted", link: ".hidden/protobuf.mod", target: "missing"},
		{name: "unused Buf metadata omitted", link: "buf.yaml", target: "missing"},
		{name: "unused Buf metadata beneath selected root omitted", link: "buf.yaml", target: "missing", metadata: "roots .\n"},
		{name: "excluded omitted", link: "proto/private/alias.proto", target: "missing", metadata: "buf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			manifest := "module " + repository + "\nroots proto\n"
			if tt.metadata != "" && tt.metadata != "buf" {
				manifest = "module " + repository + "\n" + tt.metadata
			}
			data := map[string]string{"protobuf.mod": manifest, "data/file.proto": "target", "data/deeper/marker": "marker", "file.proto": "wrong", "proto/regular.proto": "regular"}
			if tt.metadata == "buf" {
				delete(data, "protobuf.mod")
				data["buf.yaml"] = "version: v1\nbuild:\n  roots: [proto]\n  excludes: [proto/private]\n"
			}
			for name, contents := range data {
				file := filepath.Join(repository, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
				require.NoError(t, os.WriteFile(file, []byte(contents), 0o644))
			}
			require.NoError(t, os.Symlink("data/deeper", filepath.Join(repository, "jump")))
			link := filepath.Join(repository, tt.link)
			require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o755))
			require.NoError(t, os.Symlink(tt.target, link))
			runTestGit(t, repository, "init", "-q")
			runTestGit(t, repository, "add", ".")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
			fetched, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, "")
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(fetched.Lock.Hash, "h1:"))
		})
	}
}

func TestSnapshotGitlinkBoundaries(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, root, alias string
		fail              bool
	}{
		{name: "opaque descendant", root: "proto"},
		{name: "declared root", root: "proto/submodule", fail: true},
		{name: "selected alias", root: "proto", alias: "submodule/file.proto", fail: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(repository, "protobuf.mod"), []byte("module "+repository+"\nroots "+tt.root+"\n"), 0o644))
			require.NoError(t, os.MkdirAll(filepath.Join(repository, "proto"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(repository, "proto/file.proto"), []byte("file"), 0o644))
			if tt.alias != "" {
				require.NoError(t, os.Symlink(tt.alias, filepath.Join(repository, "proto/alias.proto")))
			}
			runTestGit(t, repository, "init", "-q")
			runTestGit(t, repository, "add", ".")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
			commit := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
			runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "160000", commit, "proto/submodule")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "gitlink")
			fetched, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, "")
			if tt.fail {
				require.ErrorIs(t, err, gitsnapshot.ErrGitlink)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, fetched.Lock.Hash)
		})
	}
}

func TestSnapshotBufWorkspaceUsesLogicalDirectoryAlias(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	files := map[string]string{
		"buf.work.yaml":       "version: v1\ndirectories:\n  - alias\n",
		"physical/buf.yaml":   "version: v1\n",
		"physical/file.proto": "syntax = \"proto3\";\n",
	}
	for name, data := range files {
		file := filepath.Join(repository, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
		require.NoError(t, os.WriteFile(file, []byte(data), 0o644))
	}
	require.NoError(t, os.Symlink("physical", filepath.Join(repository, "alias")))
	require.NoError(t, os.Symlink("missing", filepath.Join(repository, "physical/aaa_auxiliary")))
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	cache := &Cache{root: t.TempDir()}
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"alias"}, fetched.Module.Roots)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	installed, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	bytes, err := os.ReadFile(filepath.Join(installed, "alias/file.proto"))
	require.NoError(t, err)
	assert.Equal(t, files["physical/file.proto"], string(bytes))
}

func TestSnapshotInstallLeavesObsoleteDirectoriesUntouched(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repository, "file.proto"), []byte("contents"), 0o644))
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	cache := &Cache{root: t.TempDir()}
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	obsolete := filepath.Join(cache.root, "modules", v1CacheSourceKey(repository), fetched.Lock.Commit)
	require.NoError(t, os.MkdirAll(obsolete, 0o755))
	sentinel := filepath.Join(obsolete, "foreign")
	require.NoError(t, os.WriteFile(sentinel, []byte("keep"), 0o600))
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	installed, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	assert.NotEqual(t, obsolete, installed)
	data, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	assert.Equal(t, "keep", string(data))
	other := fetched.Lock
	other.Hash = migrationTestHash(t, map[string]string{"file.proto": "changed"})
	assert.NotEqual(t, installed, v1ModuleCachePath(cache.root, other))
	require.ErrorContains(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{other}}), "hash mismatch")
	require.NoError(t, cache.VerifyCached(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
}

func TestSnapshotNeverRunsCheckoutFilters(t *testing.T) {
	repository := t.TempDir()
	files := map[string]string{"file.proto": "contents", ".gitattributes": "*.proto filter=blocker\n"}
	for name, data := range files {
		require.NoError(t, os.WriteFile(filepath.Join(repository, name), []byte(data), 0o644))
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	global := filepath.Join(t.TempDir(), "gitconfig")
	require.NoError(t, os.WriteFile(global, []byte("[filter \"blocker\"]\n smudge = false\n required = true\n"), 0o600))
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	cache := &Cache{root: t.TempDir()}
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.Equal(t, migrationTestHash(t, files), fetched.Lock.Hash)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
}

func TestSnapshotHashIncludesMaterializedLogicalFileAlias(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	files := map[string]string{"protobuf.mod": "module " + repository + "\nroots proto\n", "physical/file.proto": "target bytes"}
	for name, data := range files {
		file := filepath.Join(repository, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
		require.NoError(t, os.WriteFile(file, []byte(data), 0o644))
	}
	require.NoError(t, os.MkdirAll(filepath.Join(repository, "proto"), 0o755))
	require.NoError(t, os.Symlink("../physical/file.proto", filepath.Join(repository, "proto/alias.proto")))
	require.NoError(t, os.Symlink("missing", filepath.Join(repository, "proto/aaa_auxiliary")))
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	cache := &Cache{root: t.TempDir()}
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	files["proto/alias.proto"] = files["physical/file.proto"]
	assert.Equal(t, migrationTestHash(t, files), fetched.Lock.Hash)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	installed, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	actual, err := hashSnapshotV1Files(installed)
	require.NoError(t, err)
	assert.Equal(t, fetched.Lock.Hash, actual)
}

func TestSnapshotNestedMetadataAliasFailurePreservesBoundary(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"proto/child/item.proto": "syntax = \"proto3\";\n"}, map[string]string{"proto/child/protobuf.mod": "missing"})
	_, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, "")
	require.ErrorContains(t, err, "proto/child/protobuf.mod")
}

func TestSnapshotCachedPolicyResolvesCustomAliases(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, reference, link, target string }{
		{name: "file", reference: "./base.yaml", link: "base.yaml", target: "policies/strict.yaml"},
		{name: "directory namespace", reference: "./config/strict.yaml", link: "config", target: "policies"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := []byte("version: v1\nlinters:\n  extends: " + tt.reference + "\n")
			repository := snapshotPolicyFixture(t, map[string]string{"easyp.yaml": string(raw), "policies/strict.yaml": "version: v1\nlinters:\n  default: MINIMAL\n"}, map[string]string{tt.link: tt.target, "unused.yaml": "/outside/policy.yaml"})
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), repository, "")
			require.NoError(t, err)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			installed, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			cfg, err := v1.ParsePolicyLiteral(strings.NewReader(string(raw)))
			require.NoError(t, err)
			presence, err := v1.ParsePolicyPresence(raw)
			require.NoError(t, err)
			resolved, err := policyresolver.NewResolver(installed, nil).ResolveLint(t.Context(), policyresolver.LintInput{PolicyPath: filepath.Join(installed, "easyp.yaml"), Policy: cfg, Presence: presence, ModuleDir: installed})
			require.NoError(t, err)
			require.Equal(t, "MINIMAL", resolved.Policy.Linters.Default)
			info, err := os.Lstat(filepath.Join(installed, filepath.FromSlash(strings.TrimPrefix(tt.reference, "./"))))
			require.NoError(t, err)
			require.True(t, info.Mode().IsRegular())
		})
	}
}

func TestSnapshotReferencedPolicyAliasMustResolve(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, target string }{
		{name: "missing", target: "missing"},
		{name: "external", target: "../outside.yaml"},
		{name: "cycle", target: "base.yaml"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := snapshotPolicyFixture(t, map[string]string{"easyp.yaml": "version: v1\nlinters:\n  extends: ./base.yaml\n"}, map[string]string{"base.yaml": tt.target})
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), repository, "")
			require.NoError(t, err)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			installed, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			raw := []byte("version: v1\nlinters:\n  extends: ./base.yaml\n")
			policy, err := v1.ParsePolicyLiteral(strings.NewReader(string(raw)))
			require.NoError(t, err)
			_, err = policyresolver.NewResolver(installed, nil).ResolveLint(t.Context(), policyresolver.LintInput{PolicyPath: filepath.Join(installed, v1.PolicyFile), Policy: policy})
			require.Error(t, err)
		})
	}
}

func snapshotPolicyFixture(t *testing.T, files, links map[string]string) string {
	t.Helper()
	repository := t.TempDir()
	files[v1.ModuleFile] = "module " + repository + "\nroots proto\n"
	files["proto/main.proto"] = "syntax = \"proto3\";\n"
	for name, data := range files {
		target := filepath.Join(repository, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
		require.NoError(t, os.WriteFile(target, []byte(data), 0o644))
	}
	for name, target := range links {
		link := filepath.Join(repository, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o755))
		require.NoError(t, os.Symlink(target, link))
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	return repository
}

func TestSnapshotRemoteCustomPolicyAliasesWithoutRootPolicy(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, fragment, link, target string }{
		{name: "custom extension", fragment: "base.rules", link: "base.rules", target: "policies/strict.yaml"},
		{name: "directory namespace", fragment: "config/strict.yaml", link: "config", target: "policies"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := snapshotPolicyFixture(t, map[string]string{"policies/strict.yaml": "version: v1\nlinters:\n  default: MINIMAL\n"}, map[string]string{tt.link: tt.target, "unused.yaml": "missing"})
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), repository, "")
			require.NoError(t, err)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			installed, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			consumer := t.TempDir()
			raw := []byte("version: v1\nlinters:\n  extends: " + repository + "#" + tt.fragment + "\n")
			require.NoError(t, os.WriteFile(filepath.Join(consumer, v1.PolicyFile), raw, 0o644))
			cfg, err := v1.ParsePolicyLiteral(strings.NewReader(string(raw)))
			require.NoError(t, err)
			presence, err := v1.ParsePolicyPresence(raw)
			require.NoError(t, err)
			resolver := policyresolver.NewResolver(consumer, func(context.Context, string) (modules.PolicyGraph, error) {
				return modules.PolicyGraph{repository: {Name: repository, Directory: installed}}, nil
			})
			resolved, err := resolver.ResolveLint(t.Context(), policyresolver.LintInput{PolicyPath: filepath.Join(consumer, v1.PolicyFile), Policy: cfg, Presence: presence, ModuleDir: consumer})
			require.NoError(t, err)
			require.Equal(t, "MINIMAL", resolved.Policy.Linters.Default)
		})
	}
}

func TestSnapshotUnusedNestedFailureHonorsAncestorModule(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"proto/nested/protobuf.mod": "module example.com/nested\nroots .\n", "proto/nested/deep/item.proto": "syntax = \"proto3\";\n"}, map[string]string{"proto/nested/deep/protobuf.mod": "missing"})
	_, err := (&Cache{root: t.TempDir()}).Fetch(t.Context(), repository, "")
	require.NoError(t, err)
}
