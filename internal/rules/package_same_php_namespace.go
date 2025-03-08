package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*PackageSamePHPNamespace)(nil)

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

// Message implements lint.Rule.
func (p *PackageSamePHPNamespace) Message() string {
	return "all files in the same package must have the same php_namespace option"
}

// Validate implements lint.Rule.
func (p *PackageSamePHPNamespace) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	p.lazyInit()

	var res []core.Issue

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
				res = core.AppendIssue(res, p, option.Meta.Pos, option.Constant, option.Comments)
			}
		}
	}

	return res, nil
}
