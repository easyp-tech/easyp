package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.ModuleRule = (*PackageSamePHPNamespace)(nil)

// PackageSamePHPNamespace this rule checks that all files of a package set the php_namespace option to one value. A file without the option counts as its own value; every file of a disagreeing package is reported.
type PackageSamePHPNamespace struct {
	files groupedValues
}

// Message implements lint.Rule.
func (p *PackageSamePHPNamespace) Message() string {
	return "all files in the same package must have the same php_namespace option"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageSamePHPNamespace) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (p *PackageSamePHPNamespace) Collect(protoInfo core.ProtoInfo) error {
	_, pkg := packagePosition(protoInfo)
	value := groupedValue{path: protoInfo.Path, pos: protoast.FilePosition(protoInfo.Path)}
	if opt := fileOption(protoInfo.File, "php_namespace"); opt != nil {
		value.value, value.source, value.pos = optionString(opt.Value), opt.Value, opt.Pos
	}
	p.files.add(pkg, value)
	return nil
}

// Finish implements core.ModuleRule.
func (p *PackageSamePHPNamespace) Finish() []core.IssueInfo { return p.files.mismatches(p) }
