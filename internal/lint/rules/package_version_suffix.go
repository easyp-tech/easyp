package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageVersionSuffix)(nil)

// PackageVersionSuffix this rule enforces that the last component of a package must be a version of the form
// v\d+, v\d+test.*, v\d+(alpha|beta)\d*, or v\d+p\d+(alpha|beta)\d*, where numbers are >=1.
type PackageVersionSuffix struct{}

var matchVersionSuffix = regexp.MustCompile(`.*v\d+|.*v\d+test.*|.*v\d+(alpha|beta)\d*|.*v\d+p\d+(alpha|beta)\d*$`)

// Validate implements lint.Rule.
func (p PackageVersionSuffix) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, pkg := range protoInfo.Info.ProtoBody.Packages {
		if !matchVersionSuffix.MatchString(pkg.Name) {
			res = append(res, BuildError(pkg.Meta.Pos, pkg.Name, ErrPackageVersionSuffix))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
