package rules

import (
	"strings"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*ServiceSuffix)(nil)

// ServiceSuffix this rule enforces that all services are suffixed with Service.
type ServiceSuffix struct {
	Suffix string
}

// Message implements lint.Rule.
func (s *ServiceSuffix) Message() string {
	return "service name should have suffix"
}

// Validate enforces that all services are suffixed with Service.
func (s *ServiceSuffix) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !strings.HasSuffix(service.ServiceName, s.Suffix) {
			res = core.AppendIssue(res, s, service.Meta.Pos, service.ServiceName, service.Comments)
		}
	}

	return res, nil
}
