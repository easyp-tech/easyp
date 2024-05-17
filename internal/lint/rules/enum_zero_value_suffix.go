package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumZeroValueSuffix)(nil)

// EnumZeroValueSuffix this rule requires that all enum values have a zero value with a defined suffix.
// By default, it verifies that the zero value of all enums ends in _UNSPECIFIED, but the suffix is configurable.
type EnumZeroValueSuffix struct {
	Suffix string `json:"suffix" yaml:"suffix" ENV:"ENUM_ZERO_VALUE_SUFFIX"`
}

// Validate implements lint.Rule.
func (e EnumZeroValueSuffix) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		zeroValue := enum.EnumBody.EnumFields[0]
		if zeroValue.Ident != pascalToUpperSnake(enum.EnumName)+"_"+e.Suffix {
			res = append(res, BuildError(zeroValue.Meta.Pos, zeroValue.Ident, ErrEnumZeroValueSuffix))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
