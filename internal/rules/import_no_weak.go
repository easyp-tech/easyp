package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportNoWeak)(nil)

// ImportNoWeak this rule checks that imports are not weak.
type ImportNoWeak struct{}

// Message implements lint.Rule.
func (i *ImportNoWeak) Message() string {
	return "import should not be weak"
}

// Validate implements lint.Rule.
func (i *ImportNoWeak) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, imp := range protoInfo.File.Imports {
		if imp.Weak {
			res = core.AppendIssue(res, i, imp.Pos, imp.Location)
		}
	}

	return res, nil
}
