package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ServicePascalCase)(nil)

// ServicePascalCase this rule checks that services are PascalCase.
type ServicePascalCase struct{}

// Validate implements lint.Rule.
func (c *ServicePascalCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z]|[a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !pascalCase.MatchString(service.ServiceName) {
			res = append(res, BuildError(service.Meta.Pos, service.ServiceName, ErrServicePascalCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
