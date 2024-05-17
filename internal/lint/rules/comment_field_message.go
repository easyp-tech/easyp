package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentField)(nil)

// CommentField this rule checks that fields have non-empty comments.
type CommentField struct{}

// Validate implements lint.Rule.
func (c *CommentField) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if len(field.Comments) == 0 {
				res = append(res, BuildError(field.Meta.Pos, field.FieldName, ErrMessageFieldCommentIsEmpty))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
