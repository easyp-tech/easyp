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

// Message implements lint.Rule.
func (c *CommentMessage) Message() string {
	return "message comment is empty"
}

// Validate implements lint.Rule.
func (c *CommentMessage) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, message := range protoInfo.Info.ProtoBody.Messages {
		if len(message.Comments) == 0 {
			res = append(res, lint.BuildError(message.Meta.Pos, message.MessageName, c.Message()))
		}
	}

	return res, nil
}
