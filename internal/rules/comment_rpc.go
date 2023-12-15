package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentRPC)(nil)

// CommentRPC is a rule for checking rpc comments.
type CommentRPC struct{}

// Validate implements Rule.
func (c *CommentRPC) Validate(svc *unordered.Proto) []error {
	var res []error

	for _, service := range svc.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if len(service.Comments) == 0 {
				res = append(res, &Error{
					Err: fmt.Errorf("%d:%d:%s.%s: %w", rpc.Meta.Pos.Line, rpc.Meta.Pos.Column, service.ServiceName, rpc.RPCName, core.ErrRPCCommentIsEmpty),
				})
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
