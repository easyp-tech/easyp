package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ImportNoWeak)(nil)

// ImportNoWeak similar to the IMPORT_NO_PUBLIC rule, this rule outlaws declaring imports as weak.
// If you didn't know that was possible, forget what you just learned in this sentence.
type ImportNoWeak struct{}

// Validate implements lint.Rule.
func (i ImportNoWeak) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		if imp.Modifier == parser.ImportModifierWeak {
			res = AppendError(res, IMPORT_NO_WEAK, imp.Meta.Pos, imp.Location, imp.Comments)
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
