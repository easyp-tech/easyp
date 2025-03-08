package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*ImportNoPublic)(nil)

// ImportNoPublic this rule outlaws declaring imports as public.
// If you didn't know that was possible, forget what you just learned in this sentence.
type ImportNoPublic struct{}

// Message implements lint.Rule.
func (i *ImportNoPublic) Message() string {
	return "import should not be public"
}

// Validate implements lint.Rule.
func (i *ImportNoPublic) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		if imp.Modifier == parser.ImportModifierPublic {
			res = core.AppendIssue(res, i, imp.Meta.Pos, imp.Location, imp.Comments)
		}
	}

	return res, nil
}
