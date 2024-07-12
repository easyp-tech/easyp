package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentField)(nil)

// CommentField this rule checks that fields have non-empty comments.
type CommentField struct{}

// Message implements lint.Rule.
func (c *CommentField) Message() string {
	return "field comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentField) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if len(field.Comments) == 0 {
				res = append(res, lint.BuildError(field.Meta.Pos, field.FieldName, c.Message()))
			}
		}
	}

	return res, nil
}
