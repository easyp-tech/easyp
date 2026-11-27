package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ServicePascalCase)(nil)

// ServicePascalCase this rule checks that services are PascalCase.
type ServicePascalCase struct{}

// Message implements lint.Rule.
func (c *ServicePascalCase) Message() string {
	return "service names must be PascalCase"
}

// Validate implements lint.Rule.
func (c *ServicePascalCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, service := range protoInfo.File.Services {
		if service.Name != toPascalCase(service.Name) {
			res = core.AppendIssue(res, c, service.NamePos, service.Name)
		}
	}

	return res, nil
}
