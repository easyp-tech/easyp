package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentMessage)(nil)

// CommentMessage is a rule for checking message comments.
type CommentMessage struct{}

// Validate implements core.Rule.
func (c *CommentMessage) Validate(protoInfo core.ProtoInfo) []error {
	var res []error

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = append(res, buildError(message.Meta.Pos, message.MessageName, core.ErrMessageCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
