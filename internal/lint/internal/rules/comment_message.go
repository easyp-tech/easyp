package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentMessage)(nil)

// CommentMessage is a rule for checking message comments.
type CommentMessage struct{}

// Validate implements core.Rule.
func (c *CommentMessage) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = append(res, buildError(message.Meta.Pos, message.MessageName, lint.ErrMessageCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
