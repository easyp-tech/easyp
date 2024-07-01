package rules

import (
	"reflect"
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*RPCPascalCase)(nil)

// RPCPascalCase this rule checks that RPCs are PascalCase.
type RPCPascalCase struct{}

// Name implements lint.Rule.
func (c *RPCPascalCase) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Validate implements lint.Rule.
func (c *RPCPascalCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z][a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		for _, rpc := range service.ServiceBody.RPCs {
			if !pascalCase.MatchString(rpc.RPCName) {
				res = append(res, BuildError(protoInfo.Path, rpc.Meta.Pos, rpc.RPCName, lint.ErrRpcPascalCase))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
