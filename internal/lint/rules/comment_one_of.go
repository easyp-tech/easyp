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

// Message implements lint.Rule.
func (c *CommentOneof) Message() string {
	return "oneof comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentOneof) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range msg.MessageBody.Oneofs {
			if len(oneof.Comments) == 0 {
				res = append(res, lint.BuildError(oneof.Meta.Pos, oneof.OneofName, c.Message()))
			}
		}
	}

	return res, nil
}
