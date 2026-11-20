package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*RPCPascalCase)(nil)

// RPCPascalCase (research
// inventory #23, main-diverged): main's regexp wins over the lab's
// casing-helper equality (digit-led names would diverge; both parsers reject
// them at the rpc-name token, so no compiling corpus can carry that edge —
// the regexp itself stays the predicate), and the position is the
// declaration keyword (rpc.Meta.Pos ≡ Decl.Pos, main's test pins column 3),
// not the lab's declared-name token.
type RPCPascalCase struct{}

// Message implements lint.Rule.
func (c *RPCPascalCase) Message() string {
	return "RPC names should be PascalCase"
}

// Validate implements lint.Rule.
func (c *RPCPascalCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$")

	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			if !pascalCase.MatchString(rpc.Name) {
				res = core.AppendIssue(res, c, rpc.Pos, rpc.Name)
			}
		}
	}

	return res, nil
}
