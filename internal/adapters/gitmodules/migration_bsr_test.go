package gitmodules

import (
	"context"
	"strings"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/stretchr/testify/require"
)

type historicalTestResolver struct{ binding v1.BSRResolution }

func (r historicalTestResolver) Resolve(_ context.Context, dependency v1.BSRDependency) (v1.BSRResolution, error) {
	binding := r.binding
	binding.Dependency = dependency
	return binding, nil
}

func TestMigrationBSRPrefersHistoricalPinsOnlyForCompatibilitySnapshots(t *testing.T) {
	t.Parallel()
	const provider = "github.com/googleapis/googleapis"
	historical, fallback := strings.Repeat("8", 40), strings.Repeat("0", 40)
	for _, tc := range []struct{ name, resolution, want string }{
		{name: "compatibility snapshot", resolution: v1.BSRCompatibilitySnapshot, want: historical},
		{name: "exact BSR revision", resolution: v1.BSRExactRevision, want: fallback},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dependency := v1.BSRDependency{Module: "buf.build/googleapis/googleapis", Config: "buf.yaml", Commit: strings.Repeat("a", 32)}
			base := historicalTestResolver{binding: v1.BSRResolution{Git: v1.Requirement{Module: provider, Version: fallback}, Resolution: tc.resolution}}
			cache := NewWithBSRResolver(t.TempDir(), base)
			pins := []v1.Requirement{{Module: provider, Version: historical}}
			scoped := cache.WithMigrationPins(pins).(*Cache)
			pins[0].Version = strings.Repeat("f", 40)
			// An explicit Git edge must survive even when it equals the fallback BSR edge.
			module := v1.Module{Name: "github.com/grpc-ecosystem/grpc-gateway", Requires: []v1.Requirement{{Module: provider, Version: fallback}}, BSRDependencies: []v1.BSRDependency{dependency}}
			got, bindings, err := scoped.resolveBSRDependencies(t.Context(), module)
			require.NoError(t, err)
			require.Equal(t, tc.want, bindings[0].Git.Version)
			require.Equal(t, dependency, bindings[0].Dependency)
			require.Equal(t, tc.resolution, bindings[0].Resolution)
			require.Contains(t, got.Requires, v1.Requirement{Module: provider, Version: fallback})
			require.Contains(t, got.Requires, v1.Requirement{Module: provider, Version: tc.want})
			_, original, err := cache.resolveBSRDependencies(t.Context(), module)
			require.NoError(t, err)
			require.Equal(t, fallback, original[0].Git.Version, "migration must not alter normal resolution")
		})
	}
}
