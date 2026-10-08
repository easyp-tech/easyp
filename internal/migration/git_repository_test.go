package migration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

type uncheckedRootsRepository struct{ *mockRepository }

func (r uncheckedRootsRepository) FetchMigrationWithRoots(ctx context.Context, source, version, hash string, _ []string) (modules.Fetched, error) {
	return r.FetchMigration(ctx, source, version, hash)
}

func TestMigrationRootsCapabilityMustHonorAndPersistCheckedHints(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, logical, returnedRoot, wantError string
		provisional                            bool
	}{
		{name: "ignored explicit hint", logical: "service/a.proto", returnedRoot: ".", wantError: "root"},
		{name: "unrecorded fallback", logical: "api/service/a.proto", returnedRoot: "api", wantError: "record"},
		{name: "provisional sources", logical: "api/service/a.proto", returnedRoot: "api", wantError: "provisional", provisional: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			const dependency = "example.test/dependency"
			project := t.TempDir()
			legacy := "generate:\n  inputs: [{git_repo: {url: " + dependency + ", root: api}}]\n"
			writeFixture(t, project, v1.PolicyFile, legacy)
			writeFixture(t, project, "easyp.lock", dependency+" "+testCommit+" "+testHash+"\n")
			fetched := migrationFetched(dependency, testCommit, testCommit)
			fetched.Module.Roots = []string{tt.returnedRoot}
			fetched.Inspection = &modules.RootInspection{Files: []modules.RootProtoFile{{Path: tt.logical, Identity: dependency + ":" + tt.logical, Content: []byte("syntax = \"proto3\"; package service.v1; message A {}\n")}}, LegacyFiles: map[string]string{tt.logical: tt.logical}}
			if tt.provisional {
				fetched.Inspection.Provisional = true
				fetched.Lock.Roots = []string{tt.returnedRoot}
			}
			repository := uncheckedRootsRepository{&mockRepository{wantOldHash: testHash, fetched: map[string]modules.Fetched{dependency + "@" + testCommit: fetched}}}
			_, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: repository})
			require.ErrorContains(t, err, tt.wantError)
			assert.Equal(t, legacy, string(mustRead(t, project, v1.PolicyFile)))
		})
	}
}

func TestMigrationExplicitGitSelectionRequiresRepositoryCapability(t *testing.T) {
	t.Parallel()
	project := t.TempDir()
	writeFixture(t, project, v1.PolicyFile, "generate:\n  inputs: [{git_repo: {url: example.test/dependency, sub_directory: service}}]\n")
	repository := &mockRepository{}
	_, err := Build(t.Context(), Options{Dir: project, Module: "example.test/consumer", ResolveLock: true, Repository: repository})
	require.ErrorContains(t, err, "roots-aware migration repository")
	assert.Empty(t, repository.calls)
}
