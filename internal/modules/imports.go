package modules

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"

	"github.com/easyp-tech/easyp/wellknownimports"
)

// ReadProtoImports reads and parses import declarations in a proto file.
func ReadProtoImports(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}
	return ParseProtoImports(path, raw)
}

// ParseProtoImports follows protobuf syntax, including comments and escaped paths.
func ParseProtoImports(path string, raw []byte) ([]string, error) {
	file, err := parser.Parse(path, bytes.NewReader(raw), reporter.NewHandler(nil))
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}
	var imports []string
	for _, declaration := range file.Decls {
		if imported, ok := declaration.(*ast.ImportNode); ok {
			imports = append(imports, imported.Name.AsString())
		}
	}
	return imports, nil
}

func unresolvedV1Imports(moduleDir string, roots []string) ([]string, error) {
	return unresolvedV1ImportsWithRoots(moduleDir, roots, nil)
}

func unresolvedV1ImportsWithRoots(moduleDir string, roots, dependencyRoots []string) ([]string, error) {
	unresolved, err := findUnresolvedV1Imports(moduleDir, roots, dependencyRoots)
	if err != nil {
		return nil, fmt.Errorf("findUnresolvedV1Imports: %w", err)
	}
	missing := map[string]struct{}{}
	for _, imported := range unresolved {
		missing[imported.path] = struct{}{}
	}
	paths := make([]string, 0, len(missing))
	for path := range missing {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths, nil
}

// ValidProtoImportPath reports whether an import is a portable relative file
// path without empty, dot, or parent components.
func ValidProtoImportPath(path string) bool {
	return path != "." && fs.ValidPath(path) && !strings.ContainsAny(path, "\\:\x00")
}

type v1UnresolvedImport struct {
	owner string
	path  string
}

func (i v1UnresolvedImport) String() string {
	return fmt.Sprintf("%s imports %q", i.owner, i.path)
}

type v1ImportSource struct {
	path    string
	builtin bool
}

func (s v1ImportSource) imports() ([]string, error) {
	if !s.builtin {
		return ReadProtoImports(s.path)
	}
	raw, err := wellknownimports.Content.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}
	return ParseProtoImports(s.path, raw)
}

// findUnresolvedV1Imports parses root sources and only the dependency files
// reachable from them. Unrelated dependency protos need not be valid or complete.
func findUnresolvedV1Imports(moduleDir string, roots, dependencyRoots []string) ([]v1UnresolvedImport, error) {
	sources := make(SourceRoots, 0, len(dependencyRoots))
	for _, root := range dependencyRoots {
		sources = append(sources, SourceRoot{Path: root})
	}
	return findUnresolvedV1ImportsWithSources(moduleDir, roots, sources)
}

func findUnresolvedV1ImportsWithSources(moduleDir string, roots []string, dependencyRoots SourceRoots) ([]v1UnresolvedImport, error) {
	allRoots := make(SourceRoots, 0, len(roots)+len(dependencyRoots))
	for _, root := range roots {
		allRoots = append(allRoots, SourceRoot{Path: filepath.Join(moduleDir, root)})
	}
	allRoots = append(allRoots, dependencyRoots...)
	allowed := allRoots.FileAllowed()
	var queue []v1ImportSource
	seen := make(map[v1ImportSource]bool)
	for _, sourceRoot := range allRoots[:len(roots)] {
		err := sourceRoot.Walk(func(path string) error {
			if allowed != nil && !allowed(path) {
				return nil
			}
			source := v1ImportSource{path: path}
			if !seen[source] {
				seen[source] = true
				queue = append(queue, source)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	var unresolved []v1UnresolvedImport
	for len(queue) > 0 {
		source := queue[0]
		queue = queue[1:]
		imports, err := source.imports()
		if err != nil {
			return nil, fmt.Errorf("imports: %w", err)
		}
		seenImports := make(map[string]bool, len(imports))
		for _, importPath := range imports {
			if !ValidProtoImportPath(importPath) {
				return nil, fmt.Errorf("%s: invalid import %q", source.path, importPath)
			}
			if seenImports[importPath] {
				continue
			}
			seenImports[importPath] = true
			imported, err := resolveV1ImportSource(importPath, allRoots, allowed)
			if errors.Is(err, os.ErrNotExist) {
				unresolved = append(unresolved, v1UnresolvedImport{owner: source.path, path: importPath})
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("resolveV1ImportSource: %s imports %q: %w", source.path, importPath, err)
			}
			if !seen[imported] {
				seen[imported] = true
				queue = append(queue, imported)
			}
		}
	}
	return unresolved, nil
}

func resolveV1ImportSource(importPath string, roots SourceRoots, allowed func(string) bool) (v1ImportSource, error) {
	for _, root := range roots {
		candidate := filepath.Join(root.Path, filepath.FromSlash(importPath))
		if allowed != nil {
			// The shared filter checks both owners. This root still needs to own
			// the import's spelling, but a selected target may belong to another root.
			if !allowed(candidate) || !root.allowsPath(candidate, root.Path) {
				continue
			}
		} else if !root.allows(candidate, root.Path) {
			continue
		}
		info, err := os.Stat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return v1ImportSource{}, fmt.Errorf("Stat: %w", err)
		}
		if info.Mode().IsRegular() {
			return v1ImportSource{path: candidate}, nil
		}
	}
	info, err := fs.Stat(wellknownimports.Content, importPath)
	if err != nil {
		return v1ImportSource{}, fmt.Errorf("Stat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return v1ImportSource{}, os.ErrNotExist
	}
	return v1ImportSource{path: importPath, builtin: true}, nil
}
