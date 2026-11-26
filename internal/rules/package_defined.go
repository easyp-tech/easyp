package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*PackageDefined)(nil)

// PackageDefined this rule checks that all files have a package declaration.
type PackageDefined struct{}

// Message implements lint.Rule.
func (p *PackageDefined) Message() string {
	return "package should be defined"
}

// Validate implements lint.Rule.
func (p *PackageDefined) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	if len(protoInfo.File.Packages) == 0 {
		res = core.AppendIssue(res, p, protoast.FilePosition(protoInfo.Path), protoInfo.Path)
	}

	return res, nil
}
