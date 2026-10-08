package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// tidyImportPlanner compares checked current evidence with exact previous pins.
// It uses repository capabilities only to prove previously selected root scopes.
type tidyImportPlanner struct {
	current    graphResolveResult
	previous   v1.Lock
	repository Source
}

func newTidyImportPlanner(current graphResolveResult, previous v1.Lock, repository Source) tidyImportPlanner {
	return tidyImportPlanner{current: current, previous: cloneRootLock(previous), repository: repository}
}

type tidyImportBinding struct {
	before Fetched
	after  Fetched
	file   RootProtoFile
	names  map[string]RootProtoFile
}

func (planner tidyImportPlanner) tidyImportBindings(ctx context.Context, needed map[string]v1UnresolvedImport, view *tidySourceView) (map[string]tidyImportBinding, error) {
	bindings := make(map[string]tidyImportBinding)
	if len(needed) == 0 {
		return bindings, nil
	}
	selector, supported := planner.repository.(rootSelectionSource)
	before, after := planner.previous, planner.current.lockFile()
	current := make(map[string]v1.LockedModule, len(after.Modules))
	for _, entry := range after.Modules {
		current[entry.Source] = entry
	}
	for _, old := range before.Modules {
		entry, retained := current[old.Source]
		if retained && strings.EqualFold(old.Commit, entry.Commit) && slices.Equal(old.Roots, entry.Roots) {
			continue
		}
		if !supported {
			imported, used, err := planner.uninspectedTidyBinding(ctx, old, needed, view)
			if err != nil {
				return nil, fmt.Errorf("uninspectedTidyBinding: %w", err)
			}
			if !used {
				continue
			}
			return nil, fmt.Errorf("%s: module %s: cannot verify the previous import namespace from old %s commit %s recorded roots %v to new %s commit %s recorded roots %v; repository does not support pinned root inspection; keep the previous manifest requirement or use a repository with checked root inspection before running easyp mod tidy", imported, old.Source, old.Version, old.Commit, old.Roots, entry.Version, entry.Commit, entry.Roots)
		}
		previous, err := planner.lockedRootScope(ctx, selector, old)
		if err != nil {
			return nil, fmt.Errorf("lockedRootScope: %w", err)
		}
		if previous.Inspection == nil {
			previous, err = fetchLockedRootScope(ctx, selector, old, previous.Module.Roots)
			if err != nil {
				return nil, fmt.Errorf("fetchLockedRootScope: %w", err)
			}
		}
		previous.Lock.Version = old.Version
		next, _ := planner.current.currentRevision(entry)
		if previous.Inspection == nil || previous.Inspection.Provisional || (retained && (next.Inspection == nil || next.Inspection.Provisional || next.Lock.Hash == "" || next.Lock.Hash != entry.Hash)) {
			return nil, fmt.Errorf("module %s: cannot verify old %s commit %s roots %v and new %s commit %s roots %v from complete pinned inspections; restore the old pin or verify explicit producer roots before running easyp mod tidy", old.Source, old.Version, old.Commit, previous.Module.Roots, entry.Version, entry.Commit, next.Module.Roots)
		}
		names := make(map[string]RootProtoFile)
		if retained {
			names, err = tidyRootNamespace(next)
			if err != nil {
				return nil, fmt.Errorf("tidyRootNamespace: %w", err)
			}
		}
		previousNames, err := tidyRootNamespace(previous)
		if err != nil {
			return nil, fmt.Errorf("tidyRootNamespace: %w", err)
		}
		for name, file := range previousNames {
			if _, used := needed[name]; !used {
				continue
			}
			if prior, exists := bindings[name]; exists && prior.before.Lock.Source != old.Source {
				return nil, fmt.Errorf("old locked namespace has duplicate import %q from %s and %s", name, prior.before.Lock.Source, old.Source)
			}
			bindings[name] = tidyImportBinding{before: previous, after: next, file: file, names: names}
		}
	}
	return bindings, nil
}

