package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumZeroValueSuffix)(nil)

// EnumZeroValueSuffix this rule requires that all enum values have a zero value with a defined suffix.
// By default, it verifies that the zero value of all enums ends in _UNSPECIFIED, but the suffix is configurable.
type EnumZeroValueSuffix struct {
	Suffix string `json:"suffix" yaml:"suffix" ENV:"ENUM_ZERO_VALUE_SUFFIX"`
}

// Name implements lint.Rule.
func (e *EnumZeroValueSuffix) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(e).Elem().Name())
}

// Message implements lint.Rule.
func (e *EnumZeroValueSuffix) Message() string {
	return "enum zero value suffix is not valid"
}

// Validate implements lint.Rule.
func (e *EnumZeroValueSuffix) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		zeroValue := enum.EnumBody.EnumFields[0]
		if zeroValue.Ident != pascalToUpperSnake(enum.EnumName)+"_"+e.Suffix {
			res = append(res, lint.BuildError(
				zeroValue.Meta.Pos,
				zeroValue.Ident,
				e.Message(),
			))
		}
	}

	return res, nil
}
