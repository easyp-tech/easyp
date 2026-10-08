package migration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationGitProofFailuresLeaveConsumerFilesUnchanged(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, root, subdir, metadata, extraName, wantError string
		wrongHash, movedTag                                bool
	}{
		{name: "root-only must retain outside-root targets", root: "api", extraName: "extras/outside.proto", wantError: "extras/outside.proto"},
		{name: "producer roots are authoritative", root: "api/service", metadata: "generate:\n  inputs: [{directory: {path: ., root: api}}]\n", wantError: "authoritative roots"},
		{name: "malformed producer metadata", root: "api", metadata: "generate:\n  inputs: [{directory: {path: ., roots: [api]}}]\n", wantError: "roots"},
		{name: "empty subdirectory cannot widen targets", root: "api", subdir: "missing", wantError: "missing"},
		{name: "wrong historical hash", root: "api", wrongHash: true, wantError: "legacy hash mismatch"},
		{name: "moved historical tag", root: "api", movedTag: true, wantError: "legacy hash mismatch"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n"}
			if tt.metadata != "" {
				files[v1.PolicyFile] = tt.metadata
			}
			if tt.extraName != "" {
				files[tt.extraName] = "syntax = \"proto3\"; package extra.v1; message Extra {}\n"
			}
			repository, commit := gitSelectionRepository(t, files)
			hash := gitSelectionHash(t, files)
			if tt.name == "producer roots are authoritative" {
				hash = gitSelectionHash(t, map[string]string{"service/a.proto": files["api/service/a.proto"]})
			}
			if tt.wrongHash {
				hash = gitSelectionHash(t, map[string]string{"different.proto": "different"})
			}
			version := commit
			if tt.movedTag {
				version = "v0.4.0"
				writeFixture(t, repository, "api/service/a.proto", "syntax = \"proto3\"; package service.v1; message Changed {}\n")
				gitSelectionCommand(t, repository, "add", ".")
				gitSelectionCommand(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.test", "commit", "-qm", "new source")
				gitSelectionCommand(t, repository, "tag", "-f", "v0.4.0")
			}
			project := t.TempDir()
			legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "', root: '" + tt.root + "', sub_directory: '" + tt.subdir + "'}}]\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			oldLock := repository + " " + version + " " + hash + "\n"
			writeFixture(t, project, "easyp.lock", oldLock)
			_, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
			require.ErrorContains(t, err, tt.wantError)
			assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
			assert.Equal(t, oldLock, string(mustRead(t, project, "easyp.lock")))
			entries, err := os.ReadDir(project)
			require.NoError(t, err)
			assert.Len(t, entries, 2)
		})
	}
}

func TestMigrationGitLiteralValidationPrecedesRepositoryAccess(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, field, value string }{
		{name: "root traversal", field: "root", value: "api/../service"},
		{name: "absolute subdirectory", field: "sub_directory", value: "/api"},
		{name: "root glob", field: "root", value: "api/*"},
		{name: "subdirectory placeholder", field: "sub_directory", value: "${PROTO}"},
		{name: "root backslash", field: "root", value: `api\service`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			project := t.TempDir()
			legacy := "generate:\n  inputs: [{git_repo: {url: example.test/dependency, " + tt.field + ": '" + tt.value + "'}}]\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			repository := &mockRepository{}
			_, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: repository})
			require.ErrorContains(t, err, "git_repo."+tt.field)
			assert.Empty(t, repository.calls)
			assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
		})
	}
}

func TestMigrationGitCompatibleSubdirectoriesAreUnioned(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
		"api/other/b.proto":   "syntax = \"proto3\"; package other.v1; message B {}\n",
		"api/unused/c.proto":  "syntax = \"proto3\"; package unused.v1; message C {}\n",
	}
	repository, commit := gitSelectionRepository(t, files)
	project := t.TempDir()
	input := func(root, subdir string) string {
		return "{git_repo: {url: '" + repository + "', root: '" + root + "', sub_directory: '" + subdir + "'}}"
	}
	legacy := "generate:\n  inputs: [" + input("./api/", "api/service/") + ", " + input("api", "api/other") + ", " + input("api", "api/service") + "]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, files)+"\n")
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
	require.NoError(t, err)
	assert.Equal(t, []v1.GenerateModule{{Module: repository, Paths: []string{"api/other", "api/service"}}}, gen.Generate.Modules)
	writeFixture(t, project, v1.PolicyFile, strings.Replace(legacy, input("api", "api/other"), input("api/other", "api/other"), 1))
	_, err = Build(t.Context(), Options{Dir: project, Module: "example.test/consumer"})
	require.ErrorContains(t, err, "conflicts")
}

func TestMigrationGitProofRechecksReachableLocalSourcesBeforeApply(t *testing.T) {
	t.Parallel()
	files := map[string]string{"api/service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n"}
	repository, commit := gitSelectionRepository(t, files)
	project := t.TempDir()
	legacy := "generate:\n  inputs: [{directory: selected}, {git_repo: {url: '" + repository + "', root: api}}]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	writeFixture(t, project, "selected/main.proto", "syntax = \"proto3\"; package selected.v1; import \"shared/import.proto\"; message Main { shared.v1.Import value = 1; }\n")
	writeFixture(t, project, "shared/import.proto", "syntax = \"proto3\"; package shared.v1; message Import {}\n")
	writeFixture(t, project, "easyp.lock", repository+" "+commit+" "+gitSelectionHash(t, files)+"\n")
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	writeFixture(t, project, "shared/import.proto", "syntax = \"proto3\"; package shared.v1; message Import { string changed = 1; }\n")
	require.ErrorContains(t, plan.CheckUnchanged(), "bindings changed")
	require.ErrorContains(t, plan.Apply(), "bindings changed")
	assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
	assert.NoFileExists(t, filepath.Join(project, "easyp.yaml.v0.bak"))
}

func TestMigrationProducerFilteredImportCannotFallThroughToAnotherDependency(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"buf.yaml":        "version: v1\nbuild: {excludes: [private]}\n",
		"service/a.proto": "syntax = \"proto3\"; package service.v1; import \"private/b.proto\"; message A { private.v1.B value = 1; }\n",
		"private/b.proto": "syntax = \"proto3\"; package private.v1; message B { string original = 1; }\n",
	}
	first, firstCommit := gitSelectionRepository(t, files)
	otherFiles := map[string]string{"private/b.proto": "syntax = \"proto3\"; package private.v1; message B { int32 replacement = 1; }\n"}
	second, secondCommit := gitSelectionRepository(t, otherFiles)
	project := t.TempDir()
	legacy := "generate:\n  inputs: [{git_repo: {url: '" + first + "', sub_directory: service}}]\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	oldLock := first + " " + firstCommit + " " + gitSelectionHash(t, files) + "\n" + second + " " + secondCommit + " " + gitSelectionHash(t, otherFiles) + "\n"
	writeFixture(t, project, "easyp.lock", oldLock)
	_, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.ErrorContains(t, err, "private/b.proto")
	assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
	assert.Equal(t, oldLock, string(mustRead(t, project, "easyp.lock")))
	assert.NoFileExists(t, filepath.Join(project, "easyp.yaml.v0.bak"))
}
