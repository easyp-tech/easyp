package gitmodules

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/modules"
)

// PreparationSource advertises preparation for this cache receiver. Embedding
// wrappers that override fetch behavior must explicitly opt in for themselves.
func (c *Cache) PreparationSource() modules.Source { return c }

// PrepareFetch acquires an independent immutable revision before the solver
// consumes it. Cache publication keeps the existing snapshot/repository locks.
// The callback defers metadata backends, which need not support concurrent use.
// No temporary checkout survives preparation, including discarded graph edges.
func (c *Cache) PrepareFetch(ctx context.Context, request modules.FetchRequest) (modules.PreparedFetch, error) {
	fetched, err := c.fetchSnapshot(ctx, request.Source, request.Version, request.Roots, request.Inspect || len(request.Roots) > 0, true)
	if err != nil {
		return nil, fmt.Errorf("fetchSnapshot: %w", err)
	}
	return func(ctx context.Context) (modules.Fetched, error) {
		if err := ctx.Err(); err != nil {
			return modules.Fetched{}, fmt.Errorf("Err: %w", err)
		}
		module, bindings, err := c.resolveBSRDependencies(ctx, fetched.Module)
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("resolveBSRDependencies: %w", err)
		}
		fetched.Module, fetched.Lock.BSR = module, bindings
		return fetched, nil
	}, nil
}
