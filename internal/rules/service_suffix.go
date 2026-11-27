package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ServiceSuffix)(nil)

// ServiceSuffix this rule checks that services end with the configured suffix.
type ServiceSuffix struct {
	Suffix string
}

// Message implements lint.Rule.
func (s *ServiceSuffix) Message() string {
	return "service name should have suffix"
}

// Validate implements lint.Rule.
func (s *ServiceSuffix) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, service := range protoInfo.File.Services {
		if !strings.HasSuffix(service.Name, s.Suffix) {
			res = core.AppendIssue(res, s, service.NamePos, service.Name)
		}
	}

	return res, nil
}
