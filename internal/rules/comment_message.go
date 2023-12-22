package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentMessage)(nil)

// CommentMessage is a rule for checking message comments.
type CommentMessage struct{}

// Validate implements Rule.
func (c *CommentMessage) Validate(svc *unordered.Proto) []error {
	var res []error

	for _, message := range svc.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = append(res, &Error{
				Err: fmt.Errorf("%d:%d:%s: %w", message.Meta.Pos.Line, message.Meta.Pos.Column, message.MessageName, core.ErrMessageCommentIsEmpty),
			})
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
