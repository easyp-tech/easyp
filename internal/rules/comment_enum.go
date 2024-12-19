package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentEnum)(nil)

// CommentEnum this rule checks that enums have non-empty comments.
type CommentEnum struct{}

// Message implements lint.Rule.
func (c *CommentEnum) Message() string {
	return "enum comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentEnum) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if len(enum.Comments) == 0 {
			res = core.AppendIssue(res, c, enum.Meta.Pos, enum.EnumName, enum.Comments)
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			if len(enum.Comments) == 0 {
				res = core.AppendIssue(res, c, enum.Meta.Pos, enum.EnumName, enum.Comments)
			}
		}
	}

	return res, nil
}
