package rules

import (
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportUsed)(nil)

// ImportUsed this rule checks that all the imports declared across your Protobuf files are actually used.
type ImportUsed struct {
	instrParser  instructionParser
	isImportUsed map[core.ImportPath]bool
	pkgToImport  map[string][]core.ImportPath
}

// Message implements lint.Rule.
func (i *ImportUsed) Message() string {
	return "import is not used"
}

// Validate implements lint.Rule.
func (i *ImportUsed) Validate(checkingProto core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	var sourcePkgName string
	if len(checkingProto.Info.ProtoBody.Packages) > 0 {
		sourcePkgName = checkingProto.Info.ProtoBody.Packages[0].Name
	}
	i.instrParser = instructionParser{
		sourcePkgName: sourcePkgName,
	}

	// collects flags if import was used
	i.isImportUsed = make(map[core.ImportPath]bool)

	// collects pkg name -> import path
	i.pkgToImport = make(map[string][]core.ImportPath)
	for importPath, proto := range checkingProto.ProtoFilesFromImport {
		if len(proto.ProtoBody.Packages) == 0 {
			// skip if package is omitted
			continue
		}

		pkgName := proto.ProtoBody.Packages[0].Name
		i.pkgToImport[pkgName] = append(i.pkgToImport[pkgName], importPath)
	}

	// collects info about import in linted proto file
	importInfo := make(map[core.ImportPath]*parser.Import)
	for _, imp := range checkingProto.Info.ProtoBody.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		i.isImportUsed[importPath] = false
		importInfo[importPath] = imp
	}

	// look for import used

	i.checkInExtends(checkingProto.Info.ProtoBody.Extends, checkingProto)
	i.checkMessages(checkingProto.Info.ProtoBody.Messages, checkingProto)

	for _, service := range checkingProto.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			// look for in request
			i.checkIsImportUsed(rpc.RPCRequest.MessageType, checkingProto)

			// look for in response
			i.checkIsImportUsed(rpc.RPCResponse.MessageType, checkingProto)

			// look for in options
			for _, rpcOption := range rpc.Options {
				i.checkIsImportUsed(rpcOption.OptionName, checkingProto)
			}
		}
	}

	for imp, used := range i.isImportUsed {
		if !used {
			res = core.AppendIssue(res, i, importInfo[imp].Meta.Pos, importInfo[imp].Location, importInfo[imp].Comments)
		}
	}

	return res, nil
}

// checkIsImportUsed check if passed import is used in proto file
func (i *ImportUsed) checkIsImportUsed(key string, checkingProto core.ProtoInfo) {
	instruction := i.instrParser.parse(key)
	for _, importPath := range i.pkgToImport[instruction.pkgName] {
		proto := checkingProto.ProtoFilesFromImport[importPath]
		exist := existInProto(instruction.instruction, proto)

		if exist {
			if _, ok := i.isImportUsed[importPath]; ok {
				i.isImportUsed[importPath] = true
			}
		}
	}
}

// check used imports in messages
func (i *ImportUsed) checkMessages(messages []*unordered.Message, checkingProto core.ProtoInfo) {
	for _, msg := range messages {
		i.checkMessages(msg.MessageBody.Messages, checkingProto)

		for _, field := range msg.MessageBody.Fields {
			// look for field's type in imported files
			i.checkIsImportUsed(field.Type, checkingProto)

			// look for field's options in imported files
			for _, fieldOption := range field.FieldOptions {
				i.checkIsImportUsed(fieldOption.OptionName, checkingProto)
			}
		}
	}
}

func (i *ImportUsed) checkInExtends(extends []*unordered.Extend, checkingProto core.ProtoInfo) {
	for _, extend := range extends {
		i.checkIsImportUsed(extend.MessageType, checkingProto)
	}
}

// instructionInfo collects info about instruction in proto file
// e.g `google.api.http`:
//
//	`google.api` - package name
//	'http' - instruction name
type instructionInfo struct {
	pkgName     string
	instruction string
}

type instructionParser struct {
	sourcePkgName string
}

// parseInstruction parse input string and return its package name
// if passed input does not have package -> return pkgName as package name source proto file
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

// existInProto look for key in imported proto file
// look for used instruction (key) in imported proto file
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

	// look for key in enum
	for _, enum := range proto.ProtoBody.Enums {
		if enum.EnumName == key {
			return true
		}
	}

	return false
}
