package migration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestMigrationGitSelectionPreservesPinnedNamesAndSources(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, root, subdirectory string
		paths, roots             []string
		want                     map[string]string
	}{
		{name: "default", want: map[string]string{"api/service/a.proto": "api/service/a.proto", "api/other/b.proto": "api/other/b.proto"}},
		{name: "root_only", root: "api", roots: []string{"api"}, want: map[string]string{"service/a.proto": "api/service/a.proto", "other/b.proto": "api/other/b.proto"}},
		{name: "subdirectory_only", subdirectory: "api/service", paths: []string{"api/service"}, want: map[string]string{"api/service/a.proto": "api/service/a.proto"}},
		{name: "root_and_subdirectory", root: "api", subdirectory: "api/service", paths: []string{"api/service"}, roots: []string{"api"}, want: map[string]string{"service/a.proto": "api/service/a.proto"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
				"api/other/b.proto":   "syntax = \"proto3\"; package other.v1; message B {}\n",
			}
			repository, commit := gitSelectionRepository(t, files)
			project := t.TempDir()
			legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "@v0.4.0', root: '" + tt.root + "', sub_directory: '" + tt.subdirectory + "'}}]\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			oldLock := repository + " " + commit + " " + gitSelectionHash(t, files) + "\n"
			writeFixture(t, project, "easyp.lock", oldLock)
			cache := gitmodules.New(t.TempDir())
			preview, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", Repository: cache})
			require.NoError(t, err)
			require.True(t, preview.NeedsLockResolution())
			require.ErrorContains(t, preview.Apply(), "--resolve-lock")
			assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
			plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: cache})
			require.NoError(t, err)
			gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
			require.NoError(t, err)
			require.Equal(t, []v1.GenerateModule{{Module: repository, Paths: tt.paths}}, gen.Generate.Modules)
			assert.Empty(t, gen.Generate.Paths)
			assert.Empty(t, gen.Generate.Packages)
			manifest, err := v1.ParseModule(bytes.NewReader(outputContent(t, plan, v1.ModuleFile)))
			require.NoError(t, err)
			assert.Equal(t, []v1.Requirement{{Module: repository, Version: "v0.4.0"}}, manifest.Requires)
			var lock v1.Lock
			require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.LockFile), &lock))
			require.Len(t, lock.Modules, 1)
			assert.Equal(t, commit, lock.Modules[0].Commit)
			assert.Equal(t, "v0.4.0", lock.Modules[0].Version)
			assert.Equal(t, tt.roots, lock.Modules[0].Roots)
			require.NoError(t, cache.Install(t.Context(), lock))
			directory, module, err := cache.Cached(lock.Modules[0])
			require.NoError(t, err)
			names, contents := gitSelectionTargets(t, directory, module, tt.paths)
			assert.Equal(t, tt.want, names)
			for name, physical := range tt.want {
				assert.Equal(t, files[physical], contents[name])
			}
			require.NoError(t, plan.Apply())
			assert.Equal(t, oldLock, string(mustRead(t, project, "easyp.lock")))
			assert.Equal(t, legacy, string(mustRead(t, project, "easyp.yaml.v0.bak")))
		})
	}
}

func TestMigrationGitSelectorsKeepLocalAndRemoteScopeIndependent(t *testing.T) {
	t.Parallel()
	files := map[string]string{"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n", "api/other/b.proto": "syntax = \"proto3\"; package other.v1; message B {}\n"}
	repository, commit := gitSelectionRepository(t, files)
	project := t.TempDir()
	remote := "{git_repo: {url: '" + repository + "', root: api, sub_directory: api/service}}"
	legacy := "generate:\n  inputs: [{directory: mcp}, " + remote + ", " + remote + "]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	writeFixture(t, project, "mcp/options.proto", "syntax = \"proto3\"; package mcp.v1; message Options {}\n")
	writeFixture(t, project, "build/mcp/options.proto", "syntax = \"proto3\"; package mcp.v1; message Options {}\n")
	writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, files)+"\n")
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
	require.NoError(t, err)
	assert.Equal(t, []v1.GenerateModule{{Module: "example.test/consumer", Paths: []string{"mcp"}}, {Module: repository, Paths: []string{"api/service"}}}, gen.Generate.Modules)
	assert.Empty(t, gen.Generate.Paths)
	assert.Empty(t, gen.Generate.Packages)
}

