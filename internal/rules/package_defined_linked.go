package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageDefined)(nil)

// PackageDefined (research
// inventory #2, main-diverged position): a rule for checking that all files
// have a package declaration. Main emits the zero position literal
// (Filename only; Line 0, Column 0) — the lab's protoast.FilePosition (1:1)
// is forbidden here.
type PackageDefined struct{}

// Message implements lint.Rule.
func (p *PackageDefined) Message() string {
	return "package should be defined"
}

// Validate implements lint.Rule.
func (p *PackageDefined) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	if len(info.File.Packages) == 0 {
		res = core.AppendIssue(res, p, core.Position{
			Filename: info.Path,
		}, info.Path)
	}

	return res, nil
}
