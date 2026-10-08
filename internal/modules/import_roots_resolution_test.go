package modules

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestImportRootSourceRejectsChangedCommitImmediately(t *testing.T) {
	t.Parallel()
	const requested = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const actual = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	source := &importRootSource{Source: revisionTestSource{entry: v1.LockedModule{Source: "example.test/dep", Version: requested, Commit: actual}}}

	_, err := source.Fetch(t.Context(), "example.test/dep", requested)

	require.ErrorContains(t, err, "requested commit")
	assert.Empty(t, source.fetched)
}

func TestLockedVersionGuardDefersOnlyMarkedEmptyHashes(t *testing.T) {
	t.Parallel()
	old := v1.LockedModule{Source: "example.test/dep", Version: "v1.0.0", Commit: "old", Hash: "hash-old"}
	tests := []struct {
		name        string
		commit      string
		hash        string
		inspection  *RootInspection
		wantChanged bool
	}{
		{name: "marked_incomplete_hash", commit: "old", inspection: &RootInspection{Provisional: true}},
		{name: "marked_retagged", commit: "new", inspection: &RootInspection{Provisional: true}, wantChanged: true},
		{name: "unmarked_empty_hash", commit: "old", wantChanged: true},
		{name: "unmarked_inspection", commit: "old", inspection: &RootInspection{}, wantChanged: true},
		{name: "marked_nonempty_hash", commit: "old", hash: "hash-new", inspection: &RootInspection{Provisional: true}, wantChanged: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fetched := Fetched{Lock: v1.LockedModule{Source: old.Source, Version: old.Version, Commit: tt.commit, Hash: tt.hash}, Inspection: tt.inspection}

			err := checkLockedVersion(old, fetched)

			if tt.wantChanged {
				require.ErrorIs(t, err, ErrLockedVersionChanged)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestFinalizationRejectsProvisionalResults(t *testing.T) {
	t.Parallel()
	const name = "example.test/dep"
	entry := v1.LockedModule{Source: name, Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash}
	module := v1.Module{Name: name, Roots: []string{"."}}
	fetched := Fetched{Module: module, Lock: entry, Inspection: &RootInspection{Provisional: true}}
	source := &importRootSource{Source: provisionalRootTestSource{fetched: fetched}, fetched: map[rootRevisionKey]Fetched{rootRevision(name, entry.Commit): fetched}}
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}

	_, err := source.finalizeSelections(t.Context(), lock, []importRootModule{{name: name, roots: module.Roots, inspection: fetched.Inspection}}, [][]string{nil})

	require.ErrorContains(t, err, "provisional")
}

func TestImportRootSourceChecksTagBeforeRetainedRootFetch(t *testing.T) {
	t.Parallel()
	old := v1.LockedModule{Source: "example.test/dep", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash, Roots: []string{"api"}}
	fetched := Fetched{Module: v1.Module{Name: old.Source, Roots: []string{"."}}, Lock: v1.LockedModule{Source: old.Source, Version: old.Version, Commit: versionlessCommitANew}, Inspection: &RootInspection{Provisional: true}}
	backend := &retaggedRootTestSource{provisionalRootTestSource: provisionalRootTestSource{fetched: fetched}}
	source := &importRootSource{Source: backend, locked: map[string]v1.LockedModule{old.Source: old}}

	_, err := source.Fetch(t.Context(), old.Source, old.Version)

	require.ErrorIs(t, err, ErrLockedVersionChanged)
	assert.Zero(t, backend.selections)
}

func TestCheckedRootReplacementKeepsIntegrityGuards(t *testing.T) {
	t.Parallel()
	old := v1.LockedModule{Source: "example.test/dep", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash, Roots: []string{"sources"}}
	previous := Fetched{Module: v1.Module{Name: old.Source, Roots: old.Roots}, Lock: old}
	tests := []struct {
		name        string
		roots       []string
		commit      string
		inspection  *RootInspection
		wantChanged bool
		wantError   string
	}{
		{name: "equivalent_scope_changed_hash", roots: []string{"sources"}, commit: old.Commit, wantChanged: true},
		{name: "retagged_changed_scope", roots: []string{"api"}, commit: versionlessCommitANew, wantChanged: true},
		{name: "provisional_changed_scope", roots: []string{"api"}, commit: old.Commit, inspection: &RootInspection{Provisional: true}, wantError: "provisional"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			current := Fetched{Module: v1.Module{Name: old.Source, Roots: tt.roots}, Lock: v1.LockedModule{Source: old.Source, Version: old.Version, Commit: tt.commit, Hash: "h1:changed"}, Inspection: tt.inspection}
			source := &importRootSource{Source: provisionalRootTestSource{fetched: previous}, locked: map[string]v1.LockedModule{old.Source: old}, hints: map[string][]string{old.Source: tt.roots}}

			err := source.verifyRootFetch(t.Context(), current)

			if tt.wantChanged {
				require.ErrorIs(t, err, ErrLockedVersionChanged)
			} else {
				require.ErrorContains(t, err, tt.wantError)
			}
		})
	}
}

type retaggedRootTestSource struct {
	provisionalRootTestSource
	selections int
}

func (source *retaggedRootTestSource) FetchForRootResolution(context.Context, string, string) (Fetched, error) {
	return source.fetched, nil
}

func (source *retaggedRootTestSource) FetchWithRoots(context.Context, string, string, []string) (Fetched, error) {
	source.selections++
	return Fetched{}, errors.New("retained root is missing")
}

type provisionalRootTestSource struct{ fetched Fetched }

func TestImportRootSearchIsBounded(t *testing.T) {
	t.Parallel()
	dependency := &RootInspection{Files: []RootProtoFile{
		{Path: "anchor.proto", Identity: "anchor.proto"},
		{Path: "request.proto", Identity: "request.proto", Content: []byte(`import "anchor.proto";`)},
	}}
	for number := range 10 {
		name := fmt.Sprintf("v1/a_%02d.proto", number)
		declaration := fmt.Sprintf("choices/%02d/inner/request_%02d.proto", number, number)
		dependency.Files = append(dependency.Files, RootProtoFile{Path: declaration, Identity: declaration, Content: []byte(fmt.Sprintf("syntax = %q; import %q;", "proto3", name))})
		for _, side := range []string{"", "inner/"} {
			path := fmt.Sprintf("choices/%02d/%s%s", number, side, name)
			dependency.Files = append(dependency.Files, RootProtoFile{Path: path, Identity: path, Content: []byte(`syntax = "proto3";`)})
			// Every candidate preserves this default binding. Exhausting
			// the search must still fail rather than retain the default.
			anchor := fmt.Sprintf("choices/%02d/%sanchor.proto", number, side)
			dependency.Files = append(dependency.Files, RootProtoFile{Path: anchor, Identity: "anchor.proto"})
			collision := fmt.Sprintf("choices/%02d/%scollision.proto", number, side)
			dependency.Files = append(dependency.Files, RootProtoFile{Path: collision, Identity: collision, Content: []byte(`syntax = "proto3";`)})
		}
	}

	_, err := selectImportRoots(t.Context(), []importRootModule{
		{name: "dependency", roots: []string{"."}, inspection: dependency},
	})

	require.ErrorContains(t, err, "exceeds 1024 layouts")
	assert.ErrorContains(t, err, "explicit import roots")
}

func (source provisionalRootTestSource) Fetch(context.Context, string, string) (Fetched, error) {
	return source.fetched, nil
}

func (source provisionalRootTestSource) FetchWithRoots(context.Context, string, string, []string) (Fetched, error) {
	return source.fetched, nil
}
