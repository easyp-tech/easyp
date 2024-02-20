package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentEnum)(nil)

// CommentEnum is a rule for checking enum comments.
type CommentEnum struct{}

// Validate implements core.Rule.
func (c *CommentEnum) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if len(enum.Comments) == 0 {
			res = append(res, buildError(enum.Meta.Pos, enum.EnumName, lint.ErrEnumCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
