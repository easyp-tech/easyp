package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageVersionSuffix)(nil)

// PackageVersionSuffix this rule enforces that the last component of a package must be a version of the form
// v\d+, v\d+test.*, v\d+(alpha|beta)\d*, or v\d+p\d+(alpha|beta)\d*, where numbers are >=1.
type PackageVersionSuffix struct{}

// Message implements lint.Rule.
func (p *PackageVersionSuffix) Message() string {
	return "package name should have a version suffix"
}

var matchVersionSuffix = regexp.MustCompile(`.*v\d+|.*v\d+test.*|.*v\d+(alpha|beta)\d*|.*v\d+p\d+(alpha|beta)\d*$`)

// Validate implements lint.Rule.
func (p *PackageVersionSuffix) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, pkg := range protoInfo.Info.ProtoBody.Packages {
		if !matchVersionSuffix.MatchString(pkg.Name) {
			res = lint.AppendIssue(res, p, pkg.Meta.Pos, pkg.Name, pkg.Comments)
		}
	}

	return res, nil
}
