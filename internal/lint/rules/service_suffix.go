package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ServiceSuffix)(nil)

// ServiceSuffix this rule enforces that all services are suffixed with Service.
type ServiceSuffix struct {
	Suffix string
}

// Validate enforces that all services are suffixed with Service.
func (s ServiceSuffix) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if !strings.HasSuffix(service.ServiceName, s.Suffix) {
			res = AppendError(res, SERVICE_SUFFIX, service.Meta.Pos, service.ServiceName, service.Comments)
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
