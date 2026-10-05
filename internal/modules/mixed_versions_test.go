package modules

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMixedRequirementKindsAreOrderIndependent(t *testing.T) {
	t.Parallel()
	a, b := strings.Repeat("a", 40), strings.Repeat("b", 40)
	for _, tc := range []struct {
		name     string
		weak     bool
		pin      string
		conflict bool
	}{
		{name: "tag and same commit", pin: a}, {name: "tag and different commit", pin: b, conflict: true}, {name: "tag and versionless", weak: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var first *v1.Lock
			for _, reverse := range []bool{false, true} {
				requires := []v1.Requirement{{Module: "a", Version: "v1.0.0"}, {Module: "bridge", Version: "v1.0.0"}}
				if reverse {
					slices.Reverse(requires)
				}
				version := tc.pin
				source := &fakeSource{revisions: map[string]Fetched{
					"a@v1.0.0":      {Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: a}},
					"a@" + a:        {Lock: v1.LockedModule{Source: "a", Version: a, Commit: a}},
					"a@" + b:        {Lock: v1.LockedModule{Source: "a", Version: b, Commit: b}},
					"bridge@v1.0.0": {Module: v1.Module{Requires: []v1.Requirement{{Module: "a", Version: version}}}, Lock: v1.LockedModule{Source: "bridge", Version: "v1.0.0", Commit: b}},
				}}
				got, err := Resolve(t.Context(), v1.Module{Requires: requires}, source, nil)
				if tc.conflict {
					require.ErrorContains(t, err, "conflicting requirements")
					continue
				}
				require.NoError(t, err)
				require.Equal(t, "v1.0.0", got.Modules[0].Version)
				if first != nil {
					require.Equal(t, *first, got)
				} else {
					first = &got
				}
				require.NotContains(t, source.calls, "a@")
			}
		})
	}
}

func TestSupersededHeadDoesNotLeakDependencies(t *testing.T) {
	t.Parallel()
	a, b := strings.Repeat("a", 40), strings.Repeat("b", 40)
	source := &fakeSource{revisions: map[string]Fetched{
		"a@":              {Module: v1.Module{Requires: []v1.Requirement{{Module: "obsolete", Version: "v1.0.0"}}}, Lock: v1.LockedModule{Source: "a", Version: a, Commit: a}},
		"a@v1.0.0":        {Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: b}},
		"bridge@":         {Module: v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}}}, Lock: v1.LockedModule{Source: "bridge", Version: b, Commit: b}},
		"obsolete@v1.0.0": {Lock: v1.LockedModule{Source: "obsolete", Version: "v1.0.0", Commit: a}},
	}}
	got, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "a"}, {Module: "bridge"}}}, source, nil)
	require.NoError(t, err)
	require.Len(t, got.Modules, 2)
	require.Equal(t, "a", got.Modules[0].Source)
	require.Equal(t, "v1.0.0", got.Modules[0].Version)
}

func TestUnavailableDependencyOfSupersededHeadIsIgnored(t *testing.T) {
	t.Parallel()
	a, b := strings.Repeat("a", 40), strings.Repeat("b", 40)
	source := &fakeSource{revisions: map[string]Fetched{
		"a@":       {Module: v1.Module{Requires: []v1.Requirement{{Module: "unavailable", Version: "v1.0.0"}}}, Lock: v1.LockedModule{Source: "a", Version: a, Commit: a}},
		"a@v1.0.0": {Lock: v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: b}},
		"bridge@":  {Module: v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}}}, Lock: v1.LockedModule{Source: "bridge", Version: b, Commit: b}},
	}}
	got, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "a"}, {Module: "bridge"}}}, source, nil)
	require.NoError(t, err)
	require.Len(t, got.Modules, 2)
}
