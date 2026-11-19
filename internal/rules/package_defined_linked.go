package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*PackageDefinedLinked)(nil)

// PackageDefinedLinked is the linked-engine port of PackageDefined (research
// inventory #2, main-diverged position): a rule for checking that all files
// have a package declaration. Main emits the zero position literal
// (Filename only; Line 0, Column 0) — the lab's protoast.FilePosition (1:1)
// is forbidden here.
type PackageDefinedLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (p *PackageDefinedLinked) LintRuleName() string {
	return "PACKAGE_DEFINED"
}

// Message implements lint.LintRule.
func (p *PackageDefinedLinked) Message() string {
	return "package should be defined"
}

// Validate implements lint.LintRule.
func (p *PackageDefinedLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	if len(info.File.Packages) == 0 {
		res = core.AppendLinkedIssue(res, p, core.Position{
			Filename: info.Path,
		}, info.Path)
	}

	return res, nil
}
