package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*ImportUsedLinked)(nil)

// ImportUsedLinked is the linked-engine port of ImportUsed (research inventory
// #12, adapt + parity-risk, threat T-04-08): main's syntax-walk semantics are
// ported UNCONDITIONALLY — ProtoInfo.Resolved is never read, and the
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
// existInProto keeps main's reading of the imported file: the extend
// FIELDS' names, the top-level message names, the top-level enum names — a
// qualified reference into an imported nested message does not match and
// leaves the import unused (pinned by the syntax_walk_only parity cell).
//
// Determinism note (D-01): main reported unused imports through a map
// iteration; this port reports each distinct unused import path once, at its
// LAST written occurrence — main's importInfo overwrite map keyed by
// ImportPath. The linked engine itself accepts duplicated import statements
// and runs rules on them (no COMPILE issues), so the selection is a parity
// surface, not a parser-rejected input — pinned by the
// duplicate_import_last_occurrence parity cell (04-REVIEW WR-03). The
// engine's D-01 sort normalizes order either way.
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

// Message implements lint.Rule.
func (i *ImportUsedLinked) Message() string {
	return "import is not used"
}

// Validate implements lint.Rule.
func (i *ImportUsedLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
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

	// collects info about import in linted proto file: one entry per import
	// path, overwritten per occurrence so the LAST written statement wins —
	// main's importInfo overwrite semantics (04-REVIEW WR-03);
	// isImportUsed is flagged per occurrence as today.
	importInfo := make(map[core.ImportPath]*protoast.Import, len(info.File.Imports))
	for _, imp := range info.File.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		u.isImportUsed[importPath] = false
		importInfo[importPath] = imp // last written occurrence wins
	}

	// look for import used
	u.checkInServices(info.File.Services, info)
	u.checkInMessages(info.File.Messages, info)
	u.checkInExtends(info.File.Extends, info)
	u.checkInOptions(info.File.Options, info)

	// report every import with used == false, at the import keyword with the
	// quoted location as the source (main's imp.Meta.Pos ≡ imp.Pos and
	// imp.Location basis). Each distinct path reports exactly once — the map
	// is keyed by path — at its LAST written occurrence; the engine's D-01
	// sort normalizes the map iteration order.
	for importPath, imp := range importInfo {
		if u.isImportUsed[importPath] {
			continue
		}
		res = core.AppendIssue(res, i, imp.Pos, imp.Location)
	}

	return res, nil
}

// checkIsImportUsed check if passed import is used in proto file
func (u *importUsage) checkIsImportUsed(key string, info core.ProtoInfo) {
	instruction := u.instrParser.Parse(key)
	for _, importPath := range u.pkgToImport[instruction.PkgName] {
		proto := info.ImportedFiles[importPath]
		exist := existInProto(instruction.Instruction, proto)

		if exist {
			if _, ok := u.isImportUsed[importPath]; ok {
				u.isImportUsed[importPath] = true
			}
		}
	}
}

// check used imports in services
func (u *importUsage) checkInServices(services []*protoast.Service, info core.ProtoInfo) {
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
func (u *importUsage) checkInMessages(messages []*protoast.Message, info core.ProtoInfo) {
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

func (u *importUsage) checkInExtends(extends []*protoast.Extend, info core.ProtoInfo) {
	for _, extend := range extends {
		u.checkIsImportUsed(extend.Type, info)
	}
}

func (u *importUsage) checkInOptions(options []*protoast.Option, info core.ProtoInfo) {
	for _, option := range options {
		u.checkIsImportUsed(option.Name, info)
	}
}

// existInProto looks for key in imported proto file — the etalon named
// it existInProto; the name is taken in this package by main's byte-frozen
// import_used.go, so the port carries the Linked suffix (rename map). The
// reading is main's: the extend fields, the top-level message names and the
// top-level enum names.
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
