package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentOneOf)(nil)

// CommentOneOf is a rule for checking oneOf comments.
type CommentOneOf struct{}

// Validate implements Rule.
func (c *CommentOneOf) Validate(svc *unordered.Proto) []error {
	var res []error

	for _, msg := range svc.ProtoBody.Messages {
		for _, oneof := range msg.MessageBody.Oneofs {
			if len(oneof.Comments) == 0 {
				res = append(res, &Error{
					Err: fmt.Errorf("%d:%d:%s: %w", oneof.Meta.Pos.Line, oneof.Meta.Pos.Column, oneof.OneofName, core.ErrOneOfCommentIsEmpty),
				})
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
