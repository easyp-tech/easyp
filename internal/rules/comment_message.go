package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
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
	protoInfo.File.Walk(protoast.Visitor{Message: func(message *protoast.Message) {
		if !leadingCommentDocuments(&message.Decl) {
			res = core.AppendIssue(res, c, message.Pos, message.Name)
		}
	}})

	return res, nil
}
