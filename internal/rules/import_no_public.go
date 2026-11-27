package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportNoPublic)(nil)

// ImportNoPublic this rule checks that imports are not public.
type ImportNoPublic struct{}

// Message implements lint.Rule.
func (i *ImportNoPublic) Message() string {
	return "import should not be public"
}

// Validate implements lint.Rule.
func (i *ImportNoPublic) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, imp := range protoInfo.File.Imports {
		if imp.Public {
			res = core.AppendIssue(res, i, imp.Pos, imp.Location)
		}
	}

	return res, nil
}
