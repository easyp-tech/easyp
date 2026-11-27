package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.ModuleRule = (*PackageSameSwiftPrefix)(nil)

// PackageSameSwiftPrefix this rule checks that all files of a package set the swift_prefix option to one value. A file without the option counts as its own value; every file of a disagreeing package is reported.
type PackageSameSwiftPrefix struct {
	files groupedValues
}

// Message implements lint.Rule.
func (p *PackageSameSwiftPrefix) Message() string {
	return "all files in the same package must have the same swift_prefix option"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageSameSwiftPrefix) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (p *PackageSameSwiftPrefix) Collect(protoInfo core.ProtoInfo) error {
	_, pkg := packagePosition(protoInfo)
	value := groupedValue{path: protoInfo.Path, pos: protoast.FilePosition(protoInfo.Path)}
	if opt := fileOption(protoInfo.File, "swift_prefix"); opt != nil {
		value.value, value.source, value.pos = optionString(opt.Value), opt.Value, opt.Pos
	}
	p.files.add(pkg, value)
	return nil
}

// Finish implements core.ModuleRule.
func (p *PackageSameSwiftPrefix) Finish() []core.IssueInfo { return p.files.mismatches(p) }
