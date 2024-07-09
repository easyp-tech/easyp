package rules

import (
	"reflect"
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ServicePascalCase)(nil)

// ServicePascalCase this rule checks that services are PascalCase.
type ServicePascalCase struct{}

// Name implements lint.Rule.
func (c *ServicePascalCase) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Message implements lint.Rule.
func (c *ServicePascalCase) Message() string {
	return "service names must be PascalCase"
}

// Validate implements lint.Rule.
func (c *ServicePascalCase) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z]|[a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !pascalCase.MatchString(service.ServiceName) {
			res = append(res, lint.BuildError(service.Meta.Pos, service.ServiceName, c.Message()))
		}
	}

	return res, nil
}
