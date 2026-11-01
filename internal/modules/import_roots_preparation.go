package modules

import (
	"context"
	"fmt"
	"slices"
)

// PreparationSource explicitly opts this wrapper into independent preparation.
func (source *importRootSource) PreparationSource() Source { return source }

// PrepareFetch prepares only independent backend work. Root integrity checks,
// retained scope lookups, and namespace bookkeeping run when consumed serially.
func (source *importRootSource) PrepareFetch(ctx context.Context, request FetchRequest) (PreparedFetch, error) {
	backend, supported := sourceFetchPreparer(source.Source)
	if !supported {
		return func(ctx context.Context) (Fetched, error) {
			return source.Fetch(ctx, request.Source, request.Version)
		}, nil
	}
	hint := source.hints[request.Source]
	inspector, inspectionSupported := source.Source.(rootInspectionSource)
	if len(hint) > 0 {
		if _, supported := source.Source.(rootSelectionSource); !supported {
			return nil, fmt.Errorf("module %s: repository cannot verify explicit import roots", request.Source)
		}
	}
	request.Roots = slices.Clone(hint)
	request.Inspect = inspectionSupported
	prepared, err := backend.PrepareFetch(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("Fetch: %w", err)
	}
	return func(ctx context.Context) (Fetched, error) {
		fetched, err := prepared(ctx)
		if err != nil {
			return Fetched{}, fmt.Errorf("Fetch: %w", err)
		}
		if len(hint) > 0 && inspectionSupported && fetched.Inspection == nil {
			inspected, err := inspector.FetchForRootResolution(ctx, request.Source, fetched.Lock.Commit)
			if err != nil {
				return Fetched{}, fmt.Errorf("Fetch: %w", err)
			}
			fetched.Inspection = cloneRootInspection(inspected.Inspection)
			if fetched.Inspection != nil {
				fetched.Inspection.Provisional = false
			}
		}
		return source.acceptRootFetch(ctx, request.Source, request.Version, fetched)
	}, nil
}
