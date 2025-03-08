package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*RPCNoClientStreaming)(nil)

// RPCNoClientStreaming this rule checks that RPCs aren't client streaming.
type RPCNoClientStreaming struct {
}

// Message implements lint.Rule.
func (r *RPCNoClientStreaming) Message() string {
	return "client streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoClientStreaming) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if rpc.RPCRequest.IsStream {
				res = core.AppendIssue(res, r, rpc.Meta.Pos, rpc.RPCName, rpc.Comments)
			}
		}
	}

	return res, nil
}
