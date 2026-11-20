package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportNoPublicLinked)(nil)

// ImportNoPublicLinked is the linked-engine port of ImportNoPublic (research
// inventory #10, verbatim-from-lab): the etalon body carries over with the
// imp.Public flag replacing main's Modifier comparison; the position is
// test-verified equal to main's Meta.Pos (imp.Pos, the import keyword).
type ImportNoPublicLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (i *ImportNoPublicLinked) LintRuleName() string {
	return "IMPORT_NO_PUBLIC"
}

// Message implements lint.Rule.
func (i *ImportNoPublicLinked) Message() string {
	return "import should not be public"
}

// Validate implements lint.Rule.
func (i *ImportNoPublicLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, imp := range info.File.Imports {
		if imp.Public {
			res = core.AppendIssue(res, i, imp.Pos, imp.Location)
		}
	}

	return res, nil
}
