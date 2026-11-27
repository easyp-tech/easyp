package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*RPCNoClientStreaming)(nil)

// RPCNoClientStreaming this rule checks that RPCs are not client streaming.
type RPCNoClientStreaming struct{}

// Message implements lint.Rule.
func (r *RPCNoClientStreaming) Message() string {
	return "client streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoClientStreaming) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{RPC: func(_ *protoast.Service, rpc *protoast.RPC) {
		if rpc.Request.Stream {
			res = core.AppendIssue(res, r, rpc.Pos, rpc.Name)
		}
	}})

	return res, nil
}
