package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentEnum)(nil)

// CommentEnum this rule checks that enums have non-empty comments.
type CommentEnum struct{}

// Validate implements lint.Rule.
func (c *CommentEnum) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if len(enum.Comments) == 0 {
			res = append(res, BuildError(enum.Meta.Pos, enum.EnumName, lint.ErrEnumCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
