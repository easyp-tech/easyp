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

	// collects flags if import was used
	isImportUsed := make(map[lint.ImportPath]bool)

	// collects pkg name -> import path
	pkgToImport := make(map[string][]lint.ImportPath)
	for importPath, proto := range protoInfo.ProtoFilesFromImport {
		if len(proto.ProtoBody.Packages) == 0 {
			// skip if package is omitted
			continue
		}

		pkgName := proto.ProtoBody.Packages[0].Name
		pkgToImport[pkgName] = append(pkgToImport[pkgName], importPath)
	}

	// collects info about import in linted proto file
	importInfo := make(map[lint.ImportPath]*parser.Import)
	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		importPath := lint.ImportPath(imp.Location)
		isImportUsed[importPath] = false
		importInfo[importPath] = imp
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range msg.MessageBody.Fields {
			key := lint.ImportPath(i.formatField(field.Type))
			if _, ok := isImportUsed[key]; ok {
				isImportUsed[key] = true
			}

			for i2 := range field.FieldOptions {
				key = lint.ImportPath(formatOptionName(field.FieldOptions[i2].OptionName))
				if _, ok := isImportUsed[key]; ok {
					isImportUsed[key] = true
				}

			}
		}
	}

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			for _, rpcOption := range rpc.Options {
				optionName := formatOptionName(rpcOption.OptionName)
				_ = optionName
				key := lint.ImportPath(formatOptionName(rpcOption.OptionName))
				if _, ok := isImportUsed[key]; ok {
					isImportUsed[key] = true
				}
			}
		}
	}

	for imp, used := range isImportUsed {
		if !used {
			res = append(res, BuildError(importInfo[imp].Meta.Pos, importInfo[imp].Location, lint.ErrImportIsNotUsed))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}

func (i ImportUsed) formatField(input string) string {
	// field := strings.Trim(input, "\"")
	// field = strings.ToLower(field)

	// return field
	return input
}

// formatOptionName trims '(' and ')' from option
// long story short: convert '(google.api.http)' to 'google.api.http'
func formatOptionName(input string) string {
	// removing the parenthesis from option
	option := strings.ReplaceAll(input, "(", "")
	option = strings.ReplaceAll(option, ")", "")

	return option
}
