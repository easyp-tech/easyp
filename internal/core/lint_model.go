package core

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/protoast"
)

type (
	// Rule is the rule contract of the lint engine.
	Rule interface {
		// Message returns the message of the rule.
		Message() string
		// Validate validates one file's linked model.
		Validate(ProtoInfo) ([]Issue, error)
	}

	// ModuleRule needs every file before it can report, such as rules that
	// compare files of one package or directory. The engine passes each file
	// to Collect and asks for findings once, after the last file; it does not
	// call Validate. Files ignored for the rule never reach Collect at all
	// (main skips their Validate, so they never join the group nor seed its
	// first value — REQ-2.5 parity, 04-REVIEW CR-01); what remains
	// path-scoped is only the residual filtering — comment suppressions and
	// the drain's path filter still gate the findings by path.
	//
	// The rules that implement it after the Phase 4 port (D-03; the registry
	// source of truth is internal/rules/builder.go, which constructs 44 rules
	// — REQUIREMENTS.md says 43, research Pitfall 7):
	// DirectorySamePackage, PackageNoImportCycle, PackageSameDirectory,
	// PackageSameCsharpNamespace, PackageSameGoPackage,
	// PackageSameJavaMultipleFiles, PackageSameJavaPackage,
	// PackageSamePHPNamespace, PackageSameRubyPackage,
	// PackageSameSwiftPrefix, RPCRequestResponseUnique. PackageNoImportCycle
	// arrived with Phase 6.
	ModuleRule interface {
		Rule
		// Collect hands one file to the rule, in walk order.
		Collect(ProtoInfo) error
		// Finish reports the cross-file findings, once after the last file.
		Finish() []IssueInfo
	}

	// Resolution is what linking adds to a file.
	Resolution struct {
		// UsedImports holds the direct imports that define, or publicly
		// re-export a file that defines, a name this file refers to.
		UsedImports map[ImportPath]bool
		// RPCTypes maps an rpc's byte offset to its request and response
		// types, fully qualified without a leading dot.
		RPCTypes map[int][2]string
	}

	// ProtoInfo is the information of a proto file in the linked lint
	// engine.
	ProtoInfo struct {
		Path string
		// ImportPath is relative to a declared source root, independently of
		// the scan root.
		ImportPath string
		// File is the lint view of the file; ImportedFiles holds its direct
		// imports.
		File          *protoast.File
		ImportedFiles map[ImportPath]*protoast.File
		// Resolved holds name resolution across files; nil when the files
		// could not be linked, in which case rules fall back to the syntax.
		Resolved *Resolution
	}
)

// AppendIssue records a rule finding (the etalon's 4-arg shape). It is the
// single issue-construction point for the rules; the lint engine applies
// per-file suppressions.
func AppendIssue(issues []Issue, rule Rule, pos Position, sourceName string) []Issue {
	return append(issues, Issue{
		Position:   pos,
		SourceName: sourceName,
		Message:    rule.Message(),
		RuleName:   GetRuleName(rule),
	})
}

// GetRuleName returns the registry name of a rule: the UPPER_SNAKE spelling
// the registry, findings, and config speak, derived by pure reflection over
// the rule's canonical (unsuffixed) type name.
func GetRuleName(rule Rule) string {
	return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())
}
