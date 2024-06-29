package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameJavaMultipleFiles)(nil)

// PackageSameJavaMultipleFiles checks that all files with a given package have the same value for the java_multiple_files option.
type PackageSameJavaMultipleFiles struct {
	// dir => package
	cache map[string]string
}

func (p *PackageSameJavaMultipleFiles) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (p *PackageSameJavaMultipleFiles) Validate(protoInfo lint.ProtoInfo) []error {
	p.lazyInit()

	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		return nil
	}

	packageName := protoInfo.Info.ProtoBody.Packages[0].Name
	for _, option := range protoInfo.Info.ProtoBody.Options {
		if option.OptionName == "java_multiple_files" {
			if p.cache[packageName] == "" {
				p.cache[packageName] = option.Constant
				continue
			}

			if p.cache[packageName] != option.Constant {
				res = AppendError(res, PACKAGE_SAME_JAVA_MULTIPLE_FILES, option.Meta.Pos, option.Constant, option.Comments)
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
