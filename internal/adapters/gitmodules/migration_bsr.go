package gitmodules

import (
	"context"
	"fmt"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// WithMigrationPins creates an isolated migration repository whose approximate
// BSR snapshots prefer the caller's historical Git pins. Exact BSR resolutions
// remain immutable. The migration coordinator still verifies every pin's hash.
func (c *Cache) WithMigrationPins(pins []v1.Requirement) modules.MigrationRepository {
	historical := make(map[string]string, len(pins))
	for _, pin := range pins {
		historical[pin.Module] = pin.Version
	}
	scoped := *c
	scoped.bsrResolver = migrationBSRResolver{base: c.bsrResolver, pins: historical}
	return &scoped
}

type migrationBSRResolver struct {
	base modules.BSRResolver
	pins map[string]string
}

func (r migrationBSRResolver) Resolve(ctx context.Context, dependency v1.BSRDependency) (v1.BSRResolution, error) {
	if r.base == nil {
		return v1.BSRResolution{}, fmt.Errorf("BSR resolver is not configured")
	}
	binding, err := r.base.Resolve(ctx, dependency)
	if err != nil {
		return v1.BSRResolution{}, fmt.Errorf("Resolve: %w", err)
	}
	if err := binding.Validate(); err != nil {
		return v1.BSRResolution{}, fmt.Errorf("Validate: %w", err)
	}
	if version, exists := r.pins[binding.Git.Module]; exists && binding.Resolution == v1.BSRCompatibilitySnapshot {
		binding.Git.Version = version
		if err := binding.Validate(); err != nil {
			return v1.BSRResolution{}, fmt.Errorf("Validate: %w", err)
		}
	}
	return binding, nil
}
