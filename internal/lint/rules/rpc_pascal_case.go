package rules

import (
	"fmt"
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCPascalCase)(nil)

// RPCPascalCase this rule checks that RPCs are PascalCase.
type RPCPascalCase struct{}

// Message implements lint.Rule.
func (c *RPCPascalCase) Message() string {
	return "RPC names should be PascalCase"
}

// Validate implements lint.Rule.
func (c *RPCPascalCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !pascalCase.MatchString(rpc.RPCName) {
				fmt.Println("RPC name: ", rpc.RPCName)
				res = lint.AppendIssue(res, c, rpc.Meta.Pos, rpc.RPCName, rpc.Comments)
			}
		}
	}

	return res, nil
}
