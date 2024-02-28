package rules

import (
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ImportUsed)(nil)

// ImportUsed this rule checks that all the imports declared across your Protobuf files are actually used.
type ImportUsed struct{}

// Validate implements lint.Rule.
func (i ImportUsed) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	imports := make(map[string]bool)
	importInfo := make(map[string]*parser.Import)
	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		pkgName := filepath.Dir(imp.Location)
		pkgName = strings.Trim(pkgName, "\"")
		imports[pkgName] = false
		importInfo[pkgName] = imp
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range msg.MessageBody.Fields {
			replaced := strings.ReplaceAll(field.Type, ".", "/")
			key := filepath.Dir(replaced)
			key = strings.Trim(key, "\"")
			if _, ok := imports[key]; ok {
				imports[key] = true
			}
		}
	}

	for imp, used := range imports {
		if !used {
			res = append(res, BuildError(importInfo[imp].Meta.Pos, importInfo[imp].Location, lint.ErrImportIsNotUsed))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
