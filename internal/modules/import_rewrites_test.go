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
		fetched:        map[[2]string]Fetched{{name, entry.Commit}: next},
	}

	bindings, err := source.tidyImportBindings(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{old}}, v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}})

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
		names: map[string]RootProtoFile{"svc.proto": file, "alias.proto": file},
	}

	name, err := binding.replacement("consumer.proto", "svc.proto")

	require.NoError(t, err)
	assert.Equal(t, "svc.proto", name)
}
