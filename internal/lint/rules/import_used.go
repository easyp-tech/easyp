package rules

import (
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
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
		importPath := lint.ConvertImportPath(imp.Location)
		isImportUsed[importPath] = false
		importInfo[importPath] = imp
	}

	// look for import used

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range msg.MessageBody.Fields {
			// fieldType := field.Type

			// key := lint.ImportPath(i.formatField(field.Type))
			key := lint.ImportPath(field.Type)
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
				instruction := parseInstruction(optionName)

				// look for option in imported files
				for _, importPath := range pkgToImport[instruction.pkgName] {
					proto := protoInfo.ProtoFilesFromImport[importPath]
					exist := existInProto(instruction.instruction, proto)

					if exist {
						if _, ok := isImportUsed[importPath]; ok {
							isImportUsed[importPath] = true
						}
					}
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

// instructionInfo collects info about instruction in proto file
// e.g `google.api.http`:
// 		`google.api` - package name
// 		'http' - instruction name
type instructionInfo struct {
	pkgName     string
	instruction string
}

// formatOptionName trims '(' and ')' from option
// long story short: convert '(google.api.http)' to 'google.api.http'
func formatOptionName(input string) string {
	// removing the parenthesis from option
	option := strings.ReplaceAll(input, "(", "")
	option = strings.ReplaceAll(option, ")", "")

	return option
}

// parseInstruction parse input string and return its package name
// return empty string if passed input does not imported from another package
func parseInstruction(input string) instructionInfo {
	idx := strings.LastIndex(input, ".")
	if idx <= 0 {
		return instructionInfo{instruction: input}
	}

	return instructionInfo{
		pkgName:     input[:idx],
		instruction: input[idx+1:],
	}
}

// existInProto look for key in proto file
func existInProto(key string, proto *unordered.Proto) bool {
	// look for in extends
	for _, extend := range proto.ProtoBody.Extends {
		for _, field := range extend.ExtendBody.Fields {
			if field.FieldName == key {
				return true
			}
		}
	}

	return false
}
