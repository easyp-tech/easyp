package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageSameJavaPackage)(nil)

// PackageSameJavaPackage checks that all files with a given package have the same value for the java_package option.
type PackageSameJavaPackage struct {
	// dir => package
	cache map[string]string
}

func (p *PackageSameJavaPackage) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Message implements lint.Rule.
func (p *PackageSameJavaPackage) Message() string {
	return "all files in the same package must have the same java_package option"
}

// Validate implements lint.Rule.
func (p *PackageSameJavaPackage) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	p.lazyInit()

	var res []core.Issue

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return res, nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "java_package" {
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
