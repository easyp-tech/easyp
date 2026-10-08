package migration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile"

	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
	"github.com/easyp-tech/easyp/wellknownimports"
)

type migrationProtoSource struct {
	module, path, identity string
	content                []byte
	unavailable            bool
}

type migrationImportNamespace struct {
	primary      []map[string]migrationProtoSource
	dependencies []map[string]migrationProtoSource
}

// Native generation rejects duplicate names in all available source roots,
// including files outside its target selectors and reachable import closure.
func (n migrationImportNamespace) checkSourceCollisions() error {
	seen := make(map[string]migrationProtoSource)
	for _, namespace := range slices.Concat(n.primary, n.dependencies) {
		for _, name := range slices.Sorted(maps.Keys(namespace)) {
			file := namespace[name]
			if previous, exists := seen[name]; exists && (previous.module != file.module || previous.path != file.path) {
				return fmt.Errorf("duplicate import path %q: %s:%s and %s:%s", name, previous.module, previous.path, file.module, file.path)
			}
			seen[name] = file
		}
	}
	return nil
}

func (n migrationImportNamespace) resolve(name string) (migrationProtoSource, error) {
	if !modules.ValidProtoImportPath(name) {
		return migrationProtoSource{}, fmt.Errorf("invalid protobuf filename %q", name)
	}
	for _, primary := range n.primary {
		if file, ok := primary[name]; ok {
			if file.unavailable {
				return migrationProtoSource{}, fmt.Errorf("legacy dependency %s filename %q maps to %q outside the verified v1 source scope; its source binding cannot be preserved", file.module, name, file.path)
			}
			return file, nil
		}
	}
	var found migrationProtoSource
	for _, dependency := range n.dependencies {
		if file, ok := dependency[name]; ok {
			if file.unavailable {
				return migrationProtoSource{}, fmt.Errorf("legacy dependency %s filename %q maps to %q outside the verified v1 source scope; its source binding cannot be preserved", file.module, name, file.path)
			}
			if found.identity != "" && found.identity != file.identity {
				return migrationProtoSource{}, fmt.Errorf("ambiguous import %q resolves to %s:%s and %s:%s", name, found.module, found.path, file.module, file.path)
			}
			found = file
		}
	}
	if found.identity != "" {
		return found, nil
	}
	content, err := wellknownimports.Content.ReadFile(name)
	if err != nil {
		return migrationProtoSource{}, fmt.Errorf("ReadFile: unresolved protobuf filename %q: %w", name, err)
	}
	return migrationProtoSource{module: "protobuf", path: name, identity: "protobuf:" + name, content: content}, nil
}

func (p *Plan) proveGitSourceBindings(ctx context.Context, compile bool) (map[string]migrationProtoSource, error) {
	proof := p.git
	witnesses := make(map[string]migrationProtoSource)
	if proof == nil || len(proof.selections) == 0 {
		return witnesses, nil
	}
	legacyLocal, err := p.localImportNamespaces(false)
	if err != nil {
		return nil, fmt.Errorf("localImportNamespaces: %w", err)
	}
	// Released v0 reverses its collected import roots before compilation.
	slices.Reverse(legacyLocal)
	currentLocal, err := p.localImportNamespaces(true)
	if err != nil {
		return nil, fmt.Errorf("localImportNamespaces: %w", err)
	}
	legacyDependencies, currentDependencies := make(map[string]map[string]migrationProtoSource), make(map[string]map[string]migrationProtoSource)
	for _, name := range slices.Sorted(maps.Keys(proof.fetched)) {
		dependency := proof.fetched[name]
		if dependency.Inspection == nil || dependency.Inspection.LegacyFiles == nil {
			continue
		}
		files := make(map[string]modules.RootProtoFile)
		for _, file := range dependency.Inspection.Files {
			files[file.Path] = file
		}
		legacy := make(map[string]migrationProtoSource)
		for installed, logical := range dependency.Inspection.LegacyFiles {
			legacy[installed] = migrationProtoSource{module: name, path: logical, unavailable: true}
			if file, ok := files[logical]; ok {
				legacy[installed] = migrationGitSource(name, file)
			}
		}
		if selection, ok := proof.selections[name]; ok && selection.root != "." {
			for _, installed := range slices.Sorted(maps.Keys(dependency.Inspection.LegacyFiles)) {
				if !strings.HasPrefix(installed, selection.root+"/") {
					continue
				}
				alias := strings.TrimPrefix(installed, selection.root+"/")
				logical := dependency.Inspection.LegacyFiles[installed]
				legacy[alias] = migrationProtoSource{module: name, path: logical, unavailable: true}
				if file, ok := files[logical]; ok {
					legacy[alias] = migrationGitSource(name, file)
				}
			}
		}
		legacyDependencies[name] = legacy
		current, err := migrationRootNamespace(dependency)
		if err != nil {
			return nil, fmt.Errorf("migrationRootNamespace: %w", err)
		}
		currentDependencies[name] = make(map[string]migrationProtoSource)
		for filename, file := range current {
			currentDependencies[name][filename] = migrationGitSource(name, file)
		}
	}
	for _, entry := range proof.entries {
		legacy := migrationImportNamespace{primary: legacyLocal}
		current := migrationImportNamespace{primary: currentLocal}
		var targets []string
		if selection, remote := proof.selections[entry.Module]; remote {
			dependency := proof.fetched[entry.Module]
			if dependency.Inspection == nil {
				continue
			}
			current.primary = []map[string]migrationProtoSource{currentDependencies[entry.Module]}
			for installed := range dependency.Inspection.LegacyFiles {
				if !slices.ContainsFunc(selection.subdirs, func(subdir string) bool {
					return subdir == "." || installed == subdir || strings.HasPrefix(installed, subdir+"/")
				}) {
					continue
				}
				name := installed
				if selection.root != "." {
					name = strings.TrimPrefix(installed, selection.root+"/")
				}
				targets = append(targets, name)
			}
		} else {
			targets = slices.Sorted(maps.Keys(p.sources))
		}
		for _, name := range slices.Sorted(maps.Keys(legacyDependencies)) {
			legacy.dependencies = append(legacy.dependencies, legacyDependencies[name])
			if name != entry.Module {
				current.dependencies = append(current.dependencies, currentDependencies[name])
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

func migrationGitSource(module string, file modules.RootProtoFile) migrationProtoSource {
	return migrationProtoSource{module: module, path: file.Path, identity: file.Identity, content: file.Content}
}

func (p *Plan) localImportNamespaces(native bool) ([]map[string]migrationProtoSource, error) {
	var result []map[string]migrationProtoSource
	for _, root := range p.roots {
		importRoot := filepath.Join(p.tx.requestedRoot, root)
		files := make(map[string]string)
		if err := collectProto(p.tx.requestedRoot, importRoot, importRoot, native, files); err != nil {
			return nil, fmt.Errorf("collectProto: %w", err)
		}
		namespace := make(map[string]migrationProtoSource)
		for name, logical := range files {
			resolved, err := sourceview.ResolveLocal(context.Background(), p.tx.requestedRoot, logical)
			if err != nil {
				return nil, fmt.Errorf("ResolveLocal: %w", err)
			}
			content, err := sourceview.ReadLocal(context.Background(), p.tx.requestedRoot, logical)
			if err != nil {
				return nil, fmt.Errorf("ReadLocal: %w", err)
			}
			namespace[name] = migrationProtoSource{module: p.git.localName, path: logical, identity: p.git.localName + ":" + resolved.Path, content: content}
		}
		result = append(result, namespace)
	}
	return result, nil
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
