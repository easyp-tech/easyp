package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type namespaceTransitionPolicy int

const (
	_ namespaceTransitionPolicy = iota
	namespaceTransitionsChecked
	namespaceTransitionsTidyRepairs
)

type rootNamespaceTransition struct {
	fetched    Fetched
	inspection *RootInspection
}

func (source *importRootSource) validateRootTransitions(ctx context.Context) error {
	if len(source.transitions) == 0 {
		return nil
	}
	selector, supported := source.Source.(rootSelectionSource)
	if !supported {
		return fmt.Errorf("repository cannot verify import namespace transitions")
	}
	for _, transition := range source.transitions {
		err := source.validateRootTransition(ctx, transition.fetched, transition.inspection, selector)
		if err != nil {
			return fmt.Errorf("validateRootTransition: %w", err)
		}
	}
	return nil
}

func (source *importRootSource) validateRootTransition(ctx context.Context, fetched Fetched, inspection *RootInspection, selector rootSelectionSource) error {
	old, exists := source.locked[fetched.Lock.Source]
	if !exists || !fetched.Module.RootsFromMetadata || strings.EqualFold(old.Commit, fetched.Lock.Commit) || len(source.hints[old.Source]) > 0 {
		return nil
	}
	oldRoots, newRoots := slices.Clone(old.Roots), slices.Clone(fetched.Module.Roots)
	slices.Sort(oldRoots)
	slices.Sort(newRoots)
	if len(oldRoots) > 0 && slices.Equal(oldRoots, newRoots) {
		return nil
	}
	// Compare source names from immutable revisions without installing either
	// graph. Replaying the old scope is also a proof of its original lock hash.
	previous, err := source.lockedRootScope(ctx, selector, old)
	if err != nil {
		return fmt.Errorf("lockedRootScope: %w", err)
	}
	if previous.Module.RootsFromMetadata {
		return nil
	}
	// An omitted lock selection can mean metadata-free default '.', rather
	// than prior authority. Derive that scope from verified pinned metadata.
	oldRoots = slices.Clone(previous.Module.Roots)
	slices.Sort(oldRoots)
	if slices.Equal(oldRoots, newRoots) {
		return nil
	}
	if previous.Inspection == nil {
		previous, err = fetchLockedRootScope(ctx, selector, old, oldRoots)
		if err != nil {
			return fmt.Errorf("fetchLockedRootScope: %w", err)
		}
	}
	if previous.Inspection == nil || inspection == nil || inspection.Provisional {
		return fmt.Errorf("module %s: cannot verify the old and new import namespaces from complete pinned inspections", old.Source)
	}
	before := rootNamespaceNames(previous.Module, previous.Inspection)
	after := rootNamespaceNames(fetched.Module, inspection)
	var missing []string
	newPhysical := make(map[string]bool, len(after))
	for _, file := range after {
		newPhysical[rootGitSourcePath(file, fetched.Lock)] = true
	}
	for name, file := range before {
		current, retained := after[name]
		physical := rootGitSourcePath(file, previous.Lock)
		// An existing public name can shadow a surviving old Git source.
		// Its bytes/package/basename do not establish source continuity.
		if !retained || (physical != rootGitSourcePath(current, fetched.Lock) && newPhysical[physical]) {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	slices.Sort(missing)
	changes := representativeRootRenames(missing, before, after, previous.Lock, fetched.Lock)
	return fmt.Errorf("module %s changes its import namespace: old %s commit %s has verified fallback roots %v; new %s commit %s has roots %v from authoritative dependency metadata; renamed or removed imports: %s; review consumer imports and generated SDK paths, then explicitly adopt with %s; or set its requirement to %s in protobuf.mod and run easyp mod tidy to apply verified import renames; or pin the previous revision with easyp get %s", old.Source, old.Version, old.Commit, oldRoots, fetched.Lock.Version, fetched.Lock.Commit, newRoots, strings.Join(changes, "; "), rootAdoptionCommand(fetched.Lock, newRoots), fetched.Lock.Version, quoteRootCommandArg(old.Source+"@"+old.Commit))
}

func fetchLockedRootScope(ctx context.Context, selector rootSelectionSource, entry v1.LockedModule, roots []string) (Fetched, error) {
	fetched, err := selector.FetchWithRoots(ctx, entry.Source, entry.Commit, roots)
	if err != nil {
		return Fetched{}, fmt.Errorf("FetchWithRoots: %w", err)
	}
	fetched = cloneRootFetched(fetched)
	if fetched.Module.Name != entry.Source || fetched.Lock.Source != entry.Source || !strings.EqualFold(fetched.Lock.Commit, entry.Commit) || fetched.Lock.Hash != entry.Hash {
		return Fetched{}, fmt.Errorf("%w: %s@%s: previous root scope differs from locked commit %s hash %s", ErrLockedVersionChanged, entry.Source, entry.Version, entry.Commit, entry.Hash)
	}
	if fetched.Inspection != nil && fetched.Inspection.Provisional {
		return Fetched{}, fmt.Errorf("module %s: previous root fetch remains provisional", entry.Source)
	}
	return fetched, nil
}

func rootNamespaceNames(module v1.Module, inspection *RootInspection) map[string]RootProtoFile {
	names := make(map[string]RootProtoFile)
	for _, file := range inspection.Files {
		for _, root := range module.Roots {
			name, within := importRootRelative(root, file.Path)
			if within {
				names[name] = file
			}
		}
	}
	return names
}

func representativeRootRenames(missing []string, before, after map[string]RootProtoFile, old, current v1.LockedModule) []string {
	newNames := make([]string, 0, len(after))
	for name := range after {
		newNames = append(newNames, name)
	}
	slices.Sort(newNames)
	var changes []string
	for _, name := range missing[:min(len(missing), 3)] {
		oldFile := before[name]
		physical := rootGitSourcePath(oldFile, old)
		renamed := ""
		for _, candidate := range newNames {
			file := after[candidate]
			newPhysical := rootGitSourcePath(file, current)
			if physical == newPhysical || oldFile.Path == file.Path {
				renamed = candidate
				break
			}
		}
		if renamed == "" {
			changes = append(changes, name+" is no longer exported")
			continue
		}
		changes = append(changes, name+" -> "+renamed)
	}
	return changes
}

func rootGitSourcePath(file RootProtoFile, entry v1.LockedModule) string {
	return strings.TrimPrefix(file.Identity, entry.Source+"@"+entry.Commit+":")
}

func rootAdoptionCommand(entry v1.LockedModule, roots []string) string {
	var command strings.Builder
	command.WriteString("easyp get")
	for _, root := range roots {
		command.WriteString(" --import-root ")
		command.WriteString(quoteRootCommandArg(root))
	}
	command.WriteByte(' ')
	command.WriteString(quoteRootCommandArg(entry.Source + "@" + entry.Version))
	return command.String()
}

func quoteRootCommandArg(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
