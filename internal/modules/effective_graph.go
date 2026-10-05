package modules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// EffectiveModule describes source metadata without inventing a local revision.
type EffectiveModule struct {
	Directory string
	Module    v1.Module
}

// EffectiveGraph is an ephemeral selection of local sources and verified remote
// snapshots. It is never serialized to the shared protobuf.lock.
type EffectiveGraph struct {
	Modules map[string]EffectiveModule
	Sources SourceRoots
}

// EnsureEffectiveGraph resolves main-module replacements at every graph depth.
// localPath optionally maps a used replacement into a historical snapshot.
// Unused replacements are neither opened nor acquired.
func EnsureEffectiveGraph(ctx context.Context, root string, module v1.Module, cache Cache, localPath func(string) (string, error)) (EffectiveGraph, error) {
	return ensureEffectiveGraph(ctx, root, module, cache, localPath, false)
}

func ensureEffectiveGraph(ctx context.Context, root string, module v1.Module, cache Cache, localPath func(string) (string, error), refresh bool) (EffectiveGraph, error) {
	graph := EffectiveGraph{Modules: make(map[string]EffectiveModule)}
	existing, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return graph, fmt.Errorf("ReadLock: %w", err)
	}
	pins := make(map[string]v1.LockedModule, len(existing.Modules))
	for _, entry := range existing.Modules {
		pins[entry.Source] = entry
	}
	locals := newLocalReplacements(root, module, localPath)
	remote := &effectiveRemoteSource{cache: cache, pins: pins, refresh: refresh, historical: localPath != nil, modules: make(map[[4]string]cachedEffectiveModule)}
	loader := revisionLoader{source: remote, fetched: make(map[string]Fetched), pins: pins, local: locals.lookupRequirement}
	if refresh {
		loader.pins = nil
	}
	pass, err := loader.resolve(ctx, module)
	if err != nil {
		return graph, fmt.Errorf("module %s: %w", module.Name, err)
	}
	appendModule := func(name string, entry EffectiveModule) error {
		roots, err := ModuleSources(entry.Directory, entry.Module)
		if err != nil {
			return fmt.Errorf("ModuleSources: %s: %w", name, err)
		}
		graph.Modules[name] = entry
		graph.Sources = append(graph.Sources, roots...)
		return nil
	}
	for _, name := range pass.locals {
		if err := appendModule(name, locals.modules[name]); err != nil {
			return graph, err
		}
	}
	for _, entry := range pass.lock.Modules {
		selected, err := remote.installed(ctx, entry)
		if err != nil {
			return graph, fmt.Errorf("installed: %w", err)
		}
		if err := appendModule(entry.Source, selected); err != nil {
			return graph, err
		}
	}
	return graph, nil
}

type effectiveRemoteSource struct {
	cache      Cache
	pins       map[string]v1.LockedModule
	refresh    bool
	historical bool
	modules    map[[4]string]cachedEffectiveModule
}

type cachedEffectiveModule struct {
	module EffectiveModule
	bsr    []v1.BSRResolution
}

func (s *effectiveRemoteSource) Fetch(ctx context.Context, name, version string) (Fetched, error) {
	// Reuse the published identity for this revision. A higher old tag is not
	// an extra requirement and must not override explicit tag/commit constraints.
	// Install verifies hashes even for already cached contents.
	if pin, ok := s.pins[name]; ok && !s.refresh && (version == "" || version == pin.Version || (v1.IsCommitRef(version) && strings.EqualFold(version, pin.Commit))) {
		entry, err := s.installed(ctx, pin)
		if err != nil {
			return Fetched{}, fmt.Errorf("installed: %w", err)
		}
		return Fetched{Module: entry.Module, Lock: pin}, nil
	}
	if s.historical && version == "" {
		return Fetched{}, fmt.Errorf("historical local replacement dependency %s has no locked commit or explicit version; cannot substitute current HEAD for the baseline", name)
	}
	source, ok := s.cache.(Source)
	if !ok {
		return Fetched{}, fmt.Errorf("protobuf.lock does not satisfy %s %s; repository cannot resolve new local-overlay dependencies", name, version)
	}
	guard := lockedVersionSource{Source: source, locked: s.pins}
	fetched, err := guard.Fetch(ctx, name, version)
	if err != nil {
		return Fetched{}, fmt.Errorf("Fetch: %w", err)
	}
	return fetched, nil
}

func (s *effectiveRemoteSource) installed(ctx context.Context, entry v1.LockedModule) (EffectiveModule, error) {
	key := [4]string{entry.Source, entry.Version, entry.Commit, entry.Hash}
	if selected, ok := s.modules[key]; ok && slices.Equal(selected.bsr, entry.BSR) {
		return selected.module, nil
	}
	if err := s.cache.Install(ctx, v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}); err != nil {
		return EffectiveModule{}, fmt.Errorf("Install: %w", err)
	}
	dir, module, err := s.cache.Cached(entry)
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("Cached: %w", err)
	}
	selected := EffectiveModule{Directory: dir, Module: module}
	s.modules[key] = cachedEffectiveModule{module: selected, bsr: slices.Clone(entry.BSR)}
	return selected, nil
}

func validateLocalOverlay(ctx context.Context, root string, module v1.Module, cache Cache, refresh bool) error {
	graph, err := ensureEffectiveGraph(ctx, root, module, cache, nil, refresh)
	if err != nil {
		return fmt.Errorf("ensureEffectiveGraph: %w", err)
	}
	own, err := ModuleSources(root, module)
	if err != nil {
		return fmt.Errorf("ModuleSources: %w", err)
	}
	if err := CheckSourceCollisions(append(own, graph.Sources...)); err != nil {
		return fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	unresolved, err := findUnresolvedV1ImportsWithSources(root, module.Roots, graph.Sources)
	if err != nil {
		return fmt.Errorf("findUnresolvedV1ImportsWithSources: %w", err)
	}
	if len(unresolved) > 0 {
		return fmt.Errorf("module %s: cannot resolve imports %v", module.Name, unresolved)
	}
	return nil
}
