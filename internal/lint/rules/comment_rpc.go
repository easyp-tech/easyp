package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentRPC)(nil)

// CommentRPC this rule checks that RPCs have non-empty comments.
type CommentRPC struct{}

// Name implements lint.Rule.
func (c *CommentRPC) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

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
				res = append(res, lint.BuildError(rpc.Meta.Pos, rpc.RPCName, c.Message()))
			}
		}
	}

	return res, nil
}
