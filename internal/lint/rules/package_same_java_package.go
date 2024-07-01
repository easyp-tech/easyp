package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameJavaPackage)(nil)

// PackageSameJavaPackage checks that all files with a given package have the same value for the java_package option.
type PackageSameJavaPackage struct {
	// dir => package
	cache map[string]string
}

// Name implements lint.Rule.
func (p *PackageSameJavaPackage) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(p).Elem().Name())
}

func (p *PackageSameJavaPackage) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (p *PackageSameJavaPackage) Validate(protoInfo lint.ProtoInfo) []error {
	p.lazyInit()

	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "java_package" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = append(res, BuildError(protoInfo.Path, option.Meta.Pos, option.Constant, lint.ErrPackageSameJavaPackage))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
