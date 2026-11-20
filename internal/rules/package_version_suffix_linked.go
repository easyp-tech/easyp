package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageVersionSuffix)(nil)

// PackageVersionSuffix (research inventory #31, main-diverged; threat
// T-04-12): main's regexp is UNANCHORED — a single-component "v1" matches
// `.*v\d+` — where the lab's two-component version helper
// (lint_helpers.go) would fail it. That helper is not used by this port
// (grep-clean). The position is the package declaration keyword (main
// Meta.Pos ≡ Decl.Pos).
type PackageVersionSuffix struct{}

// Message implements lint.Rule.
func (p *PackageVersionSuffix) Message() string {
	return "package name should have a version suffix"
}

// Validate implements lint.Rule.
func (p *PackageVersionSuffix) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	// main's matchVersionSuffix regexp, byte-verbatim and UNANCHORED.
	// Declared locally: the frozen old file owns the package-level
	// declaration of the same name.
	versionSuffix := regexp.MustCompile(`.*v\d+|.*v\d+test.*|.*v\d+(alpha|beta)\d*|.*v\d+p\d+(alpha|beta)\d*$`)

	for _, pkg := range info.File.Packages {
		if !versionSuffix.MatchString(pkg.Name) {
			res = core.AppendIssue(res, p, pkg.Pos, pkg.Name)
		}
	}

	return res, nil
}
