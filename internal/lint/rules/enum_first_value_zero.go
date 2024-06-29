package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumFirstValueZero)(nil)

// EnumFirstValueZero this rule enforces that the first enum value is the zero value,
// which is a proto3 requirement on build,
// but isn't required in proto2 on build. The rule enforces that the requirement is also followed in proto2.
type EnumFirstValueZero struct{}

// Validate implements lint.Rule.
func (c *EnumFirstValueZero) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if val := enum.EnumBody.EnumFields[0]; val.Number != "0" {
			res = AppendError(res, ENUM_FIRST_VALUE_ZERO, val.Meta.Pos, val.Number, val.Comments)
		}

	}

	if len(res) == 0 {
		return nil
	}
	return res
}
