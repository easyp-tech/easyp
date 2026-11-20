package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportNoWeak)(nil)

// ImportNoWeak (research
// inventory #11, verbatim-from-lab): the etalon body carries over with the
// imp.Weak flag replacing main's Modifier comparison; the position is
// test-verified equal to main's Meta.Pos (imp.Pos, the import keyword).
type ImportNoWeak struct{}

// Message implements lint.Rule.
func (i *ImportNoWeak) Message() string {
	return "import should not be weak"
}

// Validate implements lint.Rule.
func (i *ImportNoWeak) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, imp := range info.File.Imports {
		if imp.Weak {
			res = core.AppendIssue(res, i, imp.Pos, imp.Location)
		}
	}

	return res, nil
}
