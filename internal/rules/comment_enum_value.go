package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*CommentEnumValue)(nil)

// CommentEnumValue this rule checks that enum values have non-empty comments.
type CommentEnumValue struct{}

// Message implements lint.Rule.
func (c *CommentEnumValue) Message() string {
	return "enum value comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentEnumValue) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{EnumValue: func(_ *protoast.Enum, value *protoast.EnumValue) {
		if !leadingCommentDocuments(&value.Decl) {
			res = core.AppendIssue(res, c, value.Pos, value.Name)
		}
	}})

	return res, nil
}
