package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameGoPackage)(nil)

// PackageSameGoPackage checks that all files with a given package have the same value for the go_package option.
type PackageSameGoPackage struct {
	// dir => package
	cache map[string]string
}

// Name implements lint.Rule.
func (p *PackageSameGoPackage) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(p).Elem().Name())
}

func (p *PackageSameGoPackage) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (p *PackageSameGoPackage) Validate(protoInfo lint.ProtoInfo) []error {
	p.lazyInit()

	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "go_package" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = append(res, BuildError(protoInfo.Path, option.Meta.Pos, option.Constant, lint.ErrPackageSameGoPackage))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
