package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.ModuleRule = (*PackageSameRubyPackage)(nil)

// PackageSameRubyPackage this rule checks that all files of a package set the ruby_package option to one value. A file without the option counts as its own value; every file of a disagreeing package is reported.
type PackageSameRubyPackage struct {
	files groupedValues
}

// Message implements lint.Rule.
func (p *PackageSameRubyPackage) Message() string {
	return "all files in the same package must have the same ruby_package option"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageSameRubyPackage) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (p *PackageSameRubyPackage) Collect(protoInfo core.ProtoInfo) error {
	_, pkg := packagePosition(protoInfo)
	value := groupedValue{path: protoInfo.Path, pos: protoast.FilePosition(protoInfo.Path)}
	if opt := fileOption(protoInfo.File, "ruby_package"); opt != nil {
		value.value, value.source, value.pos = optionString(opt.Value), opt.Value, opt.Pos
	}
	p.files.add(pkg, value)
	return nil
}

// Finish implements core.ModuleRule.
func (p *PackageSameRubyPackage) Finish() []core.IssueInfo { return p.files.mismatches(p) }
