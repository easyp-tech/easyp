package core

// The linked lint engine is a port of the compiler-lab etalon
// (easyp-tech/compiler-lab @ 08cef10, easyp-cli/internal/core). The etalon is
// the post-flip tree where the old engine's names are gone; this package still
// runs both engines side by side until Phase 5 deletes the old path, so every
// identifier the two engines share by name was renamed here. Rename map
// (etalon -> new), kept for the Phase 5 consolidation pass:
//
//	(method) Lint                -> LintLinked
//	(struct) ProtoInfo           -> LinkedFileInfo
//	(interface) Rule             -> LintRule
//	(interface) ModuleRule       -> ModuleRule (name was free; re-typed)
//	(method) commentSuppressions -> linkedCommentSuppressions
//	(struct) directiveComment    -> linkedDirectiveComment
//	(struct) parsedDirective     -> linkedParsedDirective
//	(method) readFilesFromImport -> linkedFilesFromImport
//	(method) readFileFromImport  -> linkedFileFromImport
//	(method) shouldIgnore        -> shouldIgnoreLinked (a mirror, not shared:
//	   the original's parameter type pins it to the old rule contract)
//	(func) AppendIssue (5-arg)   -> AppendLinkedIssue (below; the etalon's
//	   4-arg shape: the linked contract carries no comments parameter)
//	(func) GetRuleName           -> LinkedRuleName (below), plus the additive
//	   optional-method preference: a rule implementing LintRuleName() string
//	   reports that name before the reflection fallback — suffixed port types
//	   (<Name>Linked, Phase 4) return their canonical registry name through it
//
// Shared-by-value symbols are reused in place from the old files and never
// redeclared: Position, Issue, IssueInfo, ImportPath, PackageName,
// GetRuleName, buildError, suppressionRange/suppressionRanges/contains,
// parseDirectivePrefix, compactDirectiveNames, openImportFile,
// sourceImportPath, openDiskSource, CompileRuleName, linkSources,
// ErrInvalidRule.

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/protoast"
)

type (
	// LintRule is the rule contract of the linked lint engine.
	LintRule interface {
		// Message returns the message of the rule.
		Message() string
		// Validate validates one file's linked model.
		Validate(LinkedFileInfo) ([]Issue, error)
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
	// source of truth is internal/rules/builder.go, which constructs 41 rules
	// — REQUIREMENTS.md says 43, research Pitfall 7):
	// DirectorySamePackage, PackageSameDirectory, PackageSameCsharpNamespace,
	// PackageSameGoPackage, PackageSameJavaMultipleFiles,
	// PackageSameJavaPackage, PackageSamePHPNamespace, PackageSameRubyPackage,
	// PackageSameSwiftPrefix, RPCRequestResponseUnique. PackageNoImportCycle
	// arrives with Phase 6.
	ModuleRule interface {
		LintRule
		// Collect hands one file to the rule, in walk order.
		Collect(LinkedFileInfo) error
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

	// LinkedFileInfo is the information of a proto file in the linked lint
	// engine.
	LinkedFileInfo struct {
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

// AppendLinkedIssue records a linked-engine rule finding (rename map: the
// etalon's 4-arg AppendIssue). It is the single issue-construction point for
// the ported rules; the lint engine applies per-file suppressions.
func AppendLinkedIssue(issues []Issue, rule LintRule, pos Position, sourceName string) []Issue {
	return append(issues, Issue{
		Position:   pos,
		SourceName: sourceName,
		Message:    rule.Message(),
		RuleName:   LinkedRuleName(rule),
	})
}

// LinkedRuleName returns the registry name of a linked-engine rule: the same
// UPPER_SNAKE spelling the old engine derives with GetRuleName. A rule may
// override the reflection-derived name with an optional LintRuleName method —
// suffixed port types (<Name>Linked) report their canonical registry name
// through it, so no _LINKED suffix leaks into findings, config, or schemas.
func LinkedRuleName(rule LintRule) string {
	if named, ok := rule.(interface{ LintRuleName() string }); ok {
		return named.LintRuleName()
	}
	return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())
}
