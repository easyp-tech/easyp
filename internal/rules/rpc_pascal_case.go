package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*RPCPascalCase)(nil)

// RPCPascalCase this rule checks that RPC names are PascalCase.
type RPCPascalCase struct{}

// Message implements lint.Rule.
func (c *RPCPascalCase) Message() string {
	return "RPC names should be PascalCase"
}

// Validate implements lint.Rule.
func (c *RPCPascalCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{RPC: func(_ *protoast.Service, rpc *protoast.RPC) {
		if rpc.Name != toPascalCase(rpc.Name) {
			res = core.AppendIssue(res, c, rpc.NamePos, rpc.Name)
		}
	}})

	return res, nil
}
