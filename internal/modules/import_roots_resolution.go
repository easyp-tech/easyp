package modules

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type rootInspectionSource interface {
	FetchForRootResolution(context.Context, string, string) (Fetched, error)
}

type rootSelectionSource interface {
	FetchWithRoots(context.Context, string, string, []string) (Fetched, error)
}

type importRootSource struct {
	Source
	locked         map[string]v1.LockedModule
	hints          map[string][]string
	fetched        map[rootRevisionKey]Fetched
	verifiedScopes map[string]Fetched
	transitions    []rootNamespaceTransition
}

func newImportRootSource(backend Source, existing v1.Lock, hints map[string][]string) *importRootSource {
	source := &importRootSource{
		Source: backend,
		locked: make(map[string]v1.LockedModule, len(existing.Modules)),
		hints:  make(map[string][]string, len(hints)),
	}
	for _, entry := range existing.Modules {
		source.locked[entry.Source] = cloneRootPin(entry)
	}
	for name, roots := range hints {
		source.hints[name] = slices.Clone(roots)
	}
	return source
}

func (source *importRootSource) Fetch(ctx context.Context, name, version string) (Fetched, error) {
	var fetched Fetched
	var err error
	hint := source.hints[name]
	inspector, inspectionSupported := source.Source.(rootInspectionSource)
	switch {
	case len(hint) > 0:
		selector, supported := source.Source.(rootSelectionSource)
		if !supported {
			return Fetched{}, fmt.Errorf("module %s: repository cannot verify explicit import roots", name)
		}
		// Scope explicit hints before the first strict fetch so unused aliases
		// outside that root cannot preempt its selection.
		fetched, err = selector.FetchWithRoots(ctx, name, version, hint)
		if err == nil && inspectionSupported && fetched.Inspection == nil {
			var inspected Fetched
			inspected, err = inspector.FetchForRootResolution(ctx, name, fetched.Lock.Commit)
			if err == nil {
				fetched.Inspection = inspected.Inspection
				if fetched.Inspection != nil {
					inspection := *fetched.Inspection
					inspection.Provisional = false
					fetched.Inspection = &inspection
				}
			}
		}
	case inspectionSupported:
		fetched, err = inspector.FetchForRootResolution(ctx, name, version)
	default:
		fetched, err = source.Source.Fetch(ctx, name, version)
	}
	if err != nil {
		return Fetched{}, fmt.Errorf("Fetch: %w", err)
	}
	fetched = cloneRootFetched(fetched)
	if v1.IsCommitRef(version) && !strings.EqualFold(version, fetched.Lock.Commit) {
		return Fetched{}, fmt.Errorf("module %s: fetched commit %s differs from requested commit %s", name, fetched.Lock.Commit, version)
	}
	if err := source.verifyRootFetch(ctx, fetched); err != nil {
		return Fetched{}, err
	}
	if len(hint) == 0 && !fetched.Module.RootsFromMetadata {
		if old := source.locked[name]; len(old.Roots) > 0 || source.sameLockedPin(fetched.Lock) {
			if selector, supported := source.Source.(rootSelectionSource); supported {
				scoped, err := selector.FetchWithRoots(ctx, name, fetched.Lock.Commit, old.Roots)
				if err != nil {
					return Fetched{}, fmt.Errorf("FetchWithRoots: %w", err)
				}
				if scoped.Module.Name != name || scoped.Lock.Source != name || !strings.EqualFold(scoped.Lock.Commit, fetched.Lock.Commit) {
					return Fetched{}, fmt.Errorf("module %s: retained-root fetch differs from resolved commit %s", name, fetched.Lock.Commit)
				}
				scoped.Lock.Version = fetched.Lock.Version
				fetched = cloneRootFetched(scoped)
			} else {
				fetched.Module.Roots = slices.Clone(old.Roots)
				fetched.Lock.Roots = slices.Clone(old.Roots)
			}
		}
	}
	if err := source.verifyRootFetch(ctx, fetched); err != nil {
		return Fetched{}, err
	}
	if source.fetched == nil {
		source.fetched = make(map[rootRevisionKey]Fetched)
	}
	source.fetched[rootRevision(name, fetched.Lock.Commit)] = cloneRootFetched(fetched)
	return fetched, nil
}

func (source *importRootSource) rootCandidates(lock v1.Lock) []importRootModule {
	dependencies := make([]importRootModule, 0, len(lock.Modules))
	for _, entry := range lock.Modules {
		fetched := source.fetched[rootRevision(entry.Source, entry.Commit)]
		fixed := source.fixedRootScope(entry, fetched)
		dependencies = append(dependencies, importRootModule{
			name:       entry.Source,
			roots:      slices.Clone(fetched.Module.Roots),
			fixed:      fixed,
			inspection: cloneRootInspection(fetched.Inspection),
		})
	}
	return dependencies
}

