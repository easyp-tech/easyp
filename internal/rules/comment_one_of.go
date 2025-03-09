package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*CommentOneof)(nil)

// CommentOneof this rule checks that oneofs have non-empty comments.
type CommentOneof struct{}

// Message implements lint.Rule.
func (c *CommentOneof) Message() string {
	return "oneof comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentOneof) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range msg.MessageBody.Oneofs {
			if len(oneof.Comments) == 0 {
				res = core.AppendIssue(res, c, oneof.Meta.Pos, oneof.OneofName, oneof.Comments)
			}
		}
	}

	return res, nil
}
