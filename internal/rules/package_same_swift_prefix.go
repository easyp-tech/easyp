package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageSameSwiftPrefix)(nil)

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

// Message implements lint.Rule.
func (p *PackageSameSwiftPrefix) Message() string {
	return "all files in the same package must have the same swift_prefix option"
}

// Validate implements lint.Rule.
func (p *PackageSameSwiftPrefix) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	p.lazyInit()

	var res []core.Issue

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
				res = core.AppendIssue(res, p, option.Meta.Pos, option.Constant, option.Comments)
			}
		}
	}

	return res, nil
}
