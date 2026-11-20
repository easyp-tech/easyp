package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ServiceSuffix)(nil)

// ServiceSuffix (research
// inventory #32, verbatim-from-lab predicate): HasSuffix(name, Suffix) is
// identical in main and the lab; the position is main's — the service
// declaration keyword at column 1 (main Meta.Pos ≡ Decl.Pos), never the
// lab's name-token position (column 9). The Suffix config field is plumbed
// through New's defaultIfEmpty "Service", mirroring builder.go.
type ServiceSuffix struct {
	Suffix string
}

// Message implements lint.Rule.
func (s *ServiceSuffix) Message() string {
	return "service name should have suffix"
}

// Validate enforces that all services are suffixed with the configured
// suffix.
func (s *ServiceSuffix) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		if !strings.HasSuffix(service.Name, s.Suffix) {
			res = core.AppendIssue(res, s, service.Pos, service.Name)
		}
	}

	return res, nil
}
