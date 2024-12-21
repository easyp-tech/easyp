package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*RPCNoServerStreaming)(nil)

// RPCNoServerStreaming this rule checks that RPCs aren't server streaming.
type RPCNoServerStreaming struct {
}

// Message implements lint.Rule.
func (r *RPCNoServerStreaming) Message() string {
	return "server streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoServerStreaming) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCResponse.IsStream {
				res = core.AppendIssue(res, r, rpc.Meta.Pos, rpc.RPCName, rpc.Comments)
			}
		}
	}

	return res, nil
}
