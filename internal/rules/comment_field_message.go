package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*CommentField)(nil)

// CommentField this rule checks that fields have non-empty comments. Group fields are documented on the group itself.
type CommentField struct{}

// Message implements lint.Rule.
func (c *CommentField) Message() string {
	return "field comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentField) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Field: func(field *protoast.Field) {
		if !field.Group && !leadingCommentDocuments(&field.Decl) {
			res = core.AppendIssue(res, c, field.Pos, field.Name)
		}
	}})

	return res, nil
}
