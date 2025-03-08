package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*CommentEnumValue)(nil)

// CommentEnumValue this rule checks that enum values have non-empty comments.
type CommentEnumValue struct{}

// Message implements lint.Rule.
func (c *CommentEnumValue) Message() string {
	return "enum value comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentEnumValue) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, field := range enum.EnumBody.EnumFields {
			if len(field.Comments) == 0 {
				res = core.AppendIssue(
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
					res = core.AppendIssue(
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
