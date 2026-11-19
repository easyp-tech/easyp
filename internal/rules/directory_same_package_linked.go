package rules

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*DirectorySamePackageLinked)(nil)

// DirectorySamePackageLinked is the linked-engine port of DirectorySamePackage
// (research inventory #1, main-diverged + adapt): a rule for checking that all
// files in a given directory are in the same package. The walk mechanics come
// from the etalon's ModuleRule shape; the Finish semantics are main's
// first-exempt cache (the first package value per directory is exempt, later
// differing files issue at their own package positions) — NOT the lab's
// report-everyone groupedValues.mismatches. A file without a package
// declaration contributes nothing: main's loop iterates Packages only.
type DirectorySamePackageLinked struct {
	files groupedValues
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (d *DirectorySamePackageLinked) LintRuleName() string {
	return "DIRECTORY_SAME_PACKAGE"
}

// Message implements lint.LintRule.
func (d *DirectorySamePackageLinked) Message() string {
	return "all files in the same directory must have the same package name"
}

// Validate implements lint.LintRule. Lint calls Collect and Finish instead.
func (d *DirectorySamePackageLinked) Validate(core.LinkedFileInfo) ([]core.Issue, error) {
	return nil, nil
}

// Collect implements core.ModuleRule.
func (d *DirectorySamePackageLinked) Collect(info core.LinkedFileInfo) error {
	for _, pkg := range info.File.Packages {
		d.files.add(filepath.Dir(info.Path), groupedValue{
			path:   info.Path,
			value:  pkg.Name,
			source: pkg.Name,
			pos:    pkg.Pos,
		})
	}
	return nil
}

// Finish implements core.ModuleRule.
func (d *DirectorySamePackageLinked) Finish() []core.IssueInfo {
	return d.files.firstExemptMismatches(d)
}
