package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentMessage)(nil)

// CommentMessage this rule checks that messages have non-empty comments.
type CommentMessage struct{}

// Validate implements lint.Rule.
func (c *CommentMessage) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = append(res, BuildError(message.Meta.Pos, message.MessageName, ErrMessageCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
