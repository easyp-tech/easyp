package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumFirstValueZero)(nil)

// EnumFirstValueZero is a rule for checking is first enums value is zero.
type EnumFirstValueZero struct{}

// Validate implements Rule.
func (c *EnumFirstValueZero) Validate(protoInfo core.ProtoInfo) []error {
	var res []error
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if val := enum.EnumBody.EnumFields[0]; val.Number != "0" {
			res = append(res, buildError(val.Meta.Pos, val.Number, core.ErrEnumFirstValueZero))
		}

	}

	if len(res) == 0 {
		return nil
	}
	return res
}
