package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentOneof)(nil)

// CommentOneof this rule checks that oneofs have non-empty comments.
type CommentOneof struct{}

// Name implements lint.Rule.
func (c *CommentOneof) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Validate implements lint.Rule.
func (c *CommentOneof) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range msg.MessageBody.Oneofs {
			if len(oneof.Comments) == 0 {
				res = append(res, BuildError(protoInfo.Path, oneof.Meta.Pos, oneof.OneofName, lint.ErrOneOfCommentIsEmpty))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