func (planner tidyImportPlanner) lockedRootScope(ctx context.Context, selector rootSelectionSource, entry v1.LockedModule) (Fetched, error) {
	if previous, exists := planner.current.previousRootScope(entry); exists {
		return previous, nil
	}
	previous, err := fetchLockedRootScope(ctx, selector, entry, entry.Roots)
	if err != nil {
		return Fetched{}, fmt.Errorf("fetchLockedRootScope: %w", err)
	}
	return previous, nil
}

func (planner tidyImportPlanner) uninspectedTidyBinding(ctx context.Context, entry v1.LockedModule, needed map[string]v1UnresolvedImport, view *tidySourceView) (v1UnresolvedImport, bool, error) {
	repository, supported := planner.repository.(Repository)
	if !supported {
		return v1UnresolvedImport{}, false, fmt.Errorf("module %s: repository cannot verify the old locked import owners; keep commit %s roots %v or use checked root inspection before running easyp mod tidy", entry.Source, entry.Commit, entry.Roots)
	}
	names, err := view.captureOldNamespace(ctx, repository, entry)
	if err != nil {
		return v1UnresolvedImport{}, false, fmt.Errorf("captureOldNamespace: %w", err)
	}
	var imports []string
	for name := range needed {
		imports = append(imports, name)
	}
	slices.Sort(imports)
	for _, name := range imports {
		if _, exported := names[name]; exported {
			return needed[name], true, nil
		}
	}
	return v1UnresolvedImport{}, false, nil
}

func tidyRootNamespace(fetched Fetched) (map[string]RootProtoFile, error) {
	names := make(map[string]RootProtoFile)
	prefix := fetched.Lock.Source + "@" + fetched.Lock.Commit + ":"
	for _, file := range fetched.Inspection.Files {
		for _, root := range fetched.Module.Roots {
			name, within := importRootRelative(root, file.Path)
			if !within {
				continue
			}
			if !strings.HasPrefix(file.Identity, prefix) || !ValidProtoImportPath(strings.TrimPrefix(file.Identity, prefix)) {
				return nil, fmt.Errorf("module %s has an unverified pinned source identity for %s at commit %s", fetched.Lock.Source, file.Path, fetched.Lock.Commit)
			}
			if previous, exists := names[name]; exists && previous.Path != file.Path {
				return nil, fmt.Errorf("module %s has duplicate import path %q: %s and %s", fetched.Lock.Source, name, previous.Path, file.Path)
			}
			names[name] = file
		}
	}
	return names, nil
}

func (binding tidyImportBinding) replacement(path, name string) (string, error) {
	physical := rootGitSourcePath(binding.file, binding.before.Lock)
	var names []string
	for current, file := range binding.names {
		if physical == rootGitSourcePath(file, binding.after.Lock) {
			names = append(names, current)
		}
	}
	if current, retained := binding.names[name]; retained && (len(names) == 0 || rootGitSourcePath(current, binding.after.Lock) == physical) {
		// A moved source retaining its public name remains ordinary producer
		// evolution. A surviving source under a new name has priority instead.
		return name, nil
	}
	if len(names) == 1 {
		return names[0], nil
	}
	slices.Sort(names)
	reason := "the previous source is no longer exported by the same producer"
	if len(names) > 1 {
		reason = fmt.Sprintf("the previous physical source has ambiguous current names %v", names)
	}
	return "", fmt.Errorf("%s imports %q from module %s: old %s commit %s roots %v source %s; new %s commit %s roots %v: %s; restore the previous revision or adjust the consumer import after checking producer roots and generated SDK paths, then run easyp mod tidy", path, name, binding.before.Lock.Source, binding.before.Lock.Version, binding.before.Lock.Commit, binding.before.Module.Roots, binding.file.Path, binding.after.Lock.Version, binding.after.Lock.Commit, binding.after.Module.Roots, reason)
}
