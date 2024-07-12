package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameGoPackage)(nil)

// PackageSameGoPackage checks that all files with a given package have the same value for the go_package option.
type PackageSameGoPackage struct {
	// dir => package
	cache map[string]string
}

func (p *PackageSameGoPackage) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Message implements lint.Rule.
func (p *PackageSameGoPackage) Message() string {
	return "all files in the same package must have the same go_package name"
}

// Validate implements lint.Rule.
func (p *PackageSameGoPackage) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	p.lazyInit()

	var res []lint.Issue

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return res, nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "go_package" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = append(res, lint.BuildError(option.Meta.Pos, option.Constant, p.Message()))
			}
		}
	}

	return res, nil
}
