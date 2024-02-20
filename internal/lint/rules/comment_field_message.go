package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentMessageField)(nil)

// CommentMessageField is a rule for checking message field comments.
type CommentMessageField struct{}

// Validate implements core.Rule.
func (c *CommentMessageField) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if len(field.Comments) == 0 {
				res = append(res, buildError(field.Meta.Pos, field.FieldName, lint.ErrMessageFieldCommentIsEmpty))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
