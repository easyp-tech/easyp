package rules

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*PackageSameDirectory)(nil)

// PackageSameDirectory
// (research inventory #4, main-diverged + adapt): a rule for checking that all
// files with a given package are in the same directory. The walk mechanics
// come from the etalon's ModuleRule shape; the Finish semantics are main's
// first-exempt cache keyed by package (the first directory value per package
// is exempt, later differing files issue at their own package positions). A
// file without a package declaration contributes nothing: main's loop
// iterates Packages only.
type PackageSameDirectory struct {
	files groupedValues
}

// Message implements lint.Rule.
func (d *PackageSameDirectory) Message() string {
	return "different proto files in the same package should be in the same directory"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (d *PackageSameDirectory) Validate(core.ProtoInfo) ([]core.Issue, error) {
	return nil, nil
}

// Collect implements core.ModuleRule.
func (d *PackageSameDirectory) Collect(info core.ProtoInfo) error {
	for _, pkg := range info.File.Packages {
		d.files.add(pkg.Name, groupedValue{
			path:   info.Path,
			value:  filepath.Dir(info.Path),
			source: pkg.Name,
			pos:    pkg.Pos,
		})
	}
	return nil
}

// Finish implements core.ModuleRule.
func (d *PackageSameDirectory) Finish() []core.IssueInfo {
	return d.files.firstExemptMismatches(d)
}
