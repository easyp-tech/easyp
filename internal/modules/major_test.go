package modules

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type majorSource struct {
	result Fetched
	calls  int
}

func (s *majorSource) Fetch(context.Context, string, string) (Fetched, error) {
	s.calls++
	return s.result, nil
}

func TestMajorFetchedIdentity(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, requested, declared, version string }{
		{"wrong_identity", "example.com/repo/v2", "example.com/repo", "v2.0.0"},
		{"head_identity", "example.com/repo/v2", "example.com/repo", ""},
		{"commit_identity", "example.com/repo/v2", "example.com/repo", strings.Repeat("a", 40)},
		{"missing_identity", "example.com/repo/v2", "", "v2.0.0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := &majorSource{result: Fetched{Module: v1.Module{Name: tt.declared}, Lock: v1.LockedModule{Source: tt.requested, Version: tt.version, Commit: strings.Repeat("a", 40)}}}
			_, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: tt.requested, Version: tt.version}}}, source, nil)
			require.ErrorContains(t, err, "identity")
		})
	}
}

func TestMajorRequirementsBeforeFetch(t *testing.T) {
	t.Parallel()
	source := &majorSource{}
	_, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "example.com/repo", Version: "v2.0.0"}}}, source, nil)
	require.Error(t, err)
	require.Zero(t, source.calls)
	err = ValidateRequirements([]v1.Requirement{{Module: "example.com/repo", Version: "v2.0.0"}}, v1.Lock{Version: 1, Modules: []v1.LockedModule{{Source: "example.com/repo", Version: "v2.1.0"}}})
	require.Error(t, err)
}
