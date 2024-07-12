package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCNoServerStreaming)(nil)

// RPCNoServerStreaming this rule checks that RPCs aren't server streaming.
type RPCNoServerStreaming struct {
}

// Message implements lint.Rule.
func (r *RPCNoServerStreaming) Message() string {
	return "server streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoServerStreaming) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCResponse.IsStream {
				res = append(res, lint.BuildError(r, rpc.Meta.Pos, rpc.RPCName))
			}
		}
	}

	return res, nil
}
