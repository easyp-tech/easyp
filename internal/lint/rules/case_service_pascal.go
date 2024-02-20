package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ServicePascalCase)(nil)

// ServicePascalCase is a rule for checking name of service for pascal case.
type ServicePascalCase struct{}

// Validate implements core.Rule.
func (c *ServicePascalCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+([A-Z]|[a-z]+)*$")
	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !pascalCase.MatchString(service.ServiceName) {
			res = append(res, buildError(service.Meta.Pos, service.ServiceName, lint.ErrServicePascalCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
