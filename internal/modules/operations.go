package modules

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Tidy resolves v1 requirements, repairs verified consumer import renames, and
// records published commits and hashes. Local replacements are validation only.
func Tidy(ctx context.Context, root string, repository Repository) error {
	_, err := TidyWithReport(ctx, root, repository)
	return err
}

func resolveV1Lock(ctx context.Context, root string, module v1.Module, existing v1.Lock, repository Repository, preserveHeads bool) (v1.Lock, error) {
	return resolveV1LockWithRoots(ctx, root, module, existing, repository, preserveHeads, nil)
}

func resolveV1LockWithRoots(ctx context.Context, root string, module v1.Module, existing v1.Lock, repository Repository, preserveHeads bool, hints map[string][]string) (v1.Lock, error) {
	resolved, err := resolveV1Graph(ctx, module, existing, repository, graphResolveRequest{
		preserveHeads: preserveHeads,
		hints:         hints,
		transitions:   namespaceTransitionsChecked,
	})
	if err != nil {
		return v1.Lock{}, fmt.Errorf("resolveV1Graph: %w", err)
	}
	lock := resolved.lockFile()
	if err := repository.Install(ctx, lock); err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	dependencyRoots, err := CachedSources(lock, repository)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	own, err := ModuleSources(root, module)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("ModuleSources: %w", err)
	}
	if err := CheckSourceCollisions(append(own, dependencyRoots...)); err != nil {
		return v1.Lock{}, fmt.Errorf("module %s: %w", module.Name, err)
	}
	unresolved, err := findUnresolvedV1ImportsWithSources(root, module.Roots, dependencyRoots)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("findUnresolvedV1ImportsWithSources: %w", err)
	}
	if len(unresolved) > 0 {
		return v1.Lock{}, fmt.Errorf("module %s: cannot resolve imports %v", module.Name, unresolved)
	}
	return lock, nil
}

type graphResolveRequest struct {
	preserveHeads bool
	hints         map[string][]string
	transitions   namespaceTransitionPolicy
}

func resolveV1Graph(ctx context.Context, module v1.Module, existing v1.Lock, repository Repository, request graphResolveRequest) (graphResolveResult, error) {
	switch request.transitions {
	case namespaceTransitionsChecked, namespaceTransitionsTidyRepairs:
	default:
		return graphResolveResult{}, fmt.Errorf("namespace transition policy is required")
	}
	pins := make(map[string]v1.LockedModule, len(existing.Modules))
	for _, entry := range existing.Modules {
		pins[entry.Source] = cloneRootPin(entry)
	}
	source := newImportRootSource(repository, existing, request.hints)
	if !request.preserveHeads {
		pins = nil
	}
	lock, err := Resolve(ctx, module, source, pins)
	if err != nil {
		return graphResolveResult{}, fmt.Errorf("Resolve: %w", err)
	}
	lock, err = source.finalize(ctx, lock)
	if err != nil {
		return graphResolveResult{}, fmt.Errorf("finalize: %w", err)
	}
	if request.transitions == namespaceTransitionsChecked {
		err = source.validateRootTransitions(ctx)
		if err != nil {
			return graphResolveResult{}, fmt.Errorf("validateRootTransitions: %w", err)
		}
	}
	result, err := source.checkedResult(lock)
	if err != nil {
		return graphResolveResult{}, fmt.Errorf("checkedResult: %w", err)
	}
	return result, nil
}

// Download installs published pins, or needed unreplaced snapshots in local
// overlay mode. It never changes the shared manifest or lock.
func Download(ctx context.Context, root string, repository Cache) error {
	_, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	if len(module.Replaces) > 0 {
		return validateLocalOverlay(ctx, root, module, repository, false)
	}
	lock, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil {
		return fmt.Errorf("read protobuf.lock; run easyp mod tidy: %w", err)
	}
	if err := ValidateRequirements(RemoteRequirements(module), lock); err != nil {
		return err
	}
	if err := repository.Install(ctx, lock); err != nil {
		return err
	}
	if repair, ok := repository.(interface {
		RepairPolicySources(context.Context, v1.Lock) error
	}); ok {
		if err := repair.RepairPolicySources(ctx, lock); err != nil {
			return fmt.Errorf("RepairPolicySources: %w", err)
		}
	}
	_, err = CachedSources(lock, repository)
	return err
}
