package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ImportNoPublic)(nil)

// ImportNoPublic this rule outlaws declaring imports as public.
// If you didn't know that was possible, forget what you just learned in this sentence.
type ImportNoPublic struct{}

// Message implements lint.Rule.
func (i *ImportNoPublic) Message() string {
	return "import should not be public"
}

// Validate implements lint.Rule.
func (i *ImportNoPublic) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		if imp.Modifier == parser.ImportModifierPublic {
			res = append(res, lint.BuildError(i, imp.Meta.Pos, imp.Location))
		}
	}

	return res, nil
}
