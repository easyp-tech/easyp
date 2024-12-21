package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentField)(nil)

// CommentField this rule checks that fields have non-empty comments.
type CommentField struct{}

// Message implements lint.Rule.
func (c *CommentField) Message() string {
	return "field comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentField) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if len(field.Comments) == 0 {
				res = core.AppendIssue(res, c, field.Meta.Pos, field.FieldName, field.Comments)
			}
		}
	}

	return res, nil
}
