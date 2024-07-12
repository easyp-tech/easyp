package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCNoClientStreaming)(nil)

// RPCNoClientStreaming this rule checks that RPCs aren't client streaming.
type RPCNoClientStreaming struct {
}

// Message implements lint.Rule.
func (r *RPCNoClientStreaming) Message() string {
	return "client streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoClientStreaming) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.IsStream {
				res = append(res, lint.BuildError(r, rpc.Meta.Pos, rpc.RPCName))
			}
		}
	}

	return res, nil
}
