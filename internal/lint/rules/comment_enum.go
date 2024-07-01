package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentEnum)(nil)

// CommentEnum this rule checks that enums have non-empty comments.
type CommentEnum struct{}

// Name implements lint.Rule.
func (c *CommentEnum) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Validate implements lint.Rule.
func (c *CommentEnum) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if len(enum.Comments) == 0 {
			res = append(res, BuildError(
				protoInfo.Path,
				enum.Meta.Pos,
				enum.EnumName,
				lint.ErrEnumCommentIsEmpty,
			))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
