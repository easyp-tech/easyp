package modules

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
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
	missing := map[string]struct{}{}
	allRoots := make([]string, 0, len(roots)+len(dependencyRoots))
	for _, root := range roots {
		allRoots = append(allRoots, filepath.Join(moduleDir, root))
	}
	allRoots = append(allRoots, dependencyRoots...)
	for _, root := range roots {
		sourceRoot := filepath.Join(moduleDir, root)
		err := WalkProtoFiles(sourceRoot, func(path string) error {
			imports, err := ReadProtoImports(path)
			if err != nil {
				return fmt.Errorf("ReadProtoImports: %w", err)
			}
			for _, importPath := range imports {
				if strings.HasPrefix(importPath, "google/protobuf/") {
					continue
				}
				found := false
				for _, candidateRoot := range allRoots {
					candidate := filepath.Join(candidateRoot, importPath)
					info, err := os.Stat(candidate)
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					if err != nil {
						return fmt.Errorf("stat import %s: %w", importPath, err)
					}
					if info.Mode().IsRegular() {
						found = true
						break
					}
				}
				if !found {
					missing[importPath] = struct{}{}
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	paths := make([]string, 0, len(missing))
	for path := range missing {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths, nil
}
