package modules

import (
	"context"
	"fmt"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Cache verifies locked contents and reads installed module metadata.
type Cache interface {
	Install(context.Context, v1.Lock) error
	Cached(v1.LockedModule) (string, v1.Module, error)
}

// Repository resolves revisions and installs their locked contents.
type Repository interface {
	Source
	Cache
}

// VersionedRepository also lists available versions for module updates.
type VersionedRepository interface {
	Repository
	Versions(context.Context, string) ([]string, error)
}

// VendorDir is the import root used for vendored dependencies.
const VendorDir = "easyp_vendor"

// EnsureSources resolves the effective overlay or installs published locked contents.
// Local sources precede locked sources, preserving import and managed-selector order.
func EnsureSources(ctx context.Context, moduleDir string, module v1.Module, cache Cache) (SourceRoots, error) {
	if len(module.Replaces) > 0 {
		graph, err := EnsureEffectiveGraph(ctx, moduleDir, module, cache, nil)
		if err != nil {
			return nil, fmt.Errorf("EnsureEffectiveGraph: %w", err)
		}
		return graph.Sources, nil
	}
	return EnsureLockedSources(ctx, moduleDir, module, cache)
}

// LocalSources reads only the local-to-local portion of the main replacement
// graph. Consumers needing remote edges must use EnsureSources instead.
func LocalSources(moduleDir string, module v1.Module) (SourceRoots, error) {
	return localDependencySources(moduleDir, module)
}
