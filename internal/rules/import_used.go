package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportUsed)(nil)

// ImportUsed this rule checks that all the imports declared across your Protobuf files are actually used.
type ImportUsed struct {
	instrParser  core.InstructionParser
	isImportUsed map[core.ImportPath]bool
	pkgToImport  map[core.PackageName][]core.ImportPath
}

// Message implements lint.Rule.
func (i *ImportUsed) Message() string {
	return "import is not used"
}

// Validate implements core.Rule.
func (i *ImportUsed) Validate(checkingProto core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	i.instrParser = core.InstructionParser{
		SourcePkgName: core.GetPackageName(checkingProto.Info),
	}

	// collects flags if import was used
	i.isImportUsed = make(map[core.ImportPath]bool)

	// collects pkg name -> import path
	i.pkgToImport = make(map[core.PackageName][]core.ImportPath)
	for importPath, proto := range checkingProto.ProtoFilesFromImport {
		pkgName := core.GetPackageName(proto)
		if pkgName == "" {
			// skip if package is omitted
			continue
		}

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
	i.checkInServices(checkingProto.Info.ProtoBody.Services, checkingProto)
	i.checkInMessages(checkingProto.Info.ProtoBody.Messages, checkingProto)
	i.checkInExtends(checkingProto.Info.ProtoBody.Extends, checkingProto)
	i.checkInOptions(checkingProto.Info.ProtoBody.Options, checkingProto)

	for imp, used := range i.isImportUsed {
		if !used {
			res = core.AppendIssue(res, i, importInfo[imp].Meta.Pos, importInfo[imp].Location, importInfo[imp].Comments)
		}
	}

	return res, nil
}

// checkIsImportUsed check if passed import is used in proto file
func (i *ImportUsed) checkIsImportUsed(key string, checkingProto core.ProtoInfo) {
	instruction := i.instrParser.Parse(key)
	for _, importPath := range i.pkgToImport[instruction.PkgName] {
		proto := checkingProto.ProtoFilesFromImport[importPath]
		exist := existInProto(instruction.Instruction, proto)

		if exist {
			if _, ok := i.isImportUsed[importPath]; ok {
				i.isImportUsed[importPath] = true
			}
		}
	}
}

// check used imports in services
func (i *ImportUsed) checkInServices(services []*unordered.Service, checkingProto core.ProtoInfo) {
	for _, service := range services {
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
}

// check used imports in messages
func (i *ImportUsed) checkInMessages(messages []*unordered.Message, checkingProto core.ProtoInfo) {
	for _, msg := range messages {
		i.checkInMessages(msg.MessageBody.Messages, checkingProto)

		for _, field := range msg.MessageBody.Fields {
			// look for field's type in imported files
			i.checkIsImportUsed(field.Type, checkingProto)

			// look for field's options in imported files
			for _, fieldOption := range field.FieldOptions {
				i.checkIsImportUsed(fieldOption.OptionName, checkingProto)
			}
		}

		for _, oneOf := range msg.MessageBody.Oneofs {
			for _, field := range oneOf.OneofFields {
				i.checkIsImportUsed(field.Type, checkingProto)
			}
		}
	}
}

func (i *ImportUsed) checkInExtends(extends []*unordered.Extend, checkingProto core.ProtoInfo) {
	for _, extend := range extends {
		i.checkIsImportUsed(extend.MessageType, checkingProto)
	}
}

func (i *ImportUsed) checkInOptions(options []*parser.Option, checkingProto core.ProtoInfo) {
	for _, option := range options {
		i.checkIsImportUsed(option.OptionName, checkingProto)
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
