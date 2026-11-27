package modules

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"

	"github.com/easyp-tech/easyprotoc/protoc/syntax"
)

type importLiteralEdit struct {
	start int
	end   int
	value string
}

func rewriteProtoImportLiterals(path string, raw []byte, replacement func(string) (string, error)) ([]byte, [][2]string, error) {
	file, err := syntax.Parse(path, raw)
	if err != nil {
		return nil, nil, fmt.Errorf("Parse: %w", err)
	}
	var edits []importLiteralEdit
	var mappings [][2]string
	for _, imported := range file.AST.Imports {
		before := file.Text(imported.Path)
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
		// The path may be split into adjacent literals; the first one takes
		// the new path and the rest become empty, keeping comments between.
		first := true
		for i := file.TokenAt(imported.Span.Start); i < file.NumTokens(); i++ {
			kind, start, end := file.Token(i)
			if start >= imported.Span.End {
				break
			}
			if kind != syntax.KindString {
				continue
			}
			value := ""
			if first {
				value, first = after, false
			}
			edits = append(edits, importLiteralEdit{start: int(start), end: int(end), value: strconv.Quote(value)})
		}
		if first {
			return nil, nil, fmt.Errorf("unsupported import literal in %s", path)
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
