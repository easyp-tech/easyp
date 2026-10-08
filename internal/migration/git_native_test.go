package migration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationGitHintsMatchNativeRootAuthority(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, metadataRoots, extraFile, wantError string
	}{
		{name: "matching authority", metadataRoots: "roots api\n"},
		{name: "implicit native default is authority", wantError: "authoritative roots"},
		{name: "hidden legacy target is not silently removed", metadataRoots: "roots api\n", extraFile: "api/.hidden/extra.proto", wantError: ".hidden/extra.proto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n"}
			if tt.extraFile != "" {
				files[tt.extraFile] = "syntax = \"proto3\"; package hidden.v1; message Extra {}\n"
			}
			repository, _ := gitSelectionRepository(t, files)
			writeFixture(t, repository, v1.ModuleFile, "module "+repository+"\n"+tt.metadataRoots)
			gitSelectionCommand(t, repository, "add", ".")
			gitSelectionCommand(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.test", "commit", "-qm", "native metadata")
			project := t.TempDir()
			legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "', root: api}}]\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
				return
			}
			require.NoError(t, err)
			gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
			require.NoError(t, err)
			assert.Equal(t, []v1.GenerateModule{{Module: repository}}, gen.Generate.Modules)
			var lock v1.Lock
			require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.LockFile), &lock))
			require.Len(t, lock.Modules, 1)
			assert.Empty(t, lock.Modules[0].Roots, "authoritative roots need no fallback lock metadata")
		})
	}
}

func TestMigrationDefaultNativeGitInputCannotOmitHiddenLegacyTargets(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"public/a.proto":      "syntax = \"proto3\"; package public.v1; message A {}\n",
		".hidden/extra.proto": "syntax = \"proto3\"; package hidden.v1; message Extra {}\n",
	}
	repository, _ := gitSelectionRepository(t, files)
	writeFixture(t, repository, v1.ModuleFile, "module "+repository+"\n")
	gitSelectionCommand(t, repository, "add", ".")
	gitSelectionCommand(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.test", "commit", "-qm", "native metadata")
	commit := strings.TrimSpace(gitSelectionCommand(t, repository, "rev-parse", "HEAD"))
	project := t.TempDir()
	legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "@" + commit + "'}}]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	require.NoError(t, os.Chmod(filepath.Join(project, v1.PolicyFile), 0o640))
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.ErrorContains(t, err, ".hidden/extra.proto")
	assert.Nil(t, plan)
	assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
	info, err := os.Stat(filepath.Join(project, v1.PolicyFile))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	entries, err := os.ReadDir(project)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "a refused default selection creates no native candidates or backups")
}

func TestMigrationDefaultNativeGitInputKeepsBufArchiveNames(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"buf.yaml":            "version: v2\nmodules: [{path: api}]\n",
		"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
	}
	repository, _ := gitSelectionRepository(t, files)
	writeFixture(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	gitSelectionCommand(t, repository, "add", ".")
	gitSelectionCommand(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.test", "commit", "-qm", "native metadata")
	commit := strings.TrimSpace(gitSelectionCommand(t, repository, "rev-parse", "HEAD"))
	project := t.TempDir()
	legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "@" + commit + "'}}]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	cache := gitmodules.New(t.TempDir())
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: cache})
	require.NoError(t, err)
	gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
	require.NoError(t, err)
	assert.Equal(t, []v1.GenerateModule{{Module: repository}}, gen.Generate.Modules)
	var lock v1.Lock
	require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.LockFile), &lock))
	require.Len(t, lock.Modules, 1)
	assert.Equal(t, commit, lock.Modules[0].Commit)
	assert.Empty(t, lock.Modules[0].Roots)
	require.NoError(t, cache.Install(t.Context(), lock))
	directory, module, err := cache.Cached(lock.Modules[0])
	require.NoError(t, err)
	names, contents := gitSelectionTargets(t, directory, module, nil)
	assert.Equal(t, map[string]string{"service/a.proto": "api/service/a.proto"}, names)
	assert.Equal(t, map[string]string{"service/a.proto": files["api/service/a.proto"]}, contents)
	require.NoError(t, plan.Apply())
	assert.Equal(t, legacy, string(mustRead(t, project, "easyp.yaml.v0.bak")))
}

func TestMigrationNativeAuthorityKeepsProducerArchiveRootRewriting(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, producerFile, producerConfig string
		otherMetadata                      map[string]string
	}{
		{name: "Buf module roots", producerFile: "buf.yaml", producerConfig: "version: v2\nmodules: [{path: api}]\n"},
		{name: "Buf workspace roots", producerFile: "buf.work.yaml", producerConfig: "version: v1\ndirectories: [api]\n"},
		{name: "EasyP producer roots", producerFile: "easyp.yaml", producerConfig: "generate:\n  inputs: [{directory: {path: ., root: api}}]\n"},
		{name: "workspace root priority", producerFile: "buf.work.yaml", producerConfig: "version: v1\ndirectories: [api]\n", otherMetadata: map[string]string{"buf.yaml": "version: v2\nmodules: [{path: other}]\n", "easyp.yaml": "generate:\n  inputs: [{directory: {path: ., root: third}}]\n"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				tt.producerFile:       tt.producerConfig,
				"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
			}
			for name, content := range tt.otherMetadata {
				files[name] = content
			}
			repository, _ := gitSelectionRepository(t, files)
			writeFixture(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
			gitSelectionCommand(t, repository, "add", ".")
			gitSelectionCommand(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.test", "commit", "-qm", "native metadata")
			commit := strings.TrimSpace(gitSelectionCommand(t, repository, "rev-parse", "HEAD"))
			for _, selector := range []struct{ name, root, subdir, wantError string }{
				{name: "old installed subdirectory", subdir: "service"},
				{name: "repository path is not an installed selector", root: "api", subdir: "api/service", wantError: "api/service"},
			} {
				t.Run(selector.name, func(t *testing.T) {
					t.Parallel()
					project := t.TempDir()
					legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "@" + commit + "', root: '" + selector.root + "', sub_directory: '" + selector.subdir + "'}}]\n"
					writeFixture(t, project, v1.PolicyFile, legacy)
					plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
					if selector.wantError != "" {
						require.ErrorContains(t, err, selector.wantError)
						assert.Nil(t, plan)
						assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
						for _, name := range []string{v1.ModuleFile, v1.LockFile, v1.GenerateFile, "easyp.yaml.v0.bak"} {
							assert.NoFileExists(t, filepath.Join(project, name))
						}
						return
					}
					require.NoError(t, err)
					gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
					require.NoError(t, err)
					assert.Equal(t, []v1.GenerateModule{{Module: repository, Paths: []string{"api/service"}}}, gen.Generate.Modules)
					var lock v1.Lock
					require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.LockFile), &lock))
					require.Len(t, lock.Modules, 1)
					assert.Equal(t, commit, lock.Modules[0].Commit)
					assert.Empty(t, lock.Modules[0].Roots, "native roots retain authority over the effective v1 namespace")
				})
			}
		})
	}
}
