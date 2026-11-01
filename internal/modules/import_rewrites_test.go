package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestTidyBindingsRejectUnpinnedSourceIdentities(t *testing.T) {
	t.Parallel()
	const name = "example.test/dep"
	old := v1.LockedModule{Source: name, Version: "v0.4.0", Commit: versionlessCommitA, Hash: versionlessHash, Roots: []string{"api/v1"}}
	entry := v1.LockedModule{Source: name, Version: "v0.5.0", Commit: versionlessCommitANew, Hash: versionlessHash}
	previous := Fetched{Module: v1.Module{Name: name, Roots: old.Roots}, Lock: old, Inspection: &RootInspection{Files: []RootProtoFile{{Path: "api/v1/svc.proto", Identity: "unverified-same-source"}}}}
	next := Fetched{Module: v1.Module{Name: name, Roots: []string{"api"}, RootsFromMetadata: true}, Lock: entry, Inspection: &RootInspection{Files: []RootProtoFile{{Path: "api/v1/svc.proto", Identity: "unverified-same-source"}}}}
	source := &importRootSource{Source: provisionalRootTestSource{},
		verifiedScopes: map[string]Fetched{name: previous},
		fetched:        map[rootRevisionKey]Fetched{rootRevision(name, entry.Commit): next},
	}
	resolved, err := source.checkedResult(v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}})
	require.NoError(t, err)
	planner := newTidyImportPlanner(resolved, v1.Lock{Version: 1, Modules: []v1.LockedModule{old}}, source.Source)

	bindings, err := planner.tidyImportBindings(t.Context(), map[string]v1UnresolvedImport{"svc.proto": {owner: "consumer.proto", path: "svc.proto"}}, nil)

	require.ErrorContains(t, err, "pinned source identity")
	assert.Empty(t, bindings)
}

func TestTidyBindingKeepsOriginalNameForAdditionalSourceAliases(t *testing.T) {
	t.Parallel()
	old := v1.LockedModule{Source: "example.test/dep", Commit: versionlessCommitA}
	current := v1.LockedModule{Source: old.Source, Commit: versionlessCommitANew}
	previous := RootProtoFile{Path: "api/svc.proto", Identity: old.Source + "@" + old.Commit + ":physical/svc.proto"}
	file := RootProtoFile{Path: "api/svc.proto", Identity: current.Source + "@" + current.Commit + ":physical/svc.proto"}
	binding := tidyImportBinding{before: Fetched{Lock: old}, after: Fetched{Lock: current}, file: previous,
		names: map[string][]RootProtoFile{"svc.proto": {file}, "alias.proto": {file}},
	}

	name, err := binding.replacement("consumer.proto", "svc.proto")

	require.NoError(t, err)
	assert.Equal(t, "svc.proto", name)
}

func TestTidyBindingsPreserveRootOrderForIdenticalHistoricalCopies(t *testing.T) {
	t.Parallel()
	const name = "example.test/dep"
	old := v1.LockedModule{Source: name, Version: "v0.4.0", Commit: versionlessCommitA, Hash: versionlessHash, Roots: []string{"schema", "api"}}
	entry := v1.LockedModule{Source: name, Version: "v0.5.0", Commit: versionlessCommitANew, Hash: versionlessHash, Roots: []string{"schema/v1", "api"}}
	files := func(commit string) []RootProtoFile {
		var result []RootProtoFile
		// Inspection order deliberately differs from root precedence.
		for _, path := range []string{"api/v1/svc.proto", "schema/v1/svc.proto"} {
			result = append(result, RootProtoFile{Path: path, Identity: name + "@" + commit + ":" + path, Content: []byte("syntax = \"proto3\"; message Service {}")})
		}
		return result
	}
	previous := Fetched{Module: v1.Module{Name: name, Roots: old.Roots}, Lock: old, Inspection: &RootInspection{Files: files(old.Commit)}}
	next := Fetched{Module: v1.Module{Name: name, Roots: entry.Roots}, Lock: entry, Inspection: &RootInspection{Files: files(entry.Commit)}}
	repository := &rootEvidenceTestRepository{fakeRepository: &fakeRepository{}, previous: previous, current: next}
	result := graphResolveResult{lock: v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}, revisions: map[rootRevisionKey]Fetched{rootRevision(name, entry.Commit): next}}
	planner := newTidyImportPlanner(result, v1.Lock{Version: 1, Modules: []v1.LockedModule{old}}, repository)

	bindings, err := planner.tidyImportBindings(t.Context(), map[string]v1UnresolvedImport{"v1/svc.proto": {owner: "consumer.proto", path: "v1/svc.proto"}}, nil)

	require.NoError(t, err)
	assert.Equal(t, "schema/v1/svc.proto", bindings["v1/svc.proto"].file.Path)
	after, err := bindings["v1/svc.proto"].replacement("consumer.proto", "v1/svc.proto")
	require.NoError(t, err)
	assert.Equal(t, "svc.proto", after)
}
