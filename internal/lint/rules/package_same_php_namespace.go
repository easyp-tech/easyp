package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSamePHPNamespace)(nil)

// PackageSamePHPNamespace checks that all files with a given package have the same value for the php_namespace option.
type PackageSamePHPNamespace struct {
	// dir => package
	cache map[string]string
}

func (p *PackageSamePHPNamespace) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (p *PackageSamePHPNamespace) Validate(protoInfo lint.ProtoInfo) []error {
	p.lazyInit()

	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "php_namespace" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = append(res, BuildError(option.Meta.Pos, option.Constant, lint.ErrPackageSamePhpNamespace))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
