package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentEnumValue)(nil)

// CommentEnumValue this rule checks that enum values have non-empty comments.
type CommentEnumValue struct{}

// Validate implements lint.Rule.
func (c *CommentEnumValue) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if len(field.Comments) == 0 {
				res = AppendError(res, COMMENT_ENUM_VALUE, field.Meta.Pos, field.Ident, field.Comments)
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