func TestMigrationGitSubdirectoryTranslatesVerifiedArchivePaths(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"easyp.yaml":          "generate:\n  inputs: [{directory: {path: ., root: api}}]\n",
		"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
		"api/other/b.proto":   "syntax = \"proto3\"; package other.v1; message B {}\n",
	}
	repository, commit := gitSelectionRepository(t, files)
	project := t.TempDir()
	legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "', sub_directory: service}}]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	installed := map[string]string{"service/a.proto": files["api/service/a.proto"], "other/b.proto": files["api/other/b.proto"]}
	writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, installed)+"\n")
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
	require.NoError(t, err)
	assert.Equal(t, []v1.GenerateModule{{Module: repository, Paths: []string{"api/service"}}}, gen.Generate.Modules)
	var lock v1.Lock
	require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.LockFile), &lock))
	require.Len(t, lock.Modules, 1)
	assert.Empty(t, lock.Modules[0].Roots, "producer root metadata remains authoritative")
	writeFixture(t, project, v1.GenerateFile, string(outputContent(t, plan, v1.GenerateFile)))
	repeated, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err, "an existing generator that exactly matches the verified physical selectors is compatible")
	for _, output := range repeated.Outputs() {
		if output.Name == v1.GenerateFile {
			assert.True(t, output.Unchanged)
		}
	}
	conflicting := bytes.ReplaceAll(outputContent(t, plan, v1.GenerateFile), []byte("api/service"), []byte("api/other"))
	writeFixture(t, project, v1.GenerateFile, string(conflicting))
	_, err = Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.ErrorContains(t, err, "existing easyp.gen.yaml conflicts")
	assert.Equal(t, conflicting, mustRead(t, project, v1.GenerateFile))
}

func TestMigrationUnfilteredEmptyGitModuleRetainsItsScope(t *testing.T) {
	t.Parallel()
	files := map[string]string{"README": "this revision has no proto sources\n"}
	repository, commit := gitSelectionRepository(t, files)
	project := t.TempDir()
	writeFixture(t, project, v1.PolicyFile, "generate:\n  inputs: [{git_repo: {url: '"+repository+"'}}]\n")
	writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, files)+"\n")
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
	require.NoError(t, err)
	assert.Equal(t, []v1.GenerateModule{{Module: repository}}, gen.Generate.Modules)
}

func TestMigrationGitReachableImportsKeepSourceBindings(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, imported, localName, wantError string
	}{
		{name: "rooted reachable source", imported: "other/b.proto"},
		{name: "raw alias would disappear", imported: "api/other/b.proto", wantError: "api/other/b.proto"},
		{name: "local source shadows dependency", imported: "other/b.proto", localName: "other/b.proto", wantError: "duplicate import path"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"api/service/a.proto": "syntax = \"proto3\"; package service.v1; import \"" + tt.imported + "\"; message A { other.v1.B value = 1; }\n",
				"api/other/b.proto":   "syntax = \"proto3\"; package other.v1; message B { string remote = 1; }\n",
			}
			repository, commit := gitSelectionRepository(t, files)
			project := t.TempDir()
			inputs := "{git_repo: {url: '" + repository + "', root: api, sub_directory: api/service}}"
			if tt.localName != "" {
				inputs = "{directory: other}, " + inputs
				writeFixture(t, project, tt.localName, "syntax = \"proto3\"; package other.v1; message B { int32 local = 1; }\n")
			}
			legacy := "generate:\n  inputs: [" + inputs + "]\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			oldLock := repository + " " + commit + " " + gitSelectionHash(t, files) + "\n"
			writeFixture(t, project, "easyp.lock", oldLock)
			plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
			if tt.wantError == "" {
				require.NoError(t, err)
				require.NoError(t, plan.CheckUnchanged())
				return
			}
			require.ErrorContains(t, err, tt.wantError)
			assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
			assert.Equal(t, oldLock, string(mustRead(t, project, "easyp.lock")))
			for _, name := range []string{v1.ModuleFile, v1.LockFile, v1.GenerateFile, "easyp.yaml.v0.bak"} {
				assert.NoFileExists(t, filepath.Join(project, name))
			}
		})
	}
}

