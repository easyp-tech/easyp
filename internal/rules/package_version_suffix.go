package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageVersionSuffix)(nil)

// PackageVersionSuffix this rule checks that the last component of the package is a version such as v1, v1beta1 or v1test.
type PackageVersionSuffix struct{}

// Message implements lint.Rule.
func (p *PackageVersionSuffix) Message() string {
	return "package name should have a version suffix"
}

// Validate implements lint.Rule.
func (p *PackageVersionSuffix) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	for _, pkg := range protoInfo.File.Packages {
		if !packageHasVersionSuffix(pkg.Name) {
			res = core.AppendIssue(res, p, pkg.Pos, pkg.Name)
		}
	}

	return res, nil
}
