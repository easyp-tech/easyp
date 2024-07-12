package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentEnum)(nil)

// CommentEnum this rule checks that enums have non-empty comments.
type CommentEnum struct{}

// Message implements lint.Rule.
func (c *CommentEnum) Message() string {
	return "enum comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentEnum) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if len(enum.Comments) == 0 {
			res = append(res, lint.BuildError(
				c,
				enum.Meta.Pos,
				enum.EnumName,
			))
		}
	}

	return res, nil
}
