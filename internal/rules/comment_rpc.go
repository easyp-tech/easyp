package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*CommentRPC)(nil)

// CommentRPC this rule checks that RPCs have non-empty comments.
type CommentRPC struct{}

// Message implements lint.Rule.
func (c *CommentRPC) Message() string {
	return "rpc comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentRPC) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if len(service.Comments) == 0 {
				res = core.AppendIssue(res, c, rpc.Meta.Pos, rpc.RPCName, service.Comments)
			}
		}
	}

	return res, nil
}
