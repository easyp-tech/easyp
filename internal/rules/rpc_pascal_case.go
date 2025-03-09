package rules

import (
	"regexp"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*RPCPascalCase)(nil)

// RPCPascalCase this rule checks that RPCs are PascalCase.
type RPCPascalCase struct{}

// Message implements lint.Rule.
func (c *RPCPascalCase) Message() string {
	return "RPC names should be PascalCase"
}

// Validate implements lint.Rule.
func (c *RPCPascalCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !pascalCase.MatchString(rpc.RPCName) {
				res = core.AppendIssue(res, c, rpc.Meta.Pos, rpc.RPCName, rpc.Comments)
			}
		}
	}

	return res, nil
}
