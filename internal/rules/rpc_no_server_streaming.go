package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*RPCNoServerStreaming)(nil)

// RPCNoServerStreaming this rule checks that RPCs are not server streaming.
type RPCNoServerStreaming struct{}

// Message implements lint.Rule.
func (r *RPCNoServerStreaming) Message() string {
	return "server streaming RPCs are not allowed"
}

// Validate implements lint.Rule.
func (r *RPCNoServerStreaming) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{RPC: func(_ *protoast.Service, rpc *protoast.RPC) {
		if rpc.Response.Stream {
			res = core.AppendIssue(res, r, rpc.Pos, rpc.Name)
		}
	}})

	return res, nil
}
