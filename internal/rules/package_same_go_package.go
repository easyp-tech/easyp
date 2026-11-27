package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.ModuleRule = (*PackageSameGoPackage)(nil)

// PackageSameGoPackage this rule checks that all files of a package set the go_package option to one value. A file without the option counts as its own value; every file of a disagreeing package is reported.
type PackageSameGoPackage struct {
	files groupedValues
}

// Message implements lint.Rule.
func (p *PackageSameGoPackage) Message() string {
	return "all files in the same package must have the same go_package name"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageSameGoPackage) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (p *PackageSameGoPackage) Collect(protoInfo core.ProtoInfo) error {
	_, pkg := packagePosition(protoInfo)
	value := groupedValue{path: protoInfo.Path, pos: protoast.FilePosition(protoInfo.Path)}
	if opt := fileOption(protoInfo.File, "go_package"); opt != nil {
		value.value, value.source, value.pos = optionString(opt.Value), opt.Value, opt.Pos
	}
	p.files.add(pkg, value)
	return nil
}

// Finish implements core.ModuleRule.
func (p *PackageSameGoPackage) Finish() []core.IssueInfo { return p.files.mismatches(p) }
