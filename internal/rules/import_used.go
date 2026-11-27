package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*ImportUsed)(nil)

// ImportUsed this rule checks that all the imports declared across your Protobuf files are actually used.
// It keeps no state between files, so files can be checked concurrently.
type ImportUsed struct{}

// importUsage tracks, for one file without resolved names, which imports its
// written names point to.
type importUsage struct {
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
	if checkingProto.Resolved != nil {
		return i.validateResolved(checkingProto), nil
	}

	var res []core.Issue
	u := &importUsage{}

	u.instrParser = core.InstructionParser{
		SourcePkgName: core.PackageName(checkingProto.File.PackageName()),
	}

	// collects flags if import was used
	u.isImportUsed = make(map[core.ImportPath]bool)

	// collects pkg name -> import path
	u.pkgToImport = make(map[core.PackageName][]core.ImportPath)
	for importPath, proto := range checkingProto.ImportedFiles {
		pkgName := core.PackageName(proto.PackageName())
		if pkgName == "" {
			// skip if package is omitted
			continue
		}

		u.pkgToImport[pkgName] = append(u.pkgToImport[pkgName], importPath)
	}

	// collects info about import in linted proto file
	importInfo := make(map[core.ImportPath]*protoast.Import)
	for _, imp := range checkingProto.File.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		u.isImportUsed[importPath] = false
		importInfo[importPath] = imp
	}

	// look for import used
	u.checkInServices(checkingProto.File.Services, checkingProto)
	u.checkInMessages(checkingProto.File.Messages, checkingProto)
	u.checkInExtends(checkingProto.File.Extends, checkingProto)
	u.checkInOptions(checkingProto.File.Options, checkingProto)

	for imp, used := range u.isImportUsed {
		if !used {
			res = core.AppendIssue(res, i, importInfo[imp].Pos, importInfo[imp].Location)
		}
	}

	return res, nil
}

// validateResolved reports non-public imports that no resolved name in the file
// comes from, directly or through the import's public imports.
func (i *ImportUsed) validateResolved(checkingProto core.ProtoInfo) []core.Issue {
	var res []core.Issue
	for _, imp := range checkingProto.File.Imports {
		if imp.Public || checkingProto.Resolved.UsedImports[core.ImportPath(imp.Path)] {
			continue
		}
		res = core.AppendIssue(res, i, imp.Pos, imp.Location)
	}
	return res
}

// checkIsImportUsed check if passed import is used in proto file
func (u *importUsage) checkIsImportUsed(key string, checkingProto core.ProtoInfo) {
	instruction := u.instrParser.Parse(key)
	for _, importPath := range u.pkgToImport[instruction.PkgName] {
		proto := checkingProto.ImportedFiles[importPath]
		exist := existInProto(instruction.Instruction, proto)

		if exist {
			if _, ok := u.isImportUsed[importPath]; ok {
				u.isImportUsed[importPath] = true
			}
		}
	}
}

// check used imports in services
func (u *importUsage) checkInServices(services []*protoast.Service, checkingProto core.ProtoInfo) {
	for _, service := range services {
		for _, rpc := range service.RPCs {
			// look for in request
			u.checkIsImportUsed(rpc.Request.Type, checkingProto)

			// look for in response
			u.checkIsImportUsed(rpc.Response.Type, checkingProto)

			// look for in options
			for _, rpcOption := range rpc.Options {
				u.checkIsImportUsed(rpcOption.Name, checkingProto)
			}
		}
	}
}

// check used imports in messages
func (u *importUsage) checkInMessages(messages []*protoast.Message, checkingProto core.ProtoInfo) {
	for _, msg := range messages {
		u.checkInMessages(msg.Messages, checkingProto)

		for _, field := range msg.Fields {
			// look for field's type in imported files
			u.checkIsImportUsed(field.Type, checkingProto)

			// look for field's options in imported files
			for _, fieldOption := range field.Options {
				u.checkIsImportUsed(fieldOption.Name, checkingProto)
			}
		}

		for _, oneOf := range msg.Oneofs {
			for _, field := range oneOf.Fields {
				u.checkIsImportUsed(field.Type, checkingProto)
			}
		}
	}
}

func (u *importUsage) checkInExtends(extends []*protoast.Extend, checkingProto core.ProtoInfo) {
	for _, extend := range extends {
		u.checkIsImportUsed(extend.Type, checkingProto)
	}
}

func (u *importUsage) checkInOptions(options []*protoast.Option, checkingProto core.ProtoInfo) {
	for _, option := range options {
		u.checkIsImportUsed(option.Name, checkingProto)
	}
}

// existInProto look for key in imported proto file
// look for used instruction (key) in imported proto file
func existInProto(key string, proto *protoast.File) bool {
	// look for key in extends
	for _, extend := range proto.Extends {
		for _, field := range extend.Fields {
			if field.Name == key {
				return true
			}
		}
	}

	// look for key in messages
	for _, message := range proto.Messages {
		if message.Name == key {
			return true
		}
	}

	// look for key in enum
	for _, enum := range proto.Enums {
		if enum.Name == key {
			return true
		}
	}

	return false
}
