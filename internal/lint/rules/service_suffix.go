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

// Message implements lint.Rule.
func (s *ServiceSuffix) Message() string {
	return "service name should have suffix"
}

// Validate enforces that all services are suffixed with Service.
func (s *ServiceSuffix) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !strings.HasSuffix(service.ServiceName, s.Suffix) {
			res = append(res, lint.BuildError(service.Meta.Pos, service.ServiceName, s.Message()))
		}
	}

	return res, nil
}
