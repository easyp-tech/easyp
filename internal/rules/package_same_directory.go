package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"path/filepath"
)

var _ core.ModuleRule = (*PackageSameDirectory)(nil)

// PackageSameDirectory this rule checks that all files of a package are in one directory; every file of a split package is reported.
type PackageSameDirectory struct {
	files groupedValues
}

// Message implements lint.Rule.
func (d *PackageSameDirectory) Message() string {
	return "different proto files in the same package should be in the same directory"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (d *PackageSameDirectory) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (d *PackageSameDirectory) Collect(protoInfo core.ProtoInfo) error {
	pos, pkg := packagePosition(protoInfo)
	d.files.add(pkg, groupedValue{path: protoInfo.Path, value: filepath.Dir(protoInfo.Path), source: pkg, pos: pos})
	return nil
}

// Finish implements core.ModuleRule.
func (d *PackageSameDirectory) Finish() []core.IssueInfo { return d.files.mismatches(d) }
