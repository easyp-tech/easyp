package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*ServiceSuffixLinked)(nil)

// ServiceSuffixLinked is the linked-engine port of ServiceSuffix (research
// inventory #32, verbatim-from-lab predicate): HasSuffix(name, Suffix) is
// identical in main and the lab; the position is main's — the service
// declaration keyword at column 1 (main Meta.Pos ≡ Decl.Pos), never the
// lab's name-token position (column 9). The Suffix config field is plumbed
// through New's defaultIfEmpty "Service", mirroring builder.go.
type ServiceSuffixLinked struct {
	Suffix string
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (s *ServiceSuffixLinked) LintRuleName() string {
	return "SERVICE_SUFFIX"
}

// Message implements lint.Rule.
func (s *ServiceSuffixLinked) Message() string {
	return "service name should have suffix"
}

// Validate enforces that all services are suffixed with the configured
// suffix.
func (s *ServiceSuffixLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, service := range info.File.Services {
		if !strings.HasSuffix(service.Name, s.Suffix) {
			res = core.AppendIssue(res, s, service.Pos, service.Name)
		}
	}

	return res, nil
}
