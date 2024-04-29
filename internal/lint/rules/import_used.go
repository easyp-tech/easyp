package rules

import (
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
		pkgName := i.formatImportName(imp.Location)
		imports[pkgName] = false
		importInfo[pkgName] = imp
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range msg.MessageBody.Fields {
			key := i.formatField(field.Type)
			if _, ok := imports[key]; ok {
				imports[key] = true
			}

			for i2 := range field.FieldOptions {
				key = i.formatOption(field.FieldOptions[i2].OptionName)
				if _, ok := imports[key]; ok {
					imports[key] = true
				}
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

func (i ImportUsed) formatImportName(input string) string {
	importName := strings.Trim(input, "\"")
	importName = strings.ToLower(importName)

	return importName
}

func (i ImportUsed) formatField(input string) string {
	// replacing . with / to match the import name
	field := strings.ReplaceAll(input, ".", "/")
	field = strings.Trim(field, "\"")
	field = strings.ToLower(field)
	// adding .proto to match the import name file
	field += ".proto"

	return field
}

func (i ImportUsed) formatOption(input string) string {
	// removing the parenthesis from option
	option := strings.ReplaceAll(input, "(", "")
	option = strings.ReplaceAll(option, ")", "")
	// replacing . with / to match the import name
	option = strings.ReplaceAll(option, ".", "/")
	option = strings.Trim(option, "\"")
	option = strings.ToLower(option)
	// adding .proto to match the import name file
	option += ".proto"

	return option

}
