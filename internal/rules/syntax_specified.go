package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*SyntaxSpecified)(nil)

// SyntaxSpecified this rule checks that every file has a syntax or edition statement.
type SyntaxSpecified struct{}

// Message implements lint.Rule.
func (s *SyntaxSpecified) Message() string {
	return "syntax should be specified"
}

// Validate implements lint.Rule.
func (s *SyntaxSpecified) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	if !protoInfo.File.SyntaxSpecified {
		res = core.AppendIssue(res, s, protoast.FilePosition(protoInfo.Path), protoInfo.Path)
	}

	return res, nil
}
