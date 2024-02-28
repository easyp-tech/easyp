package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentRPC)(nil)

// CommentRPC this rule checks that RPCs have non-empty comments.
type CommentRPC struct{}

// Validate implements lint.Rule.
func (c *CommentRPC) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if len(service.Comments) == 0 {
				res = append(res, BuildError(rpc.Meta.Pos, rpc.RPCName, lint.ErrRPCCommentIsEmpty))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
