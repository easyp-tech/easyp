package api

import (
	"bytes"
	"fmt"
	"os"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

func readV1ProtoImports(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}
	return parseV1ProtoImports(path, raw)
}

func parseV1ProtoImports(path string, raw []byte) ([]string, error) {
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
