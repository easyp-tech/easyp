package rules

import (
	"reflect"

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

// Name implements lint.Rule.
func (p *PackageSamePHPNamespace) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(p).Elem().Name())
}

// Message implements lint.Rule.
func (p *PackageSamePHPNamespace) Message() string {
	return "all files in the same package must have the same php_namespace option"
}

// Validate implements lint.Rule.
func (p *PackageSamePHPNamespace) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	p.lazyInit()

	var res []lint.Issue

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return res, nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "php_namespace" {
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
