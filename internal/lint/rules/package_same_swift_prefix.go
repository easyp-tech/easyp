package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameSwiftPrefix)(nil)

// PackageSameSwiftPrefix checks that all files with a given package have the same value for the swift_prefix option.
type PackageSameSwiftPrefix struct {
	// dir => package
	cache map[string]string
}

func (p *PackageSameSwiftPrefix) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (p *PackageSameSwiftPrefix) Validate(protoInfo lint.ProtoInfo) []error {
	p.lazyInit()

	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "ruby_package" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = append(res, BuildError(option.Meta.Pos, option.Constant, ErrPackageSameSwiftPrefix))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
