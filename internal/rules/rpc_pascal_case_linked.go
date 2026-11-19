package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*RPCPascalCaseLinked)(nil)

// RPCPascalCaseLinked is the linked-engine port of RPCPascalCase (research
// inventory #23, main-diverged): main's regexp wins over the lab's
// casing-helper equality (digit-led names would diverge; both parsers reject
// them at the rpc-name token, so no compiling corpus can carry that edge —
// the regexp itself stays the predicate), and the position is the
// declaration keyword (rpc.Meta.Pos ≡ Decl.Pos, main's test pins column 3),
// not the lab's declared-name token.
type RPCPascalCaseLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *RPCPascalCaseLinked) LintRuleName() string {
	return "RPC_PASCAL_CASE"
}

// Message implements lint.LintRule.
func (c *RPCPascalCaseLinked) Message() string {
	return "RPC names should be PascalCase"
}

// Validate implements lint.LintRule.
func (c *RPCPascalCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$")

	for _, service := range info.File.Services {
		for _, rpc := range service.RPCs {
			if !pascalCase.MatchString(rpc.Name) {
				res = core.AppendLinkedIssue(res, c, rpc.Pos, rpc.Name)
			}
		}
	}

	return res, nil
}
