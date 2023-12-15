package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentEnum)(nil)

// CommentEnum is a rule for checking enum comments.
type CommentEnum struct{}

// Validate implements Rule.
func (c *CommentEnum) Validate(svc *unordered.Proto) []error {
	var res []error

	for _, enum := range svc.ProtoBody.Enums {
		if len(enum.Comments) == 0 {
			res = append(res, &Error{
				Err: fmt.Errorf("%d:%d:%s: %w", enum.Meta.Pos.Line, enum.Meta.Pos.Column, enum.EnumName, core.ErrEnumCommentIsEmpty),
			})
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
