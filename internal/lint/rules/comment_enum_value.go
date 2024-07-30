package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentEnumValue)(nil)

// CommentEnumValue this rule checks that enum values have non-empty comments.
type CommentEnumValue struct{}

// Message implements lint.Rule.
func (c *CommentEnumValue) Message() string {
	return "enum value comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentEnumValue) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if len(field.Comments) == 0 {
				res = lint.AppendIssue(
					res,
					c,
					field.Meta.Pos,
					field.Ident,
					field.Comments)
			}
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			for _, field := range enum.EnumBody.EnumFields {
				if len(field.Comments) == 0 {
					res = lint.AppendIssue(
						res,
						c,
						field.Meta.Pos,
						field.Ident,
						field.Comments)
				}
			}
		}
	}

	return res, nil
}