func TestMigrationGitImportRootPrecedesInstalledDefaultPaths(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"api/service/a.proto": "syntax = \"proto3\"; package service.v1; import \"other/b.proto\"; message A { other.v1.B value = 1; }\n",
		"api/other/b.proto":   "syntax = \"proto3\"; package other.v1; message B { string inside = 1; }\n",
		"other/b.proto":       "syntax = \"proto3\"; package other.v1; message B { int32 outside = 1; }\n",
	}
	repository, commit := gitSelectionRepository(t, files)
	project := t.TempDir()
	writeFixture(t, project, v1.PolicyFile, "generate:\n  inputs: [{git_repo: {url: '"+repository+"', root: api, sub_directory: api/service}}]\n")
	writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, files)+"\n")
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	binding := plan.git.bindings[repository+":other/b.proto"]
	assert.Equal(t, "api/other/b.proto", binding.path)
	assert.Equal(t, files["api/other/b.proto"], string(binding.content))
}

func TestMigrationGitSubdirectoryExcludesUnselectedProducerFilteredFiles(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		wholeTree bool
	}{
		{name: "released archive hash"},
		{name: "historical whole tree hash", wholeTree: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"buf.yaml":        "version: v1\nbuild: {excludes: [private]}\n",
				"service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
				"private/b.proto": "syntax = \"proto3\"; package private.v1; message B {}\n",
			}
			repository, commit := gitSelectionRepository(t, files)
			installed := map[string]string{"service/a.proto": files["service/a.proto"], "private/b.proto": files["private/b.proto"]}
			if tt.wholeTree {
				installed = files
			}
			project := t.TempDir()
			writeFixture(t, project, v1.PolicyFile, "generate:\n  inputs: [{git_repo: {url: '"+repository+"', sub_directory: service}}]\n")
			writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, installed)+"\n")
			cache := gitmodules.New(t.TempDir())
			plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: cache})
			require.NoError(t, err)
			gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
			require.NoError(t, err)
			assert.Equal(t, []v1.GenerateModule{{Module: repository, Paths: []string{"service"}}}, gen.Generate.Modules)
			var lock v1.Lock
			require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.LockFile), &lock))
			require.Len(t, lock.Modules, 1)
			require.NoError(t, cache.Install(t.Context(), lock))
			directory, module, err := cache.Cached(lock.Modules[0])
			require.NoError(t, err)
			names, contents := gitSelectionTargets(t, directory, module, []string{"service"})
			assert.Equal(t, map[string]string{"service/a.proto": "service/a.proto"}, names)
			assert.Equal(t, map[string]string{"service/a.proto": files["service/a.proto"]}, contents)
		})
	}
}

func gitSelectionRepository(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	directory := t.TempDir()
	for name, content := range files {
		writeFixture(t, directory, name, content)
	}
	gitSelectionCommand(t, directory, "init", "-q")
	gitSelectionCommand(t, directory, "add", ".")
	gitSelectionCommand(t, directory, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := strings.TrimSpace(gitSelectionCommand(t, directory, "rev-parse", "HEAD"))
	gitSelectionCommand(t, directory, "tag", "v0.4.0")
	return directory, commit
}

func gitSelectionCommand(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.CommandContext(t.Context(), "git", args...)
	command.Dir = directory
	out, err := command.CombinedOutput()
	require.NoError(t, err, "%s", out)
	return string(out)
}

func gitSelectionHash(t *testing.T, files map[string]string) string {
	t.Helper()
	directory := t.TempDir()
	for name, content := range files {
		writeFixture(t, directory, name, content)
	}
	hash, err := dirhash.HashDir(directory, "", dirhash.Hash1)
	require.NoError(t, err)
	return hash
}

func gitSelectionTargets(t *testing.T, directory string, module v1.Module, paths []string) (map[string]string, map[string]string) {
	t.Helper()
	roots, err := modules.ModuleSources(directory, module)
	require.NoError(t, err)
	names, contents := make(map[string]string), make(map[string]string)
	for _, root := range roots {
		err := root.Walk(func(file string) error {
			physical, err := filepath.Rel(directory, file)
			require.NoError(t, err)
			physical = filepath.ToSlash(physical)
			if len(paths) > 0 && !slices.ContainsFunc(paths, func(selector string) bool { return v1.PathSelectorMatches(selector, physical) }) {
				return nil
			}
			name, err := filepath.Rel(root.Path, file)
			require.NoError(t, err)
			raw, err := os.ReadFile(file)
			require.NoError(t, err)
			names[filepath.ToSlash(name)], contents[filepath.ToSlash(name)] = physical, string(raw)
			return nil
		})
		require.NoError(t, err)
	}
	return names, contents
}
