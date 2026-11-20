package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportNoWeakLinked)(nil)

// ImportNoWeakLinked is the linked-engine port of ImportNoWeak (research
// inventory #11, verbatim-from-lab): the etalon body carries over with the
// imp.Weak flag replacing main's Modifier comparison; the position is
// test-verified equal to main's Meta.Pos (imp.Pos, the import keyword).
type ImportNoWeakLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (i *ImportNoWeakLinked) LintRuleName() string {
	return "IMPORT_NO_WEAK"
}

// Message implements lint.Rule.
func (i *ImportNoWeakLinked) Message() string {
	return "import should not be weak"
}

// Validate implements lint.Rule.
func (i *ImportNoWeakLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, imp := range info.File.Imports {
		if imp.Weak {
			res = core.AppendIssue(res, i, imp.Pos, imp.Location)
		}
	}

	return res, nil
}
