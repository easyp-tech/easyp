package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ImportNoWeak)(nil)

// ImportNoWeak similar to the IMPORT_NO_PUBLIC rule, this rule outlaws declaring imports as weak.
// If you didn't know that was possible, forget what you just learned in this sentence.
type ImportNoWeak struct{}

// Message implements lint.Rule.
func (i *ImportNoWeak) Message() string {
	return "import should not be weak"
}

// Validate implements lint.Rule.
func (i *ImportNoWeak) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		if imp.Modifier == parser.ImportModifierWeak {
			res = core.AppendIssue(res, i, imp.Meta.Pos, imp.Location, imp.Comments)
		}
	}

	return res, nil
}
