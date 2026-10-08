package migration

import (
	"bytes"
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
