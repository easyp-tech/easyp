package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentRPC)(nil)

// CommentRPC this rule checks that RPCs have non-empty comments.
type CommentRPC struct{}

// Message implements lint.Rule.
func (c *CommentRPC) Message() string {
	return "rpc comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentRPC) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if len(service.Comments) == 0 {
				res = append(res, lint.BuildError(c, rpc.Meta.Pos, rpc.RPCName))
			}
		}
	}

	return res, nil
}
