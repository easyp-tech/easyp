package rules

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/api/factories"
	modulereflect "github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ImportUsed)(nil)

// ImportUsed this rule checks that all the imports declared across your Protobuf files are actually used.
type ImportUsed struct {
	deps          []string // FIXME: passed deps is tmp solution!
	moduleReflect *modulereflect.ModuleReflect
}

// Validate implements lint.Rule.
func (i *ImportUsed) Validate(protoInfo lint.ProtoInfo) []error {
	// TODO: tmp solution! only for PoC!
	var err error
	i.moduleReflect, err = factories.NewModuleReflect()
	if err != nil {
		panic(err)
	}

	var res []error

	// collects flags if import was used
	isImportUsed := make(map[string]bool)

	// information about imports (used for linter error message)
	importsInfo := make(map[string]*parser.Import)

	// imported proto files
	importsProto := make(map[string]*unordered.Proto)

	// package -> import name
	pkgProto := make(map[string]string)

	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		importName := i.formatImportName(imp.Location)
		isImportUsed[importName] = false
		importsInfo[importName] = imp
		importedProto := i.readImportProtoFile(importName)
		importsProto[importName] = importedProto
		if len(importedProto.ProtoBody.Packages) == 0 {
			continue
		}
		pkgName := importedProto.ProtoBody.Packages[0].Name
		pkgProto[pkgName] = importName
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range msg.MessageBody.Fields {
			key := i.formatField(field.Type)
			if _, ok := isImportUsed[key]; ok {
				isImportUsed[key] = true
			}

			for i2 := range field.FieldOptions {
				key = i.formatOption(field.FieldOptions[i2].OptionName)
				if _, ok := isImportUsed[key]; ok {
					isImportUsed[key] = true
				}

			}
		}
	}

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			for _, rpcOption := range rpc.Options {
				key := i.formatOption(rpcOption.OptionName)
				key = i.rpcOptionExclusion(key, rpcOption.Constant)
				if _, ok := isImportUsed[key]; ok {
					isImportUsed[key] = true
				}
			}
		}
	}

	for imp, used := range isImportUsed {
		if !used {
			res = append(res, BuildError(importsInfo[imp].Meta.Pos, importsInfo[imp].Location, lint.ErrImportIsNotUsed))
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

func (i ImportUsed) rpcOptionExclusion(input string, constant string) string {
	switch input {
	case "google/api/http.proto":
		for _, method := range []string{"get", "post", "put", "delete", "patch"} {
			if strings.Contains(constant, method) {
				return "google/api/annotations.proto"
			}
		}

		return input
	default:
		return input
	}
}

// readImportProtoFile reads imported proto file
// TODO: need to read from local imports
func (i ImportUsed) readImportProtoFile(importName string) *unordered.Proto {
	ctx := context.TODO()

	// try to find actual module
	for _, dep := range i.deps {
		modulepath, err := i.moduleReflect.GetModulePath(ctx, dep)
		if err != nil {
			panic(err)
		}

		// if file exists that it's the module that we are looking for
		fullPath := filepath.Join(modulepath, importName)
		fp, err := os.Open(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			panic(err)
		}

		got, err := protoparser.Parse(fp)
		if err != nil {
			panic(err)
		}
		_ = fp.Close()

		proto, err := unordered.InterpretProto(got)
		if err != nil {
			panic(err)
		}

		return proto
	}

	return nil
}
