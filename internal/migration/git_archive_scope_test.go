package migration

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/generation"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestMigrationScopedArchiveAllowsUnselectedExportIgnoredSource(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		".gitattributes":  "unused/b.proto export-ignore\n",
		"service/a.proto": "syntax = \"proto3\"; package service.v1; message A {}\n",
		"unused/b.proto":  "syntax = \"proto3\"; package unused.v1; message B {}\n",
	}
	repository, commit := gitSelectionRepository(t, files)
	archived := migrationActualProtoArchive(t, repository, commit)
	require.Equal(t, map[string]string{"service/a.proto": files["service/a.proto"]}, archived)
	project := t.TempDir()
	legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "', sub_directory: service}}]\n"
	oldLock := repository + " " + commit + " " + gitSelectionHash(t, archived) + "\n"
	writeFixture(t, project, v1.PolicyFile, legacy)
	writeFixture(t, project, "easyp.lock", oldLock)
	plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
	require.NoError(t, err)
	require.NoError(t, plan.Apply())
	assert.Equal(t, oldLock, string(mustRead(t, project, "easyp.lock")))
	assert.Equal(t, legacy, string(mustRead(t, project, "easyp.yaml.v0.bak")))
	for _, name := range []string{"first.pb", "second.pb"} {
		cold := gitmodules.New(t.TempDir())
		err := generation.Run(t.Context(), logger.NewNop(), cold, generation.Request{WorkDir: project, WorkspaceRoot: project, Frozen: true, DescriptorSetOut: filepath.Join(project, name)})
		require.NoError(t, err)
		var descriptors descriptorpb.FileDescriptorSet
		require.NoError(t, proto.Unmarshal(mustRead(t, project, name), &descriptors))
		require.Len(t, descriptors.File, 1)
		assert.Equal(t, "service/a.proto", descriptors.File[0].GetName())
	}
	assert.Equal(t, mustRead(t, project, "first.pb"), mustRead(t, project, "second.pb"))
	for name, content := range files {
		assert.Equal(t, content, string(mustRead(t, repository, name)), "the verified provider tree is unchanged")
	}
}

func TestMigrationArchiveScopeKeepsTargetImportAndByteGuards(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, subdir, affected, body, want string }{
		{name: "whole input cannot add ignored target", affected: "unused/b.proto", body: "syntax = \"proto3\"; package service.v1; message A {}\n", want: "unused/b.proto"},
		{name: "selected subtree cannot add ignored target", subdir: "service", affected: "service/b.proto", body: "syntax = \"proto3\"; package service.v1; message A {}\n", want: "service/b.proto"},
		{name: "reachable ignored import cannot bind", subdir: "service", affected: "unused/b.proto", body: "syntax = \"proto3\"; package service.v1; import \"unused/b.proto\"; message A { unused.v1.B value = 1; }\n", want: "unused/b.proto"},
		{name: "selected substituted bytes cannot change", subdir: "service", affected: "service/a.proto", body: "// $Format:%H$\nsyntax = \"proto3\"; package service.v1; message A {}\n", want: "changes"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			attribute := tt.affected + " export-ignore\n"
			if tt.name == "selected substituted bytes cannot change" {
				attribute = tt.affected + " export-subst\n"
			}
			files := map[string]string{".gitattributes": attribute, "service/a.proto": tt.body}
			if tt.affected != "service/a.proto" {
				files[tt.affected] = "syntax = \"proto3\"; package unused.v1; message B {}\n"
			}
			repository, commit := gitSelectionRepository(t, files)
			archived := migrationActualProtoArchive(t, repository, commit)
			project := t.TempDir()
			legacy := "generate:\n  inputs: [{git_repo: {url: '" + repository + "', sub_directory: '" + tt.subdir + "'}}]\n"
			oldLock := repository + " " + commit + " " + gitSelectionHash(t, archived) + "\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			writeFixture(t, project, "easyp.lock", oldLock)
			plan, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: gitmodules.New(t.TempDir())})
			require.ErrorContains(t, err, tt.want)
			assert.Nil(t, plan)
			assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
			assert.Equal(t, oldLock, string(mustRead(t, project, "easyp.lock")))
			for _, name := range []string{v1.ModuleFile, v1.LockFile, v1.GenerateFile, "easyp.yaml.v0.bak"} {
				assert.NoFileExists(t, filepath.Join(project, name))
			}
			if tt.name == "whole input cannot add ignored target" {
				_, err = gitmodules.New(t.TempDir()).FetchMigration(t.Context(), repository, commit, gitSelectionHash(t, archived))
				require.ErrorContains(t, err, "unused/b.proto", "ordinary migration fetches still require whole namespace equivalence")
			}
		})
	}
}

func migrationActualProtoArchive(t *testing.T, repository, commit string) map[string]string {
	t.Helper()
	raw := []byte(gitSelectionCommand(t, repository, "archive", "--format=zip", commit, "--", "*.proto"))
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	require.NoError(t, err)
	files := make(map[string]string)
	for _, file := range archive.File {
		if file.Mode().IsDir() {
			continue
		}
		require.True(t, file.Mode().IsRegular())
		reader, err := file.Open()
		require.NoError(t, err)
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		require.NoError(t, readErr)
		require.NoError(t, closeErr)
		files[file.Name] = string(data)
	}
	return files
}
