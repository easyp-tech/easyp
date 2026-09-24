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

func TestResolve(t *testing.T) {
	t.Parallel()
	const commitA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const commitB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	fetchErr := errors.New("module unavailable")
	tests := []struct {
		name        string
		requires    []v1.Requirement
		revisions   map[string]Fetched
		pins        map[string]v1.LockedModule
		fetchErrors map[string]error
		canceled    bool
		want        []v1.LockedModule
		wantCalls   []string
		wantError   error
		wantMessage string
	}{
		{name: "empty graph", want: []v1.LockedModule{}},
		{
			name:     "highest minimum with deterministic module order",
			requires: []v1.Requirement{{Module: "b", Version: "v1.0.0"}, {Module: "a", Version: "v1.0.0"}},
			revisions: map[string]Fetched{
				"b@v1.0.0": {Module: v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.2.0"}}}, Lock: v1.LockedModule{Source: "b", Version: "v1.0.0", Commit: commitB}},
				"a@v1.0.0": {Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: commitA}},
				"a@v1.2.0": {Lock: v1.LockedModule{Source: "a", Version: "v1.2.0", Commit: commitB}},
			},
			want:      []v1.LockedModule{{Source: "a", Version: "v1.2.0", Commit: commitB}, {Source: "b", Version: "v1.0.0", Commit: commitB}},
			wantCalls: []string{"b@v1.0.0", "a@v1.0.0", "a@v1.2.0"},
		},
		{
			name:      "repeated cyclic requirement loads a revision once",
			requires:  []v1.Requirement{{Module: "a", Version: "v1.0.0"}, {Module: "a", Version: "v1.0.0"}},
			revisions: map[string]Fetched{"a@v1.0.0": {Module: v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}}}, Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: commitA}}},
			want:      []v1.LockedModule{{Source: "a", Version: "v1.0.0", Commit: commitA}}, wantCalls: []string{"a@v1.0.0"},
		},
		{
			name:      "versionless dependency resolves head once",
			requires:  []v1.Requirement{{Module: "a"}, {Module: "a"}},
			revisions: map[string]Fetched{"a@": {Lock: v1.LockedModule{Source: "a", Version: commitA, Commit: commitA}}},
			want:      []v1.LockedModule{{Source: "a", Version: commitA, Commit: commitA}}, wantCalls: []string{"a@"},
		},
		{
			name:      "versionless requirement uses existing commit pin",
			requires:  []v1.Requirement{{Module: "a"}},
			pins:      map[string]v1.LockedModule{"a": {Source: "a", Commit: commitB}},
			revisions: map[string]Fetched{"a@" + commitB: {Lock: v1.LockedModule{Source: "a", Version: commitB, Commit: commitB}}},
			want:      []v1.LockedModule{{Source: "a", Version: commitB, Commit: commitB}}, wantCalls: []string{"a@" + commitB},
		},
		{
			name:      "explicit version ignores previous pin",
			requires:  []v1.Requirement{{Module: "a", Version: "v1.0.0"}},
			pins:      map[string]v1.LockedModule{"a": {Source: "a", Commit: commitB}},
			revisions: map[string]Fetched{"a@v1.0.0": {Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: commitA}}},
			want:      []v1.LockedModule{{Source: "a", Version: "v1.0.0", Commit: commitA}}, wantCalls: []string{"a@v1.0.0"},
		},
		{
			name:      "different commits conflict before fetching the second",
			requires:  []v1.Requirement{{Module: "a", Version: commitA}, {Module: "a", Version: commitB}},
			revisions: map[string]Fetched{"a@" + commitA: {Lock: v1.LockedModule{Source: "a", Version: commitA, Commit: commitA}}},
			wantCalls: []string{"a@" + commitA}, wantMessage: "has conflicting requirements",
		},
		{
			name:      "tag and commit conflict",
			requires:  []v1.Requirement{{Module: "a", Version: "v1.0.0"}, {Module: "a", Version: commitA}},
			revisions: map[string]Fetched{"a@v1.0.0": {Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: commitA}}},
			wantCalls: []string{"a@v1.0.0"}, wantMessage: "has conflicting requirements",
		},
		{name: "invalid version", requires: []v1.Requirement{{Module: "a", Version: "branch"}}, wantMessage: "expected semantic version or full Git commit"},
		{name: "fetch error keeps its cause", requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}}, fetchErrors: map[string]error{"a@v1.0.0": fetchErr}, wantCalls: []string{"a@v1.0.0"}, wantError: fetchErr},
		{name: "head error keeps its cause", requires: []v1.Requirement{{Module: "a"}}, fetchErrors: map[string]error{"a@": fetchErr}, wantCalls: []string{"a@"}, wantError: fetchErr},
		{name: "canceled before fetching", requires: []v1.Requirement{{Module: "a"}}, canceled: true, wantError: context.Canceled},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := &fakeSource{revisions: tt.revisions, errors: tt.fetchErrors}
			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)
			if tt.canceled {
				cancel()
			}
			root := v1.Module{Requires: tt.requires}

			lock, err := Resolve(ctx, root, source, tt.pins)

			assert.Equal(t, tt.wantCalls, source.calls)
			if tt.wantMessage != "" {
				require.ErrorContains(t, err, tt.wantMessage)
				assert.Empty(t, lock)
				return
			}
			require.ErrorIs(t, err, tt.wantError)
			if err != nil {
				assert.Empty(t, lock)
				return
			}
			assert.Equal(t, 1, lock.Version)
			assert.Equal(t, tt.want, lock.Modules)
		})
	}
}

type fakeSource struct {
	revisions map[string]Fetched
	errors    map[string]error
	calls     []string
}

func (s *fakeSource) Fetch(_ context.Context, module, version string) (Fetched, error) {
	key := module + "@" + version
	s.calls = append(s.calls, key)
	if err := s.errors[key]; err != nil {
		return Fetched{}, err
	}
	revision, ok := s.revisions[key]
	if !ok {
		return Fetched{}, fmt.Errorf("unexpected module revision %s", key)
	}
	return revision, nil
}
