package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentMessageField)(nil)

// CommentMessageField is a rule for checking message field comments.
type CommentMessageField struct{}

// Validate implements Rule.
func (c *CommentMessageField) Validate(svc *unordered.Proto) []error {
	var res []error

	for _, message := range svc.ProtoBody.Messages {
		for _, field := range message.MessageBody.Fields {
			if len(field.Comments) == 0 {
				res = append(res, &Error{
					Err: fmt.Errorf("%d:%d:%s: %w", field.Meta.Pos.Line, field.Meta.Pos.Column, field.FieldName, core.ErrMessageFieldCommentIsEmpty),
				})
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
