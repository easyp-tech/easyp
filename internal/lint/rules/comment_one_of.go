package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentOneOf)(nil)

// CommentOneOf this rule checks that oneofs have non-empty comments.
type CommentOneOf struct{}

// Validate implements lint.Rule.
func (c *CommentOneOf) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, oneof := range msg.MessageBody.Oneofs {
			if len(oneof.Comments) == 0 {
				res = append(res, BuildError(oneof.Meta.Pos, oneof.OneofName, lint.ErrOneOfCommentIsEmpty))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
