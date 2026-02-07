package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentMessage)(nil)

// CommentMessage this rule checks that messages have non-empty comments.
type CommentMessage struct{}

// Message implements lint.Rule.
func (c *CommentMessage) Message() string {
	return "message comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentMessage) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = core.AppendIssue(res, c, message.Meta.Pos, message.MessageName, message.Comments)
		}
	}

	return res, nil
}
