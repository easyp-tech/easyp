package modules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestTidyBindingsRetainCheckedEvidenceWhenAnotherSourceViewChanges(t *testing.T) {
	t.Parallel()
	const name = "example.test/dep"
	const contents = `syntax = "proto3"; message Service {}`
	old := v1.LockedModule{Source: name, Version: "v0.4.0", Commit: versionlessCommitA, Hash: versionlessHash, Roots: []string{"api/v1"}}
	entry := v1.LockedModule{Source: name, Version: "v0.5.0", Commit: versionlessCommitANew, Hash: versionlessHash, Roots: []string{"api"}}
	previous := Fetched{
		Module: v1.Module{Name: name, Roots: []string{"api/v1"}}, Lock: old,
		Inspection: &RootInspection{Files: []RootProtoFile{{Path: "api/v1/svc.proto", Identity: name + "@" + old.Commit + ":api/v1/svc.proto", Content: []byte(contents)}}},
	}
	next := Fetched{
		Module: v1.Module{Name: name, Roots: []string{"api"}, RootsFromMetadata: true}, Lock: entry,
		Inspection: &RootInspection{
			Files:       []RootProtoFile{{Path: "api/v1/svc.proto", Identity: name + "@" + entry.Commit + ":api/v1/svc.proto", Content: []byte(contents)}},
			LegacyFiles: map[string]string{"svc.proto": "api/v1/svc.proto"},
		},
	}
	repository := &rootEvidenceTestRepository{fakeRepository: &fakeRepository{}, previous: previous, current: next}
	before := v1.Lock{Version: 1, Modules: []v1.LockedModule{old}}
	resolved, err := resolveV1Graph(t.Context(), v1.Module{Name: "example.test/app", Requires: []v1.Requirement{{Module: name, Version: entry.Version}}}, before, repository, graphResolveRequest{transitions: namespaceTransitionsTidyRepairs})
	require.NoError(t, err)

	retained, exists := resolved.currentRevision(entry)
	require.True(t, exists)
	other, exists := resolved.currentRevision(entry)
	require.True(t, exists)
	other.Module.Roots[0] = "other"
	other.Lock.Roots[0] = "other"
	other.Inspection.Files[0].Content[0] = 'X'
	other.Inspection.Files[0].Identity = "unverified"
	other.Inspection.LegacyFiles["svc.proto"] = "other.proto"
	otherLock := resolved.lockFile()
	otherLock.Modules[0].Roots[0] = "other"
	otherLock.Modules[0].Hash = "changed"

	planner := newTidyImportPlanner(resolved, before, repository)
	bindings, err := planner.tidyImportBindings(t.Context(), map[string]v1UnresolvedImport{"svc.proto": {owner: "consumer.proto", path: "svc.proto"}}, nil)
	require.NoError(t, err)
	nameAfter, err := bindings["svc.proto"].replacement("consumer.proto", "svc.proto")
	require.NoError(t, err)
	assert.Equal(t, "v1/svc.proto", nameAfter)
	assert.Equal(t, entry, bindings["svc.proto"].after.Lock)
	assert.Equal(t, contents, string(retained.Inspection.Files[0].Content))
	assert.Equal(t, "api/v1/svc.proto", retained.Inspection.LegacyFiles["svc.proto"])
	current, exists := resolved.currentRevision(entry)
	require.True(t, exists)
	assert.Equal(t, retained, current)
	assert.Equal(t, entry, resolved.lockFile().Modules[0])
}

type rootEvidenceTestRepository struct {
	*fakeRepository
	previous Fetched
	current  Fetched
}

func (repository *rootEvidenceTestRepository) Fetch(_ context.Context, _, version string) (Fetched, error) {
	if version == repository.previous.Lock.Commit {
		return repository.previous, nil
	}
	return repository.current, nil
}

func (repository *rootEvidenceTestRepository) FetchForRootResolution(ctx context.Context, source, version string) (Fetched, error) {
	return repository.Fetch(ctx, source, version)
}

func (repository *rootEvidenceTestRepository) FetchWithRoots(ctx context.Context, source, version string, _ []string) (Fetched, error) {
	return repository.Fetch(ctx, source, version)
}
