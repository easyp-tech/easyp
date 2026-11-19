package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.LintRule = (*ImportUsedLinked)(nil)

// ImportUsedLinked is the linked-engine port of ImportUsed (research inventory
// #12, adapt + parity-risk, threat T-04-08): main's syntax-walk semantics are
// ported UNCONDITIONALLY — LinkedFileInfo.Resolved is never read, and the
// lab's resolution-based validateResolved is a deferred post-parity decision,
// never silently adopted (grep-enforced by the batch gate). The etalon's
// stateless importUsage local struct is kept, so files can be checked
// concurrently; the walk mirrors main's checkIn* mechanics over protoast
// fields:
//
//   - services: rpc request and response types, rpc option names;
//   - messages, RECURSING into nested messages exactly like main's
//     checkInMessages: plain field types, field option names, oneof member
//     field types (map, group and message-level option names are outside
//     main's walk — the unordered model lists them separately and main never
//     touches them);
//   - FILE-LEVEL extends only: main's checkInExtends receives
//     ProtoBody.Extends, and the unordered model keeps message-level extends
//     inside MessageBody.Extends, so they are outside main's walk too;
//   - file-level option names.
//
// existInProtoLinked keeps main's reading of the imported file: the extend
// FIELDS' names, the top-level message names, the top-level enum names — a
// qualified reference into an imported nested message does not match and
// leaves the import unused (pinned by the syntax_walk_only parity cell).
//
// Determinism note (D-01): main reported unused imports through a map
// iteration; this port iterates info.File.Imports in source order and reports
// each distinct unused import path once, at its first written occurrence (a
// duplicated import statement is rejected by protoc, so first-vs-last
// occurrence never differs on a compiling corpus). The engine's D-01 sort
// normalizes order either way.
type ImportUsedLinked struct{}

// importUsage tracks, for one file, which of its written names point into
// which import (main's receiver state extracted per call — the etalon's
// stateless shape).
type importUsage struct {
	instrParser  core.InstructionParser
	isImportUsed map[core.ImportPath]bool
	pkgToImport  map[core.PackageName][]core.ImportPath
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (i *ImportUsedLinked) LintRuleName() string {
	return "IMPORT_USED"
}

// Message implements lint.LintRule.
func (i *ImportUsedLinked) Message() string {
	return "import is not used"
}

// Validate implements lint.LintRule.
func (i *ImportUsedLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	u := &importUsage{}

	u.instrParser = core.InstructionParser{
		SourcePkgName: core.PackageName(info.File.PackageName()),
	}

	// collects flags if import was used
	u.isImportUsed = make(map[core.ImportPath]bool)

	// collects pkg name -> import path
	u.pkgToImport = make(map[core.PackageName][]core.ImportPath)
	for importPath, proto := range info.ImportedFiles {
		pkgName := core.PackageName(proto.PackageName())
		if pkgName == "" {
			// skip if package is omitted
			continue
		}

		u.pkgToImport[pkgName] = append(u.pkgToImport[pkgName], importPath)
	}

	// collects info about import in linted proto file
	for _, imp := range info.File.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		u.isImportUsed[importPath] = false
	}

	// look for import used
	u.checkInServices(info.File.Services, info)
	u.checkInMessages(info.File.Messages, info)
	u.checkInExtends(info.File.Extends, info)
	u.checkInOptions(info.File.Options, info)

	// report every import with used == false, at the import keyword with the
	// quoted location as the source (main's imp.Meta.Pos ≡ imp.Pos and
	// imp.Location basis).
	reported := make(map[core.ImportPath]bool, len(u.isImportUsed))
	for _, imp := range info.File.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		if reported[importPath] || u.isImportUsed[importPath] {
			continue
		}
		reported[importPath] = true
		res = core.AppendLinkedIssue(res, i, imp.Pos, imp.Location)
	}

	return res, nil
}

// checkIsImportUsed check if passed import is used in proto file
func (u *importUsage) checkIsImportUsed(key string, info core.LinkedFileInfo) {
	instruction := u.instrParser.Parse(key)
	for _, importPath := range u.pkgToImport[instruction.PkgName] {
		proto := info.ImportedFiles[importPath]
		exist := existInProtoLinked(instruction.Instruction, proto)

		if exist {
			if _, ok := u.isImportUsed[importPath]; ok {
				u.isImportUsed[importPath] = true
			}
		}
	}
}

// check used imports in services
func (u *importUsage) checkInServices(services []*protoast.Service, info core.LinkedFileInfo) {
	for _, service := range services {
		for _, rpc := range service.RPCs {
			// look for in request
			u.checkIsImportUsed(rpc.Request.Type, info)

			// look for in response
			u.checkIsImportUsed(rpc.Response.Type, info)

			// look for in options
			for _, rpcOption := range rpc.Options {
				u.checkIsImportUsed(rpcOption.Name, info)
			}
		}
	}
}

// check used imports in messages
func (u *importUsage) checkInMessages(messages []*protoast.Message, info core.LinkedFileInfo) {
	for _, msg := range messages {
		u.checkInMessages(msg.Messages, info)

		for _, field := range msg.Fields {
			// look for field's type in imported files
			u.checkIsImportUsed(field.Type, info)

			// look for field's options in imported files
			for _, fieldOption := range field.Options {
				u.checkIsImportUsed(fieldOption.Name, info)
			}
		}

		for _, oneOf := range msg.Oneofs {
			for _, field := range oneOf.Fields {
				u.checkIsImportUsed(field.Type, info)
			}
		}
	}
}

func (u *importUsage) checkInExtends(extends []*protoast.Extend, info core.LinkedFileInfo) {
	for _, extend := range extends {
		u.checkIsImportUsed(extend.Type, info)
	}
}

func (u *importUsage) checkInOptions(options []*protoast.Option, info core.LinkedFileInfo) {
	for _, option := range options {
		u.checkIsImportUsed(option.Name, info)
	}
}

// existInProtoLinked looks for key in imported proto file — the etalon named
// it existInProto; the name is taken in this package by main's byte-frozen
// import_used.go, so the port carries the Linked suffix (rename map). The
// reading is main's: the extend fields, the top-level message names and the
// top-level enum names.
func existInProtoLinked(key string, proto *protoast.File) bool {
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
