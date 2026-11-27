package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*CommentEnum)(nil)

// CommentEnum this rule checks that enums have non-empty comments.
type CommentEnum struct{}

// Message implements lint.Rule.
func (c *CommentEnum) Message() string {
	return "enum comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentEnum) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Enum: func(enum *protoast.Enum) {
		if !leadingCommentDocuments(&enum.Decl) {
			res = core.AppendIssue(res, c, enum.Pos, enum.Name)
		}
	}})

	return res, nil
}
