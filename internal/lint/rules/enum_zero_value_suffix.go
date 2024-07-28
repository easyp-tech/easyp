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

// Message implements lint.Rule.
func (e *EnumZeroValueSuffix) Message() string {
	return "enum zero value suffix is not valid"
}

// Validate implements lint.Rule.
func (e *EnumZeroValueSuffix) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		zeroValue := enum.EnumBody.EnumFields[0]
		if zeroValue.Ident != pascalToUpperSnake(enum.EnumName)+e.Suffix {
			res = lint.AppendIssue(
				res,
				e,
				zeroValue.Meta.Pos,
				zeroValue.Ident,
				zeroValue.Comments,
			)
		}
	}

	for _, msg := range protoInfo.Info.ProtoBody.Messages {
		for _, enum := range msg.MessageBody.Enums {
			zeroValue := enum.EnumBody.EnumFields[0]
			if zeroValue.Ident != pascalToUpperSnake(enum.EnumName)+"_"+e.Suffix {
				res = lint.AppendIssue(
					res,
					e,
					zeroValue.Meta.Pos,
					zeroValue.Ident,
					zeroValue.Comments,
				)
			}
		}
	}

	return res, nil
}
