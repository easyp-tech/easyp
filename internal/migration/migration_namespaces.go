package migration

import (
	"context"
	"fmt"
	"maps"
	"path/filepath"
	"slices"

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

func observeLocalImportNamespaces(rootDir, module string, roots []string, native bool) ([]map[string]migrationProtoSource, error) {
	var result []map[string]migrationProtoSource
	for _, root := range roots {
		importRoot := filepath.Join(rootDir, root)
		files := make(map[string]string)
		if err := collectProto(rootDir, importRoot, importRoot, native, files); err != nil {
			return nil, fmt.Errorf("collectProto: %w", err)
		}
		namespace := make(map[string]migrationProtoSource)
		for name, logical := range files {
			resolved, err := sourceview.ResolveLocal(context.Background(), rootDir, logical)
			if err != nil {
				return nil, fmt.Errorf("ResolveLocal: %w", err)
			}
			content, err := sourceview.ReadLocal(context.Background(), rootDir, logical)
			if err != nil {
				return nil, fmt.Errorf("ReadLocal: %w", err)
			}
			namespace[name] = migrationProtoSource{module: module, path: logical, identity: module + ":" + resolved.Path, content: content}
		}
		result = append(result, namespace)
	}
	return result, nil
}

// migrationNamespaces holds complete available namespaces. Target selectors are
// applied by the binding proof; collision checks must still see unselected files.
type migrationNamespaces struct {
	legacyLocal, currentLocal               []map[string]migrationProtoSource
	legacyDependencies, currentDependencies map[string]map[string]migrationProtoSource
}

func observeMigrationNamespaces(localRoot string, local *localSelectionProof, proof *gitSelectionProof) (migrationNamespaces, error) {
	legacyLocal, err := observeLocalImportNamespaces(localRoot, local.module, local.roots, false)
	if err != nil {
		return migrationNamespaces{}, fmt.Errorf("observeLocalImportNamespaces: %w", err)
	}
	// Released v0 reverses its collected import roots before compilation.
	slices.Reverse(legacyLocal)
	currentLocal, err := observeLocalImportNamespaces(localRoot, local.module, local.roots, true)
	if err != nil {
		return migrationNamespaces{}, fmt.Errorf("observeLocalImportNamespaces: %w", err)
	}
	legacyDependencies, currentDependencies, err := gitImportNamespaces(proof)
	if err != nil {
		return migrationNamespaces{}, fmt.Errorf("gitImportNamespaces: %w", err)
	}
	return migrationNamespaces{
		legacyLocal: legacyLocal, currentLocal: currentLocal,
		legacyDependencies: legacyDependencies, currentDependencies: currentDependencies,
	}, nil
}
