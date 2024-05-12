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

	var sourcePkgName string
	if len(protoInfo.Info.ProtoBody.Packages) > 0 {
		sourcePkgName = protoInfo.Info.ProtoBody.Packages[0].Name
	}
	instrParser := instructionParser{
		sourcePkgName: sourcePkgName,
	}

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

	checkImportUsed := func(key string) {
		instruction := instrParser.parse(key)
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

	// look for import used

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range msg.MessageBody.Fields {
			checkImportUsed(field.Type)

			// look for in options
			for _, rpcOption := range field.FieldOptions {
				checkImportUsed(rpcOption.OptionName)
			}
		}
	}

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			// look for in request
			checkImportUsed(rpc.RPCRequest.MessageType)
			// look for in response
			checkImportUsed(rpc.RPCResponse.MessageType)

			// look for in options
			for _, rpcOption := range rpc.Options {
				checkImportUsed(rpcOption.OptionName)
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

type instructionParser struct {
	sourcePkgName string
}

// parseInstruction parse input string and return its package name
// return empty string if passed input does not imported from another package
func (p instructionParser) parse(input string) instructionInfo {
	// check if there is brackets, and extract
	// (google.api.http) -> google.api.http
	// (buf.validate.field).string.uuid -> buf.validate.field
	// or pkg.FieldType -> pkg.FieldType
	iStart := strings.Index(input, "(")
	iEnd := strings.Index(input, ")")
	if iStart != -1 && iEnd != -1 {
		input = input[iStart+1 : iEnd]
	}

	idx := strings.LastIndex(input, ".")
	if idx <= 0 {
		return instructionInfo{
			pkgName:     p.sourcePkgName,
			instruction: input,
		}
	}

	return instructionInfo{
		pkgName:     input[:idx],
		instruction: input[idx+1:],
	}
}

// existInProto look for key in proto file
func existInProto(key string, proto *unordered.Proto) bool {
	// look for key in extends
	for _, extend := range proto.ProtoBody.Extends {
		for _, field := range extend.ExtendBody.Fields {
			if field.FieldName == key {
				return true
			}
		}
	}

	// look for key in messages
	for _, message := range proto.ProtoBody.Messages {
		if message.MessageName == key {
			return true
		}
	}

	return false
}
