package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentOneOf)(nil)

// CommentOneOf is a rule for checking oneOf comments.
type CommentOneOf struct{}

// Validate implements Rule.
func (c *CommentOneOf) Validate(protoInfo core.ProtoInfo) []error {
	var res []error

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range msg.MessageBody.Oneofs {
			if len(oneof.Comments) == 0 {
				res = append(res, buildError(oneof.Meta.Pos, oneof.OneofName, core.ErrOneOfCommentIsEmpty))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
