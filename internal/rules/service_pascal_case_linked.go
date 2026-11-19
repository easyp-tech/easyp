package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*ServicePascalCaseLinked)(nil)

// ServicePascalCaseLinked is the linked-engine port of ServicePascalCase
// (research inventory #24, main-diverged): main's regexp wins over the lab's
// casing-helper equality — acronym services ("AB", "ABC") fail main but pass
// the helper — and the position is the declaration keyword (service.Meta.Pos
// ≡ Decl.Pos, main's test pins column 1), not the lab's declared-name token.
type ServicePascalCaseLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *ServicePascalCaseLinked) LintRuleName() string {
	return "SERVICE_PASCAL_CASE"
}

// Message implements lint.LintRule.
func (c *ServicePascalCaseLinked) Message() string {
	return "service names must be PascalCase"
}

// Validate implements lint.LintRule.
func (c *ServicePascalCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z]|[a-z]+)*$")

	for _, service := range info.File.Services {
		if !pascalCase.MatchString(service.Name) {
			res = core.AppendLinkedIssue(res, c, service.Pos, service.Name)
		}
	}

	return res, nil
}