func (source *importRootSource) finalize(ctx context.Context, lock v1.Lock) (v1.Lock, error) {
	dependencies := source.rootCandidates(lock)
	inspected := slices.ContainsFunc(dependencies, func(dependency importRootModule) bool { return dependency.inspection != nil })
	if !inspected {
		return lock, nil
	}
	selected, err := selectImportRoots(ctx, dependencies)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("selectImportRoots: %w", err)
	}
	return source.finalizeSelections(ctx, lock, dependencies, selected)
}

func (source *importRootSource) finalizeSelections(ctx context.Context, lock v1.Lock, dependencies []importRootModule, selected [][]string) (v1.Lock, error) {
	inspected := slices.ContainsFunc(lock.Modules, func(entry v1.LockedModule) bool {
		return source.fetched[rootRevision(entry.Source, entry.Commit)].Inspection != nil
	})
	if !inspected {
		return lock, nil
	}
	selector, supported := source.Source.(rootSelectionSource)
	if !supported {
		return v1.Lock{}, fmt.Errorf("repository cannot finalize inspected import roots")
	}
	for offset, entry := range lock.Modules {
		key := rootRevision(entry.Source, entry.Commit)
		previous := source.fetched[key]
		dependency := dependencies[offset]
		if previous.Inspection == nil {
			continue
		}
		roots := selected[offset]
		if dependency.fixed && !previous.Module.RootsFromMetadata {
			roots = dependency.roots
			if len(source.hints[entry.Source]) == 0 && source.sameLockedPin(entry) {
				roots = source.locked[entry.Source].Roots
			}
		}
		// Finalization uses the exact MVS-selected revision, never HEAD. Keep
		// its semantic Version separately from the exact-commit fetch request.
		fetched, err := selector.FetchWithRoots(ctx, entry.Source, entry.Commit, roots)
		if err != nil {
			return v1.Lock{}, fmt.Errorf("FetchWithRoots: %w", err)
		}
		fetched = cloneRootFetched(fetched)
		if fetched.Module.Name != entry.Source || fetched.Lock.Source != entry.Source || !strings.EqualFold(fetched.Lock.Commit, entry.Commit) {
			return v1.Lock{}, fmt.Errorf("module %s: finalized revision differs from resolved commit %s", entry.Source, entry.Commit)
		}
		if fetched.Inspection != nil && fetched.Inspection.Provisional {
			return v1.Lock{}, fmt.Errorf("module %s: final root fetch remains provisional", entry.Source)
		}
		fetched.Lock.Version = entry.Version
		if err := source.verifyRootFetch(ctx, fetched); err != nil {
			return v1.Lock{}, err
		}
		inspection := previous.Inspection
		// Keep the inspection used for namespace comparison inside resolution.
		// Graph orchestration applies its explicit policy after finalization.
		source.transitions = append(source.transitions, rootNamespaceTransition{fetched: fetched, inspection: inspection})
		// Keep the complete final scope available to tidy's source-binding proof.
		if fetched.Inspection == nil && inspection != nil && fetched.Lock.Hash != "" {
			complete := *inspection
			complete.Provisional = false
			fetched.Inspection = &complete
		}
		source.fetched[key] = fetched
		lock.Modules[offset] = fetched.Lock
	}
	if err := lock.Validate(); err != nil {
		return v1.Lock{}, fmt.Errorf("Validate: %w", err)
	}
	return lock, nil
}

func inspectLocalImportRoots(directory string, module v1.Module) (importRootModule, error) {
	roots, err := ModuleSources(directory, module)
	if err != nil {
		return importRootModule{}, fmt.Errorf("ModuleSources: %w", err)
	}
	inspection := &RootInspection{}
	seen := make(map[string]bool)
	for _, root := range roots {
		err := roots.WalkSelected(root, roots.FileAllowed(), func(name string) error {
			if seen[name] {
				return nil
			}
			seen[name] = true
			content, err := roots.ReadSourceFile(name)
			if err != nil {
				return fmt.Errorf("ReadSourceFile: %w", err)
			}
			path, err := filepath.Rel(directory, name)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			inspection.Files = append(inspection.Files, RootProtoFile{Path: filepath.ToSlash(path), Identity: physicalSourcePath(name), Content: content})
			return nil
		})
		if err != nil {
			return importRootModule{}, fmt.Errorf("WalkSelected: %w", err)
		}
	}
	return importRootModule{name: module.Name, roots: module.Roots, fixed: true, inspection: inspection}, nil
}
