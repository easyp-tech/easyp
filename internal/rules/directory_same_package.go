package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"path/filepath"
)

var _ core.ModuleRule = (*DirectorySamePackage)(nil)

// DirectorySamePackage this rule checks that all files in a directory have the same package; every file of a mixed directory is reported.
type DirectorySamePackage struct {
	files groupedValues
}

// Message implements lint.Rule.
func (d *DirectorySamePackage) Message() string {
	return "all files in the same directory must have the same package name"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (d *DirectorySamePackage) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (d *DirectorySamePackage) Collect(protoInfo core.ProtoInfo) error {
	pos, pkg := packagePosition(protoInfo)
	d.files.add(filepath.Dir(protoInfo.Path), groupedValue{path: protoInfo.Path, value: pkg, source: pkg, pos: pos})
	return nil
}

// Finish implements core.ModuleRule.
func (d *DirectorySamePackage) Finish() []core.IssueInfo { return d.files.mismatches(d) }
