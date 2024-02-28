package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

// EnumNoAllowAlias this rule checks that enums are PascalCase.
type EnumNoAllowAlias struct{}

// Validate implements lint.Rule.
func (e EnumNoAllowAlias) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, opt := range enum.EnumBody.Options {
			if opt.OptionName == "allow_alias" {
				res = append(res, BuildError(enum.Meta.Pos, enum.EnumName, lint.ErrEnumNoAllowAlias))
			}
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
