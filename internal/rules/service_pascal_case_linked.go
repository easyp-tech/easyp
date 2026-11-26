package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ServicePascalCase)(nil)

// ServicePascalCase
// (research inventory #24, main-diverged): main's regexp wins over the lab's
// casing-helper equality — acronym services ("AB", "ABC") fail main but pass
// the helper — and the position is the declaration keyword (service.Meta.Pos
// ≡ Decl.Pos, main's test pins column 1), not the lab's declared-name token.
type ServicePascalCase struct{}

// Message implements lint.Rule.
func (c *ServicePascalCase) Message() string {
	return "service names must be PascalCase"
}

// Validate implements lint.Rule.
func (c *ServicePascalCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z]|[a-z]+)*$")

	for _, service := range info.File.Services {
		if !pascalCase.MatchString(service.Name) {
			res = core.AppendIssue(res, c, service.Pos, service.Name)
		}
	}

	return res, nil
}
