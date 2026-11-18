package migration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/bufbuild/protocompile"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// proveGitSourceBindings compares selected and reachable filenames/bytes, and
// optionally compiles the proved targets. Apply repeats the binding/collision
// proof with compilation deferred, using the same immutable Git observations.
func proveGitSourceBindings(ctx context.Context, localRoot string, local *localSelectionProof, proof *gitSelectionProof, compile bool) (map[string]migrationProtoSource, error) {
	witnesses := make(map[string]migrationProtoSource)
	if proof == nil {
		return witnesses, nil
	}
	filtered := false
	for _, dependency := range proof.fetched {
		filtered = filtered || len(dependency.Module.ProtoFilters) > 0
	}
	if len(proof.selections) == 0 && !filtered {
		return witnesses, nil
	}
	entries := slices.Clone(proof.entries)
	// Import-only Buf dependencies can omit fixtures that v0 archived. Prove the
	// local generation/import closure even without any git_repo generation input.
	if filtered && len(local.inputs) > 0 && !slices.ContainsFunc(entries, func(entry v1.GenerateModule) bool { return entry.Module == local.module }) {
		entries = append(entries, v1.GenerateModule{Module: local.module})
	}

	namespaces, err := observeMigrationNamespaces(localRoot, local, proof)
	if err != nil {
		return nil, fmt.Errorf("observeMigrationNamespaces: %w", err)
	}
	for _, entry := range entries {
		legacy := migrationImportNamespace{primary: namespaces.legacyLocal}
		current := migrationImportNamespace{primary: namespaces.currentLocal}
		var targets []string
		if selection, remote := proof.selections[entry.Module]; remote {
			dependency := proof.fetched[entry.Module]
			if dependency.Inspection == nil {
				continue
			}
			current.primary = []map[string]migrationProtoSource{namespaces.currentDependencies[entry.Module]}
			targets = legacyGitSelectionTargets(selection, dependency)
		} else {
			targets = slices.Sorted(maps.Keys(local.selection.files))
		}
		for _, name := range slices.Sorted(maps.Keys(namespaces.legacyDependencies)) {
			legacy.dependencies = append(legacy.dependencies, namespaces.legacyDependencies[name])
			if name != entry.Module {
				current.dependencies = append(current.dependencies, namespaces.currentDependencies[name])
			}
		}
		if err := current.checkSourceCollisions(); err != nil {
			return nil, fmt.Errorf("checkSourceCollisions: generation module %s: %w", entry.Module, err)
		}
		slices.Sort(targets)
		targets = slices.Compact(targets)
		bindings, err := proveMigrationImportBindings(ctx, entry.Module, targets, legacy, current, compile)
		if err != nil {
			return nil, fmt.Errorf("proveMigrationImportBindings: %w", err)
		}
		for name, file := range bindings {
			witnesses[entry.Module+":"+name] = file
		}
	}
	return witnesses, nil
}

func proveMigrationImportBindings(ctx context.Context, module string, targets []string, legacy, current migrationImportNamespace, compile bool) (map[string]migrationProtoSource, error) {
	bindings := make(map[string]migrationProtoSource)
	queue := slices.Clone(targets)
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("Err: %w", err)
		}
		name := queue[0]
		queue = queue[1:]
		if _, seen := bindings[name]; seen {
			continue
		}
		old, oldErr := legacy.resolve(name)
		file, err := current.resolve(name)
		if joined := errors.Join(oldErr, err); joined != nil {
			return nil, fmt.Errorf("resolve: generation module %s filename %q cannot preserve its legacy source binding: %w", module, name, joined)
		}
		if old.identity != file.identity || old.path != file.path || !bytes.Equal(old.content, file.content) {
			return nil, fmt.Errorf("generation module %s filename %q source binding changed from %s:%s to %s:%s; manual migration is required", module, name, old.module, old.path, file.module, file.path)
		}
		imports, err := modules.ParseProtoImports(name, file.content)
		if err != nil {
			return nil, fmt.Errorf("ParseProtoImports: %s:%s: %w", module, name, err)
		}
		bindings[name] = file
		queue = append(queue, imports...)
	}
	if compile && len(targets) > 0 {
		compiler := protocompile.Compiler{MaxParallelism: 1, Resolver: &protocompile.SourceResolver{Accessor: func(name string) (io.ReadCloser, error) {
			file, err := current.resolve(name)
			if err != nil {
				return nil, fmt.Errorf("resolve: %w", err)
			}
			return io.NopCloser(bytes.NewReader(file.content)), nil
		}}}
		if _, err := compiler.Compile(ctx, targets...); err != nil {
			return nil, fmt.Errorf("Compile: generation module %s: %w", module, err)
		}
	}
	return bindings, nil
}

func equalMigrationBindings(first, second map[string]migrationProtoSource) bool {
	return maps.EqualFunc(first, second, func(first, second migrationProtoSource) bool {
		return first.module == second.module && first.path == second.path && first.identity == second.identity && bytes.Equal(first.content, second.content)
	})
}
