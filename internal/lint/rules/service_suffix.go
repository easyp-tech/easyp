package rules

import (
	"reflect"
	"strings"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ServiceSuffix)(nil)

// ServiceSuffix this rule enforces that all services are suffixed with Service.
type ServiceSuffix struct {
	Suffix string
}

// Name implements lint.Rule.
func (s *ServiceSuffix) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(s).Elem().Name())
}

// Validate enforces that all services are suffixed with Service.
func (s *ServiceSuffix) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !strings.HasSuffix(service.ServiceName, s.Suffix) {
			res = append(res, BuildError(protoInfo.Path, service.Meta.Pos, service.ServiceName, lint.ErrServiceSuffix))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
