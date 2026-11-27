package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*CommentOneof)(nil)

// CommentOneof this rule checks that oneofs have non-empty comments.
type CommentOneof struct{}

// Message implements lint.Rule.
func (c *CommentOneof) Message() string {
	return "oneof comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentOneof) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Oneof: func(oneof *protoast.Oneof) {
		if !leadingCommentDocuments(&oneof.Decl) {
			res = core.AppendIssue(res, c, oneof.Pos, oneof.Name)
		}
	}})

	return res, nil
}
