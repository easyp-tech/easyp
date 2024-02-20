package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumFirstValueZero)(nil)

// EnumFirstValueZero is a rule for checking is first enums value is zero.
type EnumFirstValueZero struct{}

// Validate implements core.Rule.
func (c *EnumFirstValueZero) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if val := enum.EnumBody.EnumFields[0]; val.Number != "0" {
			res = append(res, buildError(val.Meta.Pos, val.Number, lint.ErrEnumFirstValueZero))
		}

	}

	if len(res) == 0 {
		return nil
	}
	return res
}
