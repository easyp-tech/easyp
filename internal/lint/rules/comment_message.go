package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentMessage)(nil)

// CommentMessage this rule checks that messages have non-empty comments.
type CommentMessage struct{}

// Message implements lint.Rule.
func (c *CommentMessage) Message() string {
	return "message comment is empty"
}

// Validate implements lint.Rule.
func (c *CommentMessage) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = lint.AppendIssue(res, c, message.Meta.Pos, message.MessageName, message.Comments)
		}
	}

	return res, nil
}
