package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func (source *importRootSource) validateRootTransition(ctx context.Context, fetched Fetched, inspection *RootInspection, selector rootSelectionSource) error {
	old := source.locked[fetched.Lock.Source]
	if len(old.Roots) == 0 || !fetched.Module.RootsFromMetadata || strings.EqualFold(old.Commit, fetched.Lock.Commit) || len(source.hints[old.Source]) > 0 {
		return nil
	}
	oldRoots, newRoots := slices.Clone(old.Roots), slices.Clone(fetched.Module.Roots)
	slices.Sort(oldRoots)
	slices.Sort(newRoots)
	if slices.Equal(oldRoots, newRoots) {
		return nil
	}
	// Compare source names from immutable revisions without installing either
	// graph. Replaying the old scope is also a proof of its original lock hash.
	previous, err := selector.FetchWithRoots(ctx, old.Source, old.Commit, old.Roots)
	if err != nil {
		return fmt.Errorf("FetchWithRoots: %w", err)
	}
	if previous.Module.Name != old.Source || previous.Lock.Source != old.Source || !strings.EqualFold(previous.Lock.Commit, old.Commit) || previous.Lock.Hash != old.Hash {
		return fmt.Errorf("%w: %s@%s: previous root scope differs from locked commit %s hash %s", ErrLockedVersionChanged, old.Source, old.Version, old.Commit, old.Hash)
	}
	if previous.Inspection == nil || previous.Inspection.Provisional || inspection == nil || inspection.Provisional {
		return fmt.Errorf("module %s: cannot verify the old and new import namespaces from complete pinned inspections", old.Source)
	}
	if previous.Module.RootsFromMetadata {
		return nil
	}
	before := rootNamespaceNames(previous.Module, previous.Inspection)
	after := rootNamespaceNames(fetched.Module, inspection)
	var missing []string
	for name := range before {
		if _, retained := after[name]; !retained {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	slices.Sort(missing)
	changes := representativeRootRenames(missing, before, after, previous.Lock, fetched.Lock)
	return fmt.Errorf("module %s changes its import namespace: old %s commit %s has verified fallback roots %v; new %s commit %s has roots %v from authoritative dependency metadata; renamed or removed imports: %s; review consumer imports and generated SDK paths, then explicitly adopt with %s; or pin the previous revision with easyp get %s", old.Source, old.Version, old.Commit, fetched.Lock.Version, fetched.Lock.Commit, oldRoots, newRoots, strings.Join(changes, "; "), rootAdoptionCommand(fetched.Lock, newRoots), quoteRootCommandArg(old.Source+"@"+old.Commit))
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
		physical := strings.TrimPrefix(oldFile.Identity, old.Source+"@"+old.Commit+":")
		renamed := ""
		for _, candidate := range newNames {
			file := after[candidate]
			newPhysical := strings.TrimPrefix(file.Identity, current.Source+"@"+current.Commit+":")
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
