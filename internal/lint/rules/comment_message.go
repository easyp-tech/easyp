package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentMessage)(nil)

// CommentMessage this rule checks that messages have non-empty comments.
type CommentMessage struct{}

// Name implements lint.Rule.
func (c *CommentMessage) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Validate implements lint.Rule.
func (c *CommentMessage) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = append(res, BuildError(protoInfo.Path, message.Meta.Pos, message.MessageName, lint.ErrMessageCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
