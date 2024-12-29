package rules

import (
	"regexp"

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

	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z]|[a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !pascalCase.MatchString(service.ServiceName) {
			res = core.AppendIssue(res, c, service.Meta.Pos, service.ServiceName, service.Comments)
		}
	}

	return res, nil
}
