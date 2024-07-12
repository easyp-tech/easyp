package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameRubyPackage)(nil)

// PackageSameRubyPackage checks that all files with a given package have the same value for the ruby_package option.
type PackageSameRubyPackage struct {
	// dir => package
	cache map[string]string
}

func (p *PackageSameRubyPackage) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Message implements lint.Rule.
func (p *PackageSameRubyPackage) Message() string {
	return "all files in the same package must have the same ruby_package option"
}

// Validate implements lint.Rule.
func (p *PackageSameRubyPackage) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	p.lazyInit()

	var res []lint.Issue

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return res, nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "ruby_package" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = append(res, lint.BuildError(p, option.Meta.Pos, option.Constant))
			}
		}
	}

	return res, nil
}
