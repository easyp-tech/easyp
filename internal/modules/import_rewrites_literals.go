package modules

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

type importLiteralEdit struct {
	start int
	end   int
	value string
}

func rewriteProtoImportLiterals(path string, raw []byte, replacement func(string) (string, error)) ([]byte, [][2]string, error) {
	file, err := parser.Parse(path, bytes.NewReader(raw), reporter.NewHandler(nil))
	if err != nil {
		return nil, nil, fmt.Errorf("Parse: %w", err)
	}
	var edits []importLiteralEdit
	var mappings [][2]string
	for _, declaration := range file.Decls {
		imported, ok := declaration.(*ast.ImportNode)
		if !ok {
			continue
		}
		before := imported.Name.AsString()
		after, err := replacement(before)
		if err != nil {
			return nil, nil, fmt.Errorf("replacement: %w", err)
		}
		if before == after {
			continue
		}
		if !ValidProtoImportPath(after) {
			return nil, nil, fmt.Errorf("invalid replacement import %q", after)
		}
		var literals []*ast.StringLiteralNode
		switch name := imported.Name.(type) {
		case *ast.StringLiteralNode:
			literals = append(literals, name)
		case *ast.CompoundStringLiteralNode:
			for _, child := range name.Children() {
				literal, ok := child.(*ast.StringLiteralNode)
				if !ok {
					return nil, nil, fmt.Errorf("unsupported import literal in %s", path)
				}
				literals = append(literals, literal)
			}
		default:
			return nil, nil, fmt.Errorf("unsupported import literal in %s", path)
		}
		for index, literal := range literals {
			info := file.NodeInfo(literal)
			value := ""
			if index == 0 {
				value = after
			}
			start := info.Start().Offset
			edits = append(edits, importLiteralEdit{start: start, end: start + len(info.RawText()), value: strconv.Quote(value)})
		}
		mapping := [2]string{before, after}
		if !slices.Contains(mappings, mapping) {
			mappings = append(mappings, mapping)
		}
	}
	if len(edits) == 0 {
		return raw, nil, nil
	}
	var output bytes.Buffer
	at := 0
	for _, edit := range edits {
		output.Write(raw[at:edit.start])
		output.WriteString(edit.value)
		at = edit.end
	}
	output.Write(raw[at:])
	return output.Bytes(), mappings, nil
}
